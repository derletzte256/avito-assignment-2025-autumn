package user

import (
	"avito-assignment-2025-autumn/internal/entity"
	"avito-assignment-2025-autumn/pkg/httputil"
	"context"
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type UseCase interface {
	SetIsActive(ctx context.Context, req *entity.SetUserActiveRequest) (*entity.User, error)
	GetReviewList(ctx context.Context, userID string) (*entity.UserReviewListResponse, error)
	GetStatistics(ctx context.Context) (*entity.StatsByUsersResponse, error)
	MassDeactivateUsers(ctx context.Context, req *entity.MassDeactivateUsersRequest) (*entity.MassDeactivateUsersResponse, error)
}

type Delivery struct {
	uc     UseCase
	logger *zap.Logger
}

func NewUserDelivery(uc UseCase, logger *zap.Logger) *Delivery {
	return &Delivery{uc: uc, logger: logger}
}

func (d *Delivery) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/setIsActive", d.SetIsActive).Methods("POST")
	r.HandleFunc("/getReview", d.GetReviewList).Methods("GET")
	r.HandleFunc("/getStatistics", d.GetStatistics).Methods("GET")
	r.HandleFunc("/massDeactivate", d.Deactivate).Methods("POST")
}

func (d *Delivery) SetIsActive(w http.ResponseWriter, r *http.Request) {
	var in entity.SetUserActiveRequest
	ctx := r.Context()

	if err := httputil.ReadJSON(r, &in); err != nil {
		d.logger.Warn("failed to read request", zap.Error(err))
		if httputil.WriteAPIError(w, http.StatusBadRequest, entity.ErrorCodeInvalidInput, "invalid JSON body") != nil {
			d.logger.Warn("failed to write error", zap.Error(err))
			return
		}
		return
	}

	validate := validator.New()
	if err := validate.Struct(in); err != nil {
		var errValid validator.ValidationErrors
		ok := errors.As(err, &errValid)
		if !ok {
			d.logger.Error("validation error is not of type ValidationErrors", zap.Error(err))
			httputil.WriteInternalServerError(w, err)
			return
		}
		if writeErr := httputil.WriteAPIError(w, http.StatusBadRequest, entity.ErrorCodeInvalidInput, "validation error: "+errValid.Error()); writeErr != nil {
			d.logger.Warn("failed to write error", zap.Error(writeErr))
			return
		}
		return
	}

	updatedUser, err := d.uc.SetIsActive(ctx, &in)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			d.logger.Warn("failed to find user", zap.Error(err))
			if writeErr := httputil.WriteAPIError(w, http.StatusNotFound, entity.ErrorCodeNotFound, "resource not found"); writeErr != nil {
				d.logger.Warn("failed to write error", zap.Error(writeErr))
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
		d.logger.Warn("failed to write response", zap.Error(err))
		httputil.WriteInternalServerError(w, err)
		return
	}
}

func (d *Delivery) GetReviewList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		if err := httputil.WriteAPIError(w, http.StatusBadRequest, entity.ErrorCodeInvalidInput, "missing user_id parameter"); err != nil {
			d.logger.Warn("failed to write error", zap.Error(err))
			return
		}
		return
	}

	reviewList, err := d.uc.GetReviewList(ctx, userID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			if writeErr := httputil.WriteAPIError(w, http.StatusNotFound, entity.ErrorCodeNotFound, "user_id not found"); writeErr != nil {
				d.logger.Warn("failed to write error", zap.Error(writeErr))
				return
			}
			return
		}
		httputil.WriteInternalServerError(w, err)
		return
	}

	if err := httputil.WriteJSON(w, http.StatusOK, reviewList); err != nil {
		d.logger.Warn("failed to write response", zap.Error(err))
		httputil.WriteInternalServerError(w, err)
		return
	}

}

func (d *Delivery) GetStatistics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stats, err := d.uc.GetStatistics(ctx)
	if err != nil {
		httputil.WriteInternalServerError(w, err)
		return
	}

	if err := httputil.WriteJSON(w, http.StatusOK, stats); err != nil {
		d.logger.Warn("failed to write response", zap.Error(err))
		httputil.WriteInternalServerError(w, err)
		return
	}
}

func (d *Delivery) Deactivate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var in *entity.MassDeactivateUsersRequest
	if err := httputil.ReadJSON(r, &in); err != nil {
		d.logger.Warn("failed to read request", zap.Error(err))
		if writeErr := httputil.WriteAPIError(w, http.StatusBadRequest, entity.ErrorCodeInvalidInput, "invalid JSON body"); writeErr != nil {
			d.logger.Warn("failed to write error", zap.Error(writeErr))
			return
		}
		return
	}

	validate := validator.New()
	if err := validate.Struct(in); err != nil {
		var errValid validator.ValidationErrors
		ok := errors.As(err, &errValid)
		if !ok {
			d.logger.Error("validation error is not of type ValidationErrors", zap.Error(err))
			httputil.WriteInternalServerError(w, err)
			return
		}
		if writeErr := httputil.WriteAPIError(w, http.StatusBadRequest, entity.ErrorCodeInvalidInput, "validation error: "+errValid.Error()); writeErr != nil {
			d.logger.Warn("failed to write error", zap.Error(writeErr))
			return
		}
		return
	}

	resp, err := d.uc.MassDeactivateUsers(ctx, in)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrDuplicateUserIDs):
			if writeErr := httputil.WriteAPIError(w, http.StatusBadRequest, entity.ErrorCodeInvalidInput, "duplicate user IDs in request"); writeErr != nil {
				d.logger.Warn("failed to write error", zap.Error(writeErr))
				return
			}
		case errors.Is(err, entity.ErrNotFound):
			if writeErr := httputil.WriteAPIError(w, http.StatusNotFound, entity.ErrorCodeNotFound, "resource not found"); writeErr != nil {
				d.logger.Warn("failed to write error", zap.Error(writeErr))
				return
			}
		case errors.Is(err, entity.ErrUsersNotInSameTeam):
			if writeErr := httputil.WriteAPIError(w, http.StatusBadRequest, entity.ErrorCodeInvalidInput, "deactivated users should be from the same team"); writeErr != nil {
				d.logger.Warn("failed to write error", zap.Error(writeErr))
				return
			}
		default:
			httputil.WriteInternalServerError(w, err)
		}
		return
	}

	if err := httputil.WriteJSON(w, http.StatusOK, resp); err != nil {
		d.logger.Warn("failed to write response", zap.Error(err))
		httputil.WriteInternalServerError(w, err)
		return
	}
}
