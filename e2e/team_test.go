package e2e

import (
	"net/http"
	"testing"

	"sort"

	"github.com/stretchr/testify/require"
)

type TeamMember struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type Team struct {
	TeamName string       `json:"team_name"`
	Members  []TeamMember `json:"members"`
}

type CreateTeamResponse struct {
	Team Team `json:"team"`
}

func NormalizeTeam(t Team) Team {
	c := Team{
		TeamName: t.TeamName,
		Members:  append([]TeamMember(nil), t.Members...),
	}

	sort.Slice(c.Members, func(i, j int) bool {
		return c.Members[i].UserID < c.Members[j].UserID
	})

	return c
}

func TestCreateAndGetTeam(t *testing.T) {
	req := require.New(t)
	e := newExpect(t)

	cleanup, err := setupPostgres()
	req.NoError(err)
	if cleanup != nil {
		t.Cleanup(cleanup)
	}

	expectedTeam := Team{
		TeamName: "team-1",
		Members: []TeamMember{
			{
				UserID:   "u1",
				Username: "Alice",
				IsActive: true,
			},
			{
				UserID:   "u2",
				Username: "Bob",
				IsActive: true,
			},
		},
	}

	var createResp CreateTeamResponse
	_ = e.POST("/team/add").
		WithHeader("Content-Type", "application/json").
		WithJSON(expectedTeam).
		Expect().
		Status(http.StatusCreated).
		JSON().Object().Decode(&createResp)

	req.Equal(
		NormalizeTeam(expectedTeam),
		NormalizeTeam(createResp.Team),
		"created team should match expected payload",
	)

	var gotTeam Team
	_ = e.GET("/team/get").
		WithQuery("team_name", expectedTeam.TeamName).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Decode(&gotTeam)

	req.Equal(
		NormalizeTeam(expectedTeam),
		NormalizeTeam(gotTeam),
		"fetched team should match expected payload",
	)
}

func TestCreateDuplicateTeam(t *testing.T) {
	req := require.New(t)
	e := newExpect(t)

	cleanup, err := setupPostgres()
	req.NoError(err)
	if cleanup != nil {
		t.Cleanup(cleanup)
	}

	teamPayload := Team{
		TeamName: "team-1",
		Members: []TeamMember{
			{
				UserID:   "u1",
				Username: "Alice",
				IsActive: true,
			},
		},
	}

	_ = e.POST("/team/add").
		WithHeader("Content-Type", "application/json").
		WithJSON(teamPayload).
		Expect().
		Status(http.StatusCreated)

	_ = e.POST("/team/add").
		WithHeader("Content-Type", "application/json").
		WithJSON(teamPayload).
		Expect().
		Status(http.StatusConflict)
}

func TestGetNonExistentTeam(t *testing.T) {
	req := require.New(t)
	e := newExpect(t)

	cleanup, err := setupPostgres()
	req.NoError(err)
	if cleanup != nil {
		t.Cleanup(cleanup)
	}

	_ = e.GET("/team/get").
		WithQuery("team_name", "non-existent-team").
		Expect().
		Status(http.StatusNotFound)
}
