package e2e

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type PullRequest struct {
	PullRequestID     string     `json:"pull_request_id"`
	PullRequestName   string     `json:"pull_request_name"`
	AuthorID          string     `json:"author_id"`
	Status            string     `json:"status"`
	AssignedReviewers []string   `json:"assigned_reviewers"`
	MergedAt          *time.Time `json:"merged_at"`
}

type CreatePullRequestRequest struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
}

type MergePullRequestRequest struct {
	PullRequestID string `json:"pull_request_id"`
}

type ReassignPullRequestRequest struct {
	PullRequestID string `json:"pull_request_id"`
	OldUserID     string `json:"old_reviewer_id"`
}

type ReassignPullRequestResponse struct {
	PR         PullRequest `json:"pr"`
	ReplacedBy string      `json:"replaced_by"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Info    string `json:"info"`
}

type ErrorResponse struct {
	Error APIError `json:"error"`
}

func TestCreatePullRequest(t *testing.T) {
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

	createReq := CreatePullRequestRequest{
		PullRequestID:   "pr1",
		PullRequestName: "Initial PR",
		AuthorID:        "u-author",
	}

	var prResp PullRequest
	_ = e.POST("/pullRequest/create").
		WithHeader("Content-Type", "application/json").
		WithJSON(createReq).
		Expect().
		Status(http.StatusCreated).
		JSON().Object().Decode(&prResp)

	req.Equal(createReq.PullRequestID, prResp.PullRequestID)
	req.Equal(createReq.PullRequestName, prResp.PullRequestName)
	req.Equal(createReq.AuthorID, prResp.AuthorID)
	req.Equal("OPEN", prResp.Status)

	req.Len(prResp.AssignedReviewers, 2)

	expectedReviewers := map[string]struct{}{
		"u-review-1": {},
		"u-review-2": {},
	}

	for _, rID := range prResp.AssignedReviewers {
		_, ok := expectedReviewers[rID]
		req.True(ok, "assigned reviewer should be a team member and not the author")
	}
}

func TestMergePullRequest(t *testing.T) {
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

	createReq := CreatePullRequestRequest{
		PullRequestID:   "pr1",
		PullRequestName: "Merge PR",
		AuthorID:        "u-author",
	}

	var createdPR PullRequest
	_ = e.POST("/pullRequest/create").
		WithHeader("Content-Type", "application/json").
		WithJSON(createReq).
		Expect().
		Status(http.StatusCreated).
		JSON().Object().Decode(&createdPR)

	mergeReq := MergePullRequestRequest{
		PullRequestID: createReq.PullRequestID,
	}

	var mergedPR PullRequest
	_ = e.POST("/pullRequest/merge").
		WithHeader("Content-Type", "application/json").
		WithJSON(mergeReq).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Decode(&mergedPR)

	req.Equal(createReq.PullRequestID, mergedPR.PullRequestID)
	req.Equal("MERGED", mergedPR.Status)
	req.NotNil(mergedPR.MergedAt)
	req.Equal(createdPR.AuthorID, mergedPR.AuthorID)
}

func TestReassignPullRequest(t *testing.T) {
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
			{
				UserID:   "u-review-3",
				Username: "Reviewer3",
				IsActive: true,
			},
		},
	}

	_ = e.POST("/team/add").
		WithHeader("Content-Type", "application/json").
		WithJSON(team).
		Expect().
		Status(http.StatusCreated)

	createReq := CreatePullRequestRequest{
		PullRequestID:   "pr1",
		PullRequestName: "Reassign PR",
		AuthorID:        "u-author",
	}

	var createdPR PullRequest
	_ = e.POST("/pullRequest/create").
		WithHeader("Content-Type", "application/json").
		WithJSON(createReq).
		Expect().
		Status(http.StatusCreated).
		JSON().Object().Decode(&createdPR)

	req.Len(createdPR.AssignedReviewers, 2)

	oldReviewerID := createdPR.AssignedReviewers[0]

	reassignReq := ReassignPullRequestRequest{
		PullRequestID: createReq.PullRequestID,
		OldUserID:     oldReviewerID,
	}

	var reassignResp ReassignPullRequestResponse
	_ = e.POST("/pullRequest/reassign").
		WithHeader("Content-Type", "application/json").
		WithJSON(reassignReq).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Decode(&reassignResp)

	req.Equal(createReq.PullRequestID, reassignResp.PR.PullRequestID)
	req.NotEmpty(reassignResp.ReplacedBy)

	req.NotEqual(oldReviewerID, reassignResp.ReplacedBy)
	req.NotEqual(createReq.AuthorID, reassignResp.ReplacedBy)

	req.Len(reassignResp.PR.AssignedReviewers, len(createdPR.AssignedReviewers))

	foundOld := false
	foundNew := false
	for _, rID := range reassignResp.PR.AssignedReviewers {
		if rID == oldReviewerID {
			foundOld = true
		}
		if rID == reassignResp.ReplacedBy {
			foundNew = true
		}
	}

	req.False(foundOld, "old reviewer should be removed from assigned reviewers")
	req.True(foundNew, "new reviewer should be present in assigned reviewers")
}

func TestCreatePullRequest_ValidationError(t *testing.T) {
	req := require.New(t)
	e := newExpect(t)

	cleanup, err := setupPostgres()
	req.NoError(err)
	if cleanup != nil {
		t.Cleanup(cleanup)
	}

	createReq := CreatePullRequestRequest{
		PullRequestID:   "",
		PullRequestName: "Invalid PR",
		AuthorID:        "author-invalid",
	}

	var errResp ErrorResponse
	_ = e.POST("/pullRequest/create").
		WithHeader("Content-Type", "application/json").
		WithJSON(createReq).
		Expect().
		Status(http.StatusBadRequest).
		JSON().Object().Decode(&errResp)

	req.Equal("INVALID_INPUT", errResp.Error.Code)
}

func TestCreatePullRequest_AuthorNotFound(t *testing.T) {
	req := require.New(t)
	e := newExpect(t)

	cleanup, err := setupPostgres()
	req.NoError(err)
	if cleanup != nil {
		t.Cleanup(cleanup)
	}

	createReq := CreatePullRequestRequest{
		PullRequestID:   "pr1",
		PullRequestName: "Missing author PR",
		AuthorID:        "user-missing",
	}

	var errResp ErrorResponse
	_ = e.POST("/pullRequest/create").
		WithHeader("Content-Type", "application/json").
		WithJSON(createReq).
		Expect().
		Status(http.StatusNotFound).
		JSON().Object().Decode(&errResp)

	req.Equal("NOT_FOUND", errResp.Error.Code)
}

func TestCreatePullRequest_DuplicateID(t *testing.T) {
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

	createReq := CreatePullRequestRequest{
		PullRequestID:   "pr1",
		PullRequestName: "Initial PR duplicate",
		AuthorID:        "u-author",
	}

	_ = e.POST("/pullRequest/create").
		WithHeader("Content-Type", "application/json").
		WithJSON(createReq).
		Expect().
		Status(http.StatusCreated)

	var errResp ErrorResponse
	_ = e.POST("/pullRequest/create").
		WithHeader("Content-Type", "application/json").
		WithJSON(createReq).
		Expect().
		Status(http.StatusBadRequest).
		JSON().Object().Decode(&errResp)

	req.Equal("PR_EXISTS", errResp.Error.Code)
}

func TestMergePullRequest_NotFound(t *testing.T) {
	req := require.New(t)
	e := newExpect(t)

	cleanup, err := setupPostgres()
	req.NoError(err)
	if cleanup != nil {
		t.Cleanup(cleanup)
	}

	mergeReq := MergePullRequestRequest{
		PullRequestID: "pr-non-existent",
	}

	var errResp ErrorResponse
	_ = e.POST("/pullRequest/merge").
		WithHeader("Content-Type", "application/json").
		WithJSON(mergeReq).
		Expect().
		Status(http.StatusNotFound).
		JSON().Object().Decode(&errResp)

	req.Equal("NOT_FOUND", errResp.Error.Code)
}

func TestReassignPullRequest_PRNotFound(t *testing.T) {
	req := require.New(t)
	e := newExpect(t)

	cleanup, err := setupPostgres()
	req.NoError(err)
	if cleanup != nil {
		t.Cleanup(cleanup)
	}

	reassignReq := ReassignPullRequestRequest{
		PullRequestID: "pr-reassign-missing",
		OldUserID:     "user-any",
	}

	var errResp ErrorResponse
	_ = e.POST("/pullRequest/reassign").
		WithHeader("Content-Type", "application/json").
		WithJSON(reassignReq).
		Expect().
		Status(http.StatusNotFound).
		JSON().Object().Decode(&errResp)

	req.Equal("NOT_FOUND", errResp.Error.Code)
}

func TestReassignPullRequest_UserNotFound(t *testing.T) {
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
		},
	}

	_ = e.POST("/team/add").
		WithHeader("Content-Type", "application/json").
		WithJSON(team).
		Expect().
		Status(http.StatusCreated)

	createReq := CreatePullRequestRequest{
		PullRequestID:   "pr1",
		PullRequestName: "Reassign user not found",
		AuthorID:        "u-author",
	}

	_ = e.POST("/pullRequest/create").
		WithHeader("Content-Type", "application/json").
		WithJSON(createReq).
		Expect().
		Status(http.StatusCreated)

	reassignReq := ReassignPullRequestRequest{
		PullRequestID: createReq.PullRequestID,
		OldUserID:     "user-not-found",
	}

	var errResp ErrorResponse
	_ = e.POST("/pullRequest/reassign").
		WithHeader("Content-Type", "application/json").
		WithJSON(reassignReq).
		Expect().
		Status(http.StatusNotFound).
		JSON().Object().Decode(&errResp)

	req.Equal("NOT_FOUND", errResp.Error.Code)
}

func TestReassignPullRequest_Merged(t *testing.T) {
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

	createReq := CreatePullRequestRequest{
		PullRequestID:   "pr1",
		PullRequestName: "Reassign merged PR",
		AuthorID:        "u-author",
	}

	var createdPR PullRequest
	_ = e.POST("/pullRequest/create").
		WithHeader("Content-Type", "application/json").
		WithJSON(createReq).
		Expect().
		Status(http.StatusCreated).
		JSON().Object().Decode(&createdPR)

	req.NotEmpty(createdPR.AssignedReviewers)

	mergeReq := MergePullRequestRequest{
		PullRequestID: createReq.PullRequestID,
	}

	_ = e.POST("/pullRequest/merge").
		WithHeader("Content-Type", "application/json").
		WithJSON(mergeReq).
		Expect().
		Status(http.StatusOK)

	reassignReq := ReassignPullRequestRequest{
		PullRequestID: createReq.PullRequestID,
		OldUserID:     createdPR.AssignedReviewers[0],
	}

	var errResp ErrorResponse
	_ = e.POST("/pullRequest/reassign").
		WithHeader("Content-Type", "application/json").
		WithJSON(reassignReq).
		Expect().
		Status(http.StatusBadRequest).
		JSON().Object().Decode(&errResp)

	req.Equal("PR_MERGED", errResp.Error.Code)
}

func TestReassignPullRequest_ReviewerNotAssigned(t *testing.T) {
	req := require.New(t)
	e := newExpect(t)

	cleanup, err := setupPostgres()
	req.NoError(err)
	if cleanup != nil {
		t.Cleanup(cleanup)
	}

	teamWithPR := Team{
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

	otherTeam := Team{
		TeamName: "team-2",
		Members: []TeamMember{
			{
				UserID:   "u-not-assigned-reviewer",
				Username: "NotAssigned",
				IsActive: true,
			},
		},
	}

	_ = e.POST("/team/add").
		WithHeader("Content-Type", "application/json").
		WithJSON(teamWithPR).
		Expect().
		Status(http.StatusCreated)

	_ = e.POST("/team/add").
		WithHeader("Content-Type", "application/json").
		WithJSON(otherTeam).
		Expect().
		Status(http.StatusCreated)

	createReq := CreatePullRequestRequest{
		PullRequestID:   "pr1",
		PullRequestName: "Reassign not assigned",
		AuthorID:        "u-author",
	}

	_ = e.POST("/pullRequest/create").
		WithHeader("Content-Type", "application/json").
		WithJSON(createReq).
		Expect().
		Status(http.StatusCreated)

	reassignReq := ReassignPullRequestRequest{
		PullRequestID: createReq.PullRequestID,
		OldUserID:     "u-not-assigned-reviewer",
	}

	var errResp ErrorResponse
	_ = e.POST("/pullRequest/reassign").
		WithHeader("Content-Type", "application/json").
		WithJSON(reassignReq).
		Expect().
		Status(http.StatusBadRequest).
		JSON().Object().Decode(&errResp)

	req.Equal("NOT_ASSIGNED", errResp.Error.Code)
}

func TestReassignPullRequest_NoCandidate(t *testing.T) {
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
		},
	}

	_ = e.POST("/team/add").
		WithHeader("Content-Type", "application/json").
		WithJSON(team).
		Expect().
		Status(http.StatusCreated)

	createReq := CreatePullRequestRequest{
		PullRequestID:   "pr1",
		PullRequestName: "Reassign no candidate",
		AuthorID:        "u-author",
	}

	var createdPR PullRequest
	_ = e.POST("/pullRequest/create").
		WithHeader("Content-Type", "application/json").
		WithJSON(createReq).
		Expect().
		Status(http.StatusCreated).
		JSON().Object().Decode(&createdPR)

	req.Len(createdPR.AssignedReviewers, 1)

	reassignReq := ReassignPullRequestRequest{
		PullRequestID: createReq.PullRequestID,
		OldUserID:     createdPR.AssignedReviewers[0],
	}

	var errResp ErrorResponse
	_ = e.POST("/pullRequest/reassign").
		WithHeader("Content-Type", "application/json").
		WithJSON(reassignReq).
		Expect().
		Status(http.StatusConflict).
		JSON().Object().Decode(&errResp)

	req.Equal("NO_CANDIDATE", errResp.Error.Code)
}
