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

type PullRequestUseCase interface {
	CreatePullRequest(ctx context.Context, pr *entity.CreatePullRequestRequest) (*entity.PullRequest, error)
	MergePullRequest(ctx context.Context, pr *entity.MergePullRequestRequest) (*entity.PullRequest, error)
}

type PullRequestDelivery struct {
	uc     PullRequestUseCase
	logger *zap.Logger
}

func NewPullRequestDelivery(uc PullRequestUseCase, logger *zap.Logger) *PullRequestDelivery {
	return &PullRequestDelivery{uc: uc, logger: logger}
}

func (prd *PullRequestDelivery) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/create", prd.CreatePullRequest).Methods(http.MethodPost)
	r.HandleFunc("/merge", prd.MergePullRequest).Methods(http.MethodPost)
}

func (prd *PullRequestDelivery) CreatePullRequest(w http.ResponseWriter, r *http.Request) {
	var in *entity.CreatePullRequestRequest
	ctx := context.Background()

	if err := httputil.ReadJSON(r, &in); err != nil {
		prd.logger.Warn("failed to read request", zap.Error(err))
		if writeErr := httputil.WriteAPIError(w, http.StatusBadRequest, entity.ErrorCodeInvalidInput, "invalid JSON body"); writeErr != nil {
			prd.logger.Warn("failed to write error", zap.Error(writeErr))
			return
		}
		return
	}

	pr, err := prd.uc.CreatePullRequest(ctx, in)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrNotFound):
			if writeErr := httputil.WriteAPIError(w, http.StatusNotFound, entity.ErrorCodeNotFound, "resource not found"); writeErr != nil {
				prd.logger.Warn("failed to write error", zap.Error(writeErr))
				return
			}
		case errors.Is(err, entity.ErrAlreadyExists):
			if writeErr := httputil.WriteAPIError(w, http.StatusBadRequest, entity.ErrorCodePRExists, "PR id already exists"); writeErr != nil {
				prd.logger.Warn("failed to write error", zap.Error(writeErr))
				return
			}
		default:
			httputil.WriteInternalServerError(w, err)
			return
		}
		return
	}

	if err := httputil.WriteJSON(w, http.StatusCreated, pr); err != nil {
		prd.logger.Warn("failed to write response", zap.Error(err))
		httputil.WriteInternalServerError(w, err)
		return
	}
}

func (prd *PullRequestDelivery) MergePullRequest(w http.ResponseWriter, r *http.Request) {
	var in *entity.MergePullRequestRequest
	ctx := context.Background()

	if err := httputil.ReadJSON(r, &in); err != nil {
		prd.logger.Warn("failed to read request", zap.Error(err))
		if writeErr := httputil.WriteAPIError(w, http.StatusBadRequest, entity.ErrorCodeInvalidInput, "invalid JSON body"); writeErr != nil {
			prd.logger.Warn("failed to write error", zap.Error(writeErr))
			return
		}
		return
	}

	pr, err := prd.uc.MergePullRequest(ctx, in)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrNotFound):
			if writeErr := httputil.WriteAPIError(w, http.StatusNotFound, entity.ErrorCodeNotFound, "resource not found"); writeErr != nil {
				prd.logger.Warn("failed to write error", zap.Error(writeErr))
				return
			}
		default:
			httputil.WriteInternalServerError(w, err)
			return
		}
		return
	}

	if err := httputil.WriteJSON(w, http.StatusOK, pr); err != nil {
		prd.logger.Warn("failed to write response", zap.Error(err))
		httputil.WriteInternalServerError(w, err)
		return
	}
}
