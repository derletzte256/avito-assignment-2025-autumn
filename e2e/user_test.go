package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

type User struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

type SetUserActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

type SetUserActiveResponse struct {
	User User `json:"user"`
}

type PullRequestShort struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
	Status          string `json:"status"`
}

type UserReviewListResponse struct {
	UserID       string             `json:"user_id"`
	PullRequests []PullRequestShort `json:"pull_requests"`
}

type UserStatistics struct {
	UserID          string `json:"user_id"`
	AuthoredPRCount int    `json:"authored_pr_count"`
	OnReviewPRCount int    `json:"on_review_pr_count"`
	ReviewedPRCount int    `json:"reviewed_pr_count"`
}

type StatsByUsersResponse struct {
	UsersStat map[string]UserStatistics `json:"users_stat"`
}

type MassDeactivateUsersRequest struct {
	TeamName string   `json:"team_name"`
	UserIDs  []string `json:"user_ids"`
}

type ReviewReassignment struct {
	PullRequestID string `json:"pull_request_id"`
	OldReviewerID string `json:"old_reviewer_id"`
	NewReviewerID string `json:"new_reviewer_id"`
}

type UnreassignedReview struct {
	PullRequestID string `json:"pull_request_id"`
	ReviewerID    string `json:"reviewer_id"`
}

type MassDeactivateUsersResponse struct {
	TeamName            string               `json:"team_name"`
	ReviewReassignments []ReviewReassignment `json:"review_reassignments"`
	UnreassignedReviews []UnreassignedReview `json:"unreassigned_reviews"`
}

func TestSetUserIsActive(t *testing.T) {
	req := require.New(t)
	e := newExpect(t)

	cleanup, err := setupPostgres()
	req.NoError(err)
	if cleanup != nil {
		t.Cleanup(cleanup)
	}

	team := Team{
		TeamName: "team-1",
		Members: []TeamMember{
			{
				UserID:   "u1",
				Username: "Alice",
				IsActive: true,
			},
		},
	}

	var createTeamResp CreateTeamResponse
	_ = e.POST("/team/add").
		WithHeader("Content-Type", "application/json").
		WithJSON(team).
		Expect().
		Status(http.StatusCreated).
		JSON().Object().Decode(&createTeamResp)

	setReq := SetUserActiveRequest{
		UserID:   "u1",
		IsActive: false,
	}

	var setResp SetUserActiveResponse
	_ = e.POST("/users/setIsActive").
		WithHeader("Content-Type", "application/json").
		WithJSON(setReq).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Decode(&setResp)

	req.Equal(setReq.UserID, setResp.User.UserID)
	req.Equal("Alice", setResp.User.Username)
	req.Equal(team.TeamName, setResp.User.TeamName)
	req.False(setResp.User.IsActive)
}

func TestGetUserReviewList(t *testing.T) {
	req := require.New(t)
	e := newExpect(t)

	cleanup, err := setupPostgres()
	req.NoError(err)
	if cleanup != nil {
		t.Cleanup(cleanup)
	}

	team := Team{
		TeamName: "team-1",
		Members: []TeamMember{
			{
				UserID:   "u-author",
				Username: "Author",
				IsActive: true,
			},
			{
				UserID:   "u-review-1",
				Username: "Reviewer1",
				IsActive: true,
			},
			{
				UserID:   "u-review-2",
				Username: "Reviewer2",
				IsActive: true,
			},
		},
	}

	_ = e.POST("/team/add").
		WithHeader("Content-Type", "application/json").
		WithJSON(team).
		Expect().
		Status(http.StatusCreated)

	createPRReq := CreatePullRequestRequest{
		PullRequestID:   "pr1",
		PullRequestName: "Add feature X",
		AuthorID:        "u-author",
	}

	var prResp PullRequest
	_ = e.POST("/pullRequest/create").
		WithHeader("Content-Type", "application/json").
		WithJSON(createPRReq).
		Expect().
		Status(http.StatusCreated).
		JSON().Object().Decode(&prResp)

	req.NotEmpty(prResp.AssignedReviewers)

	reviewerID := prResp.AssignedReviewers[0]

	var reviewListResp UserReviewListResponse
	_ = e.GET("/users/getReview").
		WithQuery("user_id", reviewerID).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Decode(&reviewListResp)

	req.Equal(reviewerID, reviewListResp.UserID)
	req.NotEmpty(reviewListResp.PullRequests)

	found := false
	for _, pr := range reviewListResp.PullRequests {
		if pr.PullRequestID == createPRReq.PullRequestID {
			req.Equal(createPRReq.AuthorID, pr.AuthorID)
			req.Equal("OPEN", pr.Status)
			found = true
			break
		}
	}
	req.True(found, "created pull request should be in user's review list")
}

func TestGetUserReviewList_UserNotFound(t *testing.T) {
	req := require.New(t)
	e := newExpect(t)

	cleanup, err := setupPostgres()
	req.NoError(err)
	if cleanup != nil {
		t.Cleanup(cleanup)
	}

	var errResp ErrorResponse
	_ = e.GET("/users/getReview").
		WithQuery("user_id", "user-review-not-found").
		Expect().
		Status(http.StatusNotFound).
		JSON().Object().Decode(&errResp)

	req.Equal("NOT_FOUND", errResp.Error.Code)
}

func TestGetUserStatistics(t *testing.T) {
	req := require.New(t)
	e := newExpect(t)

	cleanup, err := setupPostgres()
	req.NoError(err)
	if cleanup != nil {
		t.Cleanup(cleanup)
	}

	team := Team{
		TeamName: "team-1",
		Members: []TeamMember{
			{
				UserID:   "u1",
				Username: "Stats1",
				IsActive: true,
			},
			{
				UserID:   "u2",
				Username: "Stats2",
				IsActive: true,
			},
		},
	}

	_ = e.POST("/team/add").
		WithHeader("Content-Type", "application/json").
		WithJSON(team).
		Expect().
		Status(http.StatusCreated)

	var statsResp StatsByUsersResponse
	_ = e.GET("/users/getStatistics").
		Expect().
		Status(http.StatusOK).
		JSON().Object().Decode(&statsResp)

	req.Contains(statsResp.UsersStat, "u1")
	req.Contains(statsResp.UsersStat, "u2")

	user1Stats := statsResp.UsersStat["u1"]
	user2Stats := statsResp.UsersStat["u2"]

	req.Equal("u1", user1Stats.UserID)
	req.Equal("u2", user2Stats.UserID)

	req.Equal(0, user1Stats.AuthoredPRCount)
	req.Equal(0, user1Stats.OnReviewPRCount)
	req.Equal(0, user1Stats.ReviewedPRCount)

	req.Equal(0, user2Stats.AuthoredPRCount)
	req.Equal(0, user2Stats.OnReviewPRCount)
	req.Equal(0, user2Stats.ReviewedPRCount)
}

func TestMassDeactivateUsers(t *testing.T) {
	req := require.New(t)
	e := newExpect(t)

	cleanup, err := setupPostgres()
	req.NoError(err)
	if cleanup != nil {
		t.Cleanup(cleanup)
	}

	team := Team{
		TeamName: "team-1",
		Members: []TeamMember{
			{
				UserID:   "u1",
				Username: "ToDeactivate",
				IsActive: true,
			},
		},
	}

	_ = e.POST("/team/add").
		WithHeader("Content-Type", "application/json").
		WithJSON(team).
		Expect().
		Status(http.StatusCreated)

	deactivateReq := MassDeactivateUsersRequest{
		TeamName: team.TeamName,
		UserIDs:  []string{"u1"},
	}

	var deactivateResp MassDeactivateUsersResponse
	_ = e.POST("/users/massDeactivate").
		WithHeader("Content-Type", "application/json").
		WithJSON(deactivateReq).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Decode(&deactivateResp)

	req.Equal(team.TeamName, deactivateResp.TeamName)
	req.Empty(deactivateResp.ReviewReassignments)
	req.Empty(deactivateResp.UnreassignedReviews)
}

func TestMassDeactivateUsers_WithReassignment(t *testing.T) {
	req := require.New(t)
	e := newExpect(t)

	cleanup, err := setupPostgres()
	req.NoError(err)
	if cleanup != nil {
		t.Cleanup(cleanup)
	}

	team := Team{
		TeamName: "team-1",
		Members: []TeamMember{
			{
				UserID:   "u-author",
				Username: "Author",
				IsActive: true,
			},
			{
				UserID:   "u-deactivate",
				Username: "ToDeactivate",
				IsActive: true,
			},
			{
				UserID:   "u-active",
				Username: "ActiveReviewer",
				IsActive: true,
			},
		},
	}

	_ = e.POST("/team/add").
		WithHeader("Content-Type", "application/json").
		WithJSON(team).
		Expect().
		Status(http.StatusCreated)

	createPRReq := CreatePullRequestRequest{
		PullRequestID:   "pr1",
		PullRequestName: "PR for mass deactivate",
		AuthorID:        "u-author",
	}

	var createdPR PullRequest
	_ = e.POST("/pullRequest/create").
		WithHeader("Content-Type", "application/json").
		WithJSON(createPRReq).
		Expect().
		Status(http.StatusCreated).
		JSON().Object().Decode(&createdPR)

	req.NotEmpty(createdPR.AssignedReviewers)

	foundDeactivated := false
	for _, rID := range createdPR.AssignedReviewers {
		if rID == "u-deactivate" {
			foundDeactivated = true
			break
		}
	}
	req.True(foundDeactivated, "deactivated user should be assigned as reviewer")

	deactivateReq := MassDeactivateUsersRequest{
		TeamName: team.TeamName,
		UserIDs:  []string{"u-deactivate"},
	}

	var deactivateResp MassDeactivateUsersResponse
	_ = e.POST("/users/massDeactivate").
		WithHeader("Content-Type", "application/json").
		WithJSON(deactivateReq).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Decode(&deactivateResp)

	req.Equal(team.TeamName, deactivateResp.TeamName)
	req.Len(deactivateResp.ReviewReassignments, 1)
	req.Empty(deactivateResp.UnreassignedReviews)

	reassignment := deactivateResp.ReviewReassignments[0]
	req.Equal(createPRReq.PullRequestID, reassignment.PullRequestID)
	req.Equal("u-deactivate", reassignment.OldReviewerID)
	req.NotEmpty(reassignment.NewReviewerID)
	req.NotEqual("u-deactivate", reassignment.NewReviewerID)

	allowedNew := map[string]struct{}{
		"u-author": {},
		"u-active": {},
	}
	_, ok := allowedNew[reassignment.NewReviewerID]
	req.True(ok, "new reviewer should be an active team member")

	var oldReviewList UserReviewListResponse
	_ = e.GET("/users/getReview").
		WithQuery("user_id", "u-deactivate").
		Expect().
		Status(http.StatusOK).
		JSON().Object().Decode(&oldReviewList)

	for _, pr := range oldReviewList.PullRequests {
		req.NotEqual(createPRReq.PullRequestID, pr.PullRequestID, "deactivated reviewer should not have the PR in review list")
	}

	var newReviewList UserReviewListResponse
	_ = e.GET("/users/getReview").
		WithQuery("user_id", reassignment.NewReviewerID).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Decode(&newReviewList)

	foundPR := false
	for _, pr := range newReviewList.PullRequests {
		if pr.PullRequestID == createPRReq.PullRequestID {
			foundPR = true
			break
		}
	}
	req.True(foundPR, "new reviewer should have the PR in review list")
}
