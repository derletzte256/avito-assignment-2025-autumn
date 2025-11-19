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

type PullRequestUseCase interface {
	Create(ctx context.Context, prID, name, authorID string) (*entity.PullRequest, error)
	Merge(ctx context.Context, prID string) (*entity.PullRequest, error)
	Reassign(ctx context.Context, prID, oldReviewerID string) (*entity.PullRequest, string, error)
}

type PullRequestDelivery struct {
	uc PullRequestUseCase
}

func NewPullRequestDelivery(uc PullRequestUseCase) *PullRequestDelivery {
	return &PullRequestDelivery{uc: uc}
}

func (d *PullRequestDelivery) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/pullRequest/create", d.CreatePullRequest).Methods(http.MethodPost)
	r.HandleFunc("/pullRequest/merge", d.MergePullRequest).Methods(http.MethodPost)
	r.HandleFunc("/pullRequest/reassign", d.ReassignPullRequest).Methods(http.MethodPost)
}

func (d *PullRequestDelivery) CreatePullRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var in dto.CreatePullRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if in.PullRequestID == "" || in.PullRequestName == "" || in.AuthorID == "" {
		http.Error(w, "pull_request_id, pull_request_name and author_id are required", http.StatusBadRequest)
		return
	}

	pr, err := d.uc.Create(ctx, in.PullRequestID, in.PullRequestName, in.AuthorID)
	if err != nil {
		if handled := writePullRequestDomainError(w, err); handled {
			return
		}

		writeInternalServerError(w, err)
		return
	}

	resp := dto.CreatePullRequestResponse{PR: entityPullRequestToDTO(pr)}
	if err := writeJSON(w, http.StatusCreated, resp); err != nil {
		writeInternalServerError(w, err)
		return
	}
}

func (d *PullRequestDelivery) MergePullRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var in dto.MergePullRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if in.PullRequestID == "" {
		http.Error(w, "pull_request_id is required", http.StatusBadRequest)
		return
	}

	pr, err := d.uc.Merge(ctx, in.PullRequestID)
	if err != nil {
		if handled := writePullRequestDomainError(w, err); handled {
			return
		}

		writeInternalServerError(w, err)
		return
	}

	resp := dto.MergePullRequestResponse{PR: entityPullRequestToDTO(pr)}
	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		writeInternalServerError(w, err)
		return
	}
}

func (d *PullRequestDelivery) ReassignPullRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var in dto.ReassignPullRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if in.PullRequestID == "" || in.OldUserID == "" {
		http.Error(w, "pull_request_id and old_user_id are required", http.StatusBadRequest)
		return
	}

	pr, replacedBy, err := d.uc.Reassign(ctx, in.PullRequestID, in.OldUserID)
	if err != nil {
		if handled := writePullRequestDomainError(w, err); handled {
			return
		}

		writeInternalServerError(w, err)
		return
	}

	resp := dto.ReassignPullRequestResponse{
		PullRequestResponse: dto.PullRequestResponse{PR: entityPullRequestToDTO(pr)},
		ReplacedBy:          replacedBy,
	}

	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		writeInternalServerError(w, err)
		return
	}
}

func writePullRequestDomainError(w http.ResponseWriter, err error) bool {
	var (
		status int
		code   dto.ErrorCode
		msg    string
	)

	switch {
	case errors.Is(err, entity.ErrAlreadyExists):
		status = http.StatusConflict
		code = dto.ErrorCodePRExists
		msg = "pull request already exists"
	case errors.Is(err, entity.ErrNotFound):
		status = http.StatusNotFound
		code = dto.ErrorCodeNotFound
		msg = "resource not found"
	case errors.Is(err, entity.ErrPRMerged):
		status = http.StatusConflict
		code = dto.ErrorCodePRMerged
		msg = "pull request already merged"
	case errors.Is(err, entity.ErrNotAssigned):
		status = http.StatusConflict
		code = dto.ErrorCodeNotAssigned
		msg = "reviewer not assigned"
	case errors.Is(err, entity.ErrNoCandidate):
		status = http.StatusConflict
		code = dto.ErrorCodeNoCandidate
		msg = "no replacement candidate"
	default:
		return false
	}

	if writeErr := writeAPIError(w, status, code, msg); writeErr != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}

	return true
}
