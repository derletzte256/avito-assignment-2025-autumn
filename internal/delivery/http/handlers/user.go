package handlers

import (
	"avito-assignment-2025-autumn/internal/delivery/http/dto"
	"avito-assignment-2025-autumn/internal/entity"
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
)

type UserUseCase interface {
	SetIsActive(ctx context.Context, userID string, isActive bool) (*entity.User, error)
	GetReviewList(ctx context.Context, userID string) ([]*entity.PullRequest, error)
}

type UserDelivery struct {
	uc UserUseCase
}

func NewUserDelivery(uc UserUseCase) *UserDelivery {
	return &UserDelivery{uc: uc}
}

func (d *UserDelivery) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/users/setIsActive", d.SetIsActive).Methods(http.MethodPost)
	r.HandleFunc("/users/getReview", d.GetReviewList).Methods(http.MethodGet)
}

func (d *UserDelivery) SetIsActive(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var in dto.SetUserActiveRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if in.UserID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	user, err := d.uc.SetIsActive(ctx, in.UserID, in.IsActive)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			if writeErr := writeAPIError(w, http.StatusNotFound, dto.ErrorCodeNotFound, "user not found"); writeErr != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
			return
		}

		writeInternalServerError(w, err)
		return
	}

	resp := dto.SetUserActiveResponse{User: entityUserToDTO(user)}
	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		writeInternalServerError(w, err)
		return
	}
}

func (d *UserDelivery) GetReviewList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	pullRequests, err := d.uc.GetReviewList(ctx, userID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			if writeErr := writeAPIError(w, http.StatusNotFound, dto.ErrorCodeNotFound, "user not found"); writeErr != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
			return
		}

		writeInternalServerError(w, err)
		return
	}

	resp := dto.UserReviewListResponse{
		UserID:       userID,
		PullRequests: entityPullRequestsToShortDTOs(pullRequests),
	}

	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		writeInternalServerError(w, err)
		return
	}
}
