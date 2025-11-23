package pullRequest

import (
	"context"
	"errors"
	"net/http"

	"github.com/derletzte256/avito-assignment-2025-autumn/internal/entity"
	"github.com/derletzte256/avito-assignment-2025-autumn/internal/pkg/httputil"
	"github.com/derletzte256/avito-assignment-2025-autumn/internal/pkg/logger"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type UseCase interface {
	CreatePullRequest(ctx context.Context, pr *entity.CreatePullRequestRequest) (*entity.PullRequest, error)
	MergePullRequest(ctx context.Context, pr *entity.MergePullRequestRequest) (*entity.PullRequest, error)
	ReassignPullRequest(ctx context.Context, pr *entity.ReassignPullRequestRequest) (*entity.ReassignPullRequestResponse, error)
}

type Delivery struct {
	uc UseCase
}

func NewDelivery(uc UseCase) *Delivery {
	return &Delivery{uc: uc}
}

func (d *Delivery) RegisterRoutes(r *mux.Router) {
	s := r.PathPrefix("/pullRequest").Subrouter()
	s.HandleFunc("/create", d.CreatePullRequest).Methods(http.MethodPost)
	s.HandleFunc("/merge", d.MergePullRequest).Methods(http.MethodPost)
	s.HandleFunc("/reassign", d.ReassignPullRequest).Methods(http.MethodPost)
}

func (d *Delivery) CreatePullRequest(w http.ResponseWriter, r *http.Request) {
	var in *entity.CreatePullRequestRequest
	ctx := context.Background()
	l := logger.FromCtx(ctx)

	if err := httputil.ReadJSON(r, &in); err != nil {
		l.Warn("failed to read request", zap.Error(err))
		if writeErr := httputil.WriteAPIError(w, http.StatusBadRequest, entity.ErrorCodeInvalidInput, "invalid JSON body"); writeErr != nil {
			l.Error("failed to write error", zap.Error(writeErr))
			return
		}
		return
	}

	validate := validator.New()
	if err := validate.Struct(in); err != nil {
		var errValid validator.ValidationErrors
		ok := errors.As(err, &errValid)
		if !ok {
			l.Error("validation error is not of type ValidationErrors", zap.Error(err))
			httputil.WriteInternalServerError(w, err)
			return
		}
		if writeErr := httputil.WriteAPIError(w, http.StatusBadRequest, entity.ErrorCodeInvalidInput, "validation error: "+errValid.Error()); writeErr != nil {
			l.Error("failed to write error", zap.Error(writeErr))
			return
		}
		return
	}

	pr, err := d.uc.CreatePullRequest(ctx, in)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrNotFound):
			if writeErr := httputil.WriteAPIError(w, http.StatusNotFound, entity.ErrorCodeNotFound, "resource not found"); writeErr != nil {
				l.Error("failed to write error", zap.Error(writeErr))
				return
			}
		case errors.Is(err, entity.ErrAlreadyExists):
			if writeErr := httputil.WriteAPIError(w, http.StatusBadRequest, entity.ErrorCodePRExists, "PR id already exists"); writeErr != nil {
				l.Error("failed to write error", zap.Error(writeErr))
				return
			}
		default:
			httputil.WriteInternalServerError(w, err)
			return
		}
		return
	}

	if err = httputil.WriteJSON(w, http.StatusCreated, pr); err != nil {
		l.Warn("failed to write response", zap.Error(err))
		httputil.WriteInternalServerError(w, err)
		return
	}
}

func (d *Delivery) MergePullRequest(w http.ResponseWriter, r *http.Request) {
	var in *entity.MergePullRequestRequest
	ctx := context.Background()
	l := logger.FromCtx(ctx)

	if err := httputil.ReadJSON(r, &in); err != nil {
		l.Warn("failed to read request", zap.Error(err))
		if writeErr := httputil.WriteAPIError(w, http.StatusBadRequest, entity.ErrorCodeInvalidInput, "invalid JSON body"); writeErr != nil {
			l.Error("failed to write error", zap.Error(writeErr))
			return
		}
		return
	}

	validate := validator.New()
	if err := validate.Struct(in); err != nil {
		var errValid validator.ValidationErrors
		ok := errors.As(err, &errValid)
		if !ok {
			l.Error("validation error is not of type ValidationErrors", zap.Error(err))
			httputil.WriteInternalServerError(w, err)
			return
		}
		if writeErr := httputil.WriteAPIError(w, http.StatusBadRequest, entity.ErrorCodeInvalidInput, "validation error: "+errValid.Error()); writeErr != nil {
			l.Error("failed to write error", zap.Error(writeErr))
			return
		}
		return
	}

	pr, err := d.uc.MergePullRequest(ctx, in)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrNotFound):
			if writeErr := httputil.WriteAPIError(w, http.StatusNotFound, entity.ErrorCodeNotFound, "resource not found"); writeErr != nil {
				l.Error("failed to write error", zap.Error(writeErr))
				return
			}
		default:
			httputil.WriteInternalServerError(w, err)
			return
		}
		return
	}

	if err = httputil.WriteJSON(w, http.StatusOK, pr); err != nil {
		l.Error("failed to write response", zap.Error(err))
		httputil.WriteInternalServerError(w, err)
		return
	}
}

func (d *Delivery) ReassignPullRequest(w http.ResponseWriter, r *http.Request) {
	var in *entity.ReassignPullRequestRequest
	ctx := context.Background()
	l := logger.FromCtx(ctx)

	if err := httputil.ReadJSON(r, &in); err != nil {
		l.Warn("failed to read request", zap.Error(err))
		if writeErr := httputil.WriteAPIError(w, http.StatusBadRequest, entity.ErrorCodeInvalidInput, "invalid JSON body"); writeErr != nil {
			l.Error("failed to write error", zap.Error(writeErr))
			return
		}
		return
	}

	validate := validator.New()
	if err := validate.Struct(in); err != nil {
		var errValid validator.ValidationErrors
		ok := errors.As(err, &errValid)
		if !ok {
			l.Error("validation error is not of type ValidationErrors", zap.Error(err))
			httputil.WriteInternalServerError(w, err)
			return
		}
		if writeErr := httputil.WriteAPIError(w, http.StatusBadRequest, entity.ErrorCodeInvalidInput, "validation error: "+errValid.Error()); writeErr != nil {
			l.Error("failed to write error", zap.Error(writeErr))
			return
		}
		return
	}

	pr, err := d.uc.ReassignPullRequest(ctx, in)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrNotFound):
			if writeErr := httputil.WriteAPIError(w, http.StatusNotFound, entity.ErrorCodeNotFound, "resource not found"); writeErr != nil {
				l.Error("failed to write error", zap.Error(writeErr))
				return
			}
		case errors.Is(err, entity.ErrPRMerged):
			if writeErr := httputil.WriteAPIError(w, http.StatusBadRequest, entity.ErrorCodePRMerged, "cannot reassign on merged PR"); writeErr != nil {
				l.Error("failed to write error", zap.Error(writeErr))
				return
			}
		case errors.Is(err, entity.ErrNotAssignedReviewer):
			if writeErr := httputil.WriteAPIError(w, http.StatusBadRequest, entity.ErrorCodeNotAssigned, "reviewer is not assigned to this PR"); writeErr != nil {
				l.Error("failed to write error", zap.Error(writeErr))
				return
			}
		case errors.Is(err, entity.ErrNoCandidate):
			if writeErr := httputil.WriteAPIError(w, http.StatusConflict, entity.ErrorCodeNoCandidate, "no active replacement candidate in team"); writeErr != nil {
				l.Error("failed to write error", zap.Error(writeErr))
				return
			}
		default:
			httputil.WriteInternalServerError(w, err)
			return
		}
		return
	}

	if err = httputil.WriteJSON(w, http.StatusOK, pr); err != nil {
		l.Error("failed to write response", zap.Error(err))
		httputil.WriteInternalServerError(w, err)
		return
	}
}
