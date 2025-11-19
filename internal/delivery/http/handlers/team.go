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

type TeamUseCase interface {
	CreateTeam(ctx context.Context, team *entity.Team) error
	GetByName(ctx context.Context, name string) (*entity.Team, error)
}

type TeamDelivery struct {
	uc TeamUseCase
}

func NewTeamDelivery(uc TeamUseCase) *TeamDelivery {
	return &TeamDelivery{uc: uc}
}

func (t *TeamDelivery) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/team/add", t.CreateTeam).Methods(http.MethodPost)
	r.HandleFunc("/team/get", t.GetTeam).Methods(http.MethodGet)
}

func (t *TeamDelivery) CreateTeam(w http.ResponseWriter, r *http.Request) {
	var in dto.CreateTeamRequest
	ctx := r.Context()

	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	team := dtoTeamToEntity(in)

	if err := t.uc.CreateTeam(ctx, team); err != nil {
		if errors.Is(err, entity.ErrAlreadyExists) {
			if writeErr := writeAPIError(w, http.StatusBadRequest, dto.ErrorCodeTeamExists, "team already exists"); writeErr != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
			return
		}

		writeInternalServerError(w, err)
		return
	}

	resp := dto.CreateTeamResponse{Team: entityTeamToDTO(team)}
	if err := writeJSON(w, http.StatusCreated, resp); err != nil {
		writeInternalServerError(w, err)
		return
	}

}

func (t *TeamDelivery) GetTeam(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		http.Error(w, "team_name is required", http.StatusBadRequest)
		return
	}

	team, err := t.uc.GetByName(ctx, teamName)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			if writeErr := writeAPIError(w, http.StatusNotFound, dto.ErrorCodeNotFound, "team not found"); writeErr != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
			return
		}

		writeInternalServerError(w, err)
		return
	}

	if err := writeJSON(w, http.StatusOK, entityTeamToDTO(team)); err != nil {
		writeInternalServerError(w, err)
		return
	}
}
