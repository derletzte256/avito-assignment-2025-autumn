package handlers

import (
	"avito-assignment-2025-autumn/internal/entity"
	"avito-assignment-2025-autumn/pkg/httputil"
	"context"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type UserUseCase interface {
	SetIsActive(ctx context.Context, req *entity.SetUserActiveRequest) (*entity.User, error)
	GetReviewList(ctx context.Context, userID string) (*entity.UserReviewListResponse, error)
}

type UserDelivery struct {
	uc     UserUseCase
	logger *zap.Logger
}

func NewUserDelivery(uc UserUseCase, logger *zap.Logger) *UserDelivery {
	return &UserDelivery{uc: uc, logger: logger}
}

func (u *UserDelivery) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/setIsActive", u.SetIsActive).Methods("POST")
	r.HandleFunc("/getReview", u.GetReviewList).Methods("GET")
}

func (u *UserDelivery) SetIsActive(w http.ResponseWriter, r *http.Request) {
	var in entity.SetUserActiveRequest
	ctx := r.Context()

	if err := httputil.ReadJSON(r, &in); err != nil {
		u.logger.Warn("failed to read request", zap.Error(err))
		if httputil.WriteAPIError(w, http.StatusBadRequest, entity.ErrorCodeInvalidInput, "invalid JSON body") != nil {
			u.logger.Warn("failed to write error", zap.Error(err))
			return
		}
		return
	}

	updatedUser, err := u.uc.SetIsActive(ctx, &in)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			u.logger.Warn("failed to find user", zap.Error(err))
			if writeErr := httputil.WriteAPIError(w, http.StatusNotFound, entity.ErrorCodeNotFound, "resource not found"); writeErr != nil {
				u.logger.Warn("failed to write error", zap.Error(writeErr))
				return
			}
			return
		}

		httputil.WriteInternalServerError(w, err)
		return
	}

	var resp entity.SetUserActiveResponse
	resp.User = updatedUser

	if err := httputil.WriteJSON(w, http.StatusOK, resp); err != nil {
		u.logger.Warn("failed to write response", zap.Error(err))
		httputil.WriteInternalServerError(w, err)
		return
	}
}

func (u *UserDelivery) GetReviewList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		if err := httputil.WriteAPIError(w, http.StatusBadRequest, entity.ErrorCodeInvalidInput, "missing user_id parameter"); err != nil {
			u.logger.Warn("failed to write error", zap.Error(err))
			return
		}
		return
	}

	reviewList, err := u.uc.GetReviewList(ctx, userID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			if writeErr := httputil.WriteAPIError(w, http.StatusNotFound, entity.ErrorCodeNotFound, "user_id not found"); writeErr != nil {
				u.logger.Warn("failed to write error", zap.Error(writeErr))
				return
			}
			return
		}
		httputil.WriteInternalServerError(w, err)
		return
	}

	if err := httputil.WriteJSON(w, http.StatusOK, reviewList); err != nil {
		u.logger.Warn("failed to write response", zap.Error(err))
		httputil.WriteInternalServerError(w, err)
		return
	}

}
