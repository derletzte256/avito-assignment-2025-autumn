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

type TeamUseCase interface {
	CreateTeam(ctx context.Context, team *entity.Team) error
	GetByName(ctx context.Context, name string) (*entity.Team, error)
}

type TeamDelivery struct {
	uc     TeamUseCase
	logger *zap.Logger
}

func NewTeamDelivery(uc TeamUseCase, logger *zap.Logger) *TeamDelivery {
	return &TeamDelivery{uc: uc, logger: logger}
}

func (t *TeamDelivery) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/add", t.CreateTeam).Methods(http.MethodPost)
	r.HandleFunc("/get", t.GetTeam).Methods(http.MethodGet)
}

func (t *TeamDelivery) CreateTeam(w http.ResponseWriter, r *http.Request) {
	var in entity.CreateTeamRequest
	ctx := r.Context()

	if err := httputil.ReadJSON(r, &in); err != nil {
		t.logger.Warn("failed to read request", zap.Error(err))
		if httputil.WriteAPIError(w, http.StatusBadRequest, entity.ErrorCodeInvalidInput, "invalid JSON body") != nil {
			t.logger.Warn("failed to read request", zap.Error(err))
			return
		}
		return
	}

	team := &in

	if err := t.uc.CreateTeam(ctx, team); err != nil {
		if errors.Is(err, entity.ErrAlreadyExists) {
			if writeErr := httputil.WriteAPIError(w, http.StatusConflict, entity.ErrorCodeTeamExists, "team_name already exists"); writeErr != nil {
				t.logger.Warn("failed to write error", zap.Error(writeErr))
				return
			}
			return
		}

		httputil.WriteInternalServerError(w, err)
		return
	}

	resp := entity.CreateTeamResponse{Team: team}
	if err := httputil.WriteJSON(w, http.StatusCreated, resp); err != nil {
		t.logger.Warn("failed to write response", zap.Error(err))
		httputil.WriteInternalServerError(w, err)
		return
	}

}

func (t *TeamDelivery) GetTeam(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		if writeErr := httputil.WriteAPIError(w, http.StatusBadRequest, entity.ErrorCodeInvalidInput, "team_name is required"); writeErr != nil {
			t.logger.Warn("failed to write error", zap.Error(writeErr))
			return
		}
		return
	}

	team, err := t.uc.GetByName(ctx, teamName)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			if writeErr := httputil.WriteAPIError(w, http.StatusNotFound, entity.ErrorCodeNotFound, "team not found"); writeErr != nil {
				t.logger.Warn("failed to write error", zap.Error(writeErr))
				return
			}
			return
		}

		httputil.WriteInternalServerError(w, err)
		return
	}

	if err := httputil.WriteJSON(w, http.StatusOK, team); err != nil {
		t.logger.Warn("failed to write response", zap.Error(err))
		httputil.WriteInternalServerError(w, err)
		return
	}
}
