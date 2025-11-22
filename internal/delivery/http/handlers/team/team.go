package team

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

type TeamUseCase interface {
	CreateTeam(ctx context.Context, team *entity.Team) error
	GetByName(ctx context.Context, name string) (*entity.Team, error)
}

type Delivery struct {
	uc     TeamUseCase
	logger *zap.Logger
}

func NewTeamDelivery(uc TeamUseCase, logger *zap.Logger) *Delivery {
	return &Delivery{uc: uc, logger: logger}
}

func (d *Delivery) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/add", d.CreateTeam).Methods(http.MethodPost)
	r.HandleFunc("/get", d.GetTeam).Methods(http.MethodGet)
}

func (d *Delivery) CreateTeam(w http.ResponseWriter, r *http.Request) {
	var in entity.CreateTeamRequest
	ctx := r.Context()

	if err := httputil.ReadJSON(r, &in); err != nil {
		d.logger.Warn("failed to read request", zap.Error(err))
		if httputil.WriteAPIError(w, http.StatusBadRequest, entity.ErrorCodeInvalidInput, "invalid JSON body") != nil {
			d.logger.Warn("failed to read request", zap.Error(err))
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

	team := &in

	if err := d.uc.CreateTeam(ctx, team); err != nil {
		switch {
		case errors.Is(err, entity.ErrAlreadyExists):
			if writeErr := httputil.WriteAPIError(w, http.StatusConflict, entity.ErrorCodeTeamExists, "team_name already exists"); writeErr != nil {
				d.logger.Warn("failed to write error", zap.Error(writeErr))
				return
			}
		case errors.Is(err, entity.ErrDuplicateUserIDs):
			if writeErr := httputil.WriteAPIError(w, http.StatusBadRequest, entity.ErrorCodeInvalidInput, "duplicate user IDs in members"); writeErr != nil {
				d.logger.Warn("failed to write error", zap.Error(writeErr))
				return
			}
		case errors.Is(err, entity.ErrNotFound):
			if writeErr := httputil.WriteAPIError(w, http.StatusNotFound, entity.ErrorCodeNotFound, "one or more members not found"); writeErr != nil {
				d.logger.Warn("failed to write error", zap.Error(writeErr))
				return
			}
		default:
			httputil.WriteInternalServerError(w, err)
		}
		return
	}

	resp := entity.CreateTeamResponse{Team: team}
	if err := httputil.WriteJSON(w, http.StatusCreated, resp); err != nil {
		d.logger.Warn("failed to write response", zap.Error(err))
		httputil.WriteInternalServerError(w, err)
		return
	}

}

func (d *Delivery) GetTeam(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		if writeErr := httputil.WriteAPIError(w, http.StatusBadRequest, entity.ErrorCodeInvalidInput, "team_name is required"); writeErr != nil {
			d.logger.Warn("failed to write error", zap.Error(writeErr))
			return
		}
		return
	}

	team, err := d.uc.GetByName(ctx, teamName)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			if writeErr := httputil.WriteAPIError(w, http.StatusNotFound, entity.ErrorCodeNotFound, "team not found"); writeErr != nil {
				d.logger.Warn("failed to write error", zap.Error(writeErr))
				return
			}
			return
		}

		httputil.WriteInternalServerError(w, err)
		return
	}

	if err := httputil.WriteJSON(w, http.StatusOK, team); err != nil {
		d.logger.Warn("failed to write response", zap.Error(err))
		httputil.WriteInternalServerError(w, err)
		return
	}
}
