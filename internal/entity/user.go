package entity

type User struct {
	ID       string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
	TeamName string `json:"team_name,omitempty"`
}

type UserStatistics struct {
	UserID          string `json:"user_id"`
	AuthoredPRCount int    `json:"authored_pr_count"`
	OnReviewPRCount int    `json:"on_review_pr_count"` // Where is reviewer and PR is open
	ReviewedPRCount int    `json:"reviewed_pr_count"`  // Where is reviewer and PR is merged
}
