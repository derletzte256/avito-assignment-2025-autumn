package handlers

import (
	"avito-assignment-2025-autumn/internal/delivery/http/dto"
	"avito-assignment-2025-autumn/internal/entity"
	"time"
)

func dtoTeamToEntity(in dto.CreateTeamRequest) *entity.Team {
	team := &entity.Team{
		Name:    in.TeamName,
		Members: make([]*entity.User, len(in.Members)),
	}

	for i, member := range in.Members {
		team.Members[i] = &entity.User{
			ID:       member.UserID,
			Username: member.Username,
			IsActive: member.IsActive,
		}
	}

	return team
}

func entityTeamToDTO(team *entity.Team) dto.Team {
	dtoTeam := dto.Team{
		TeamName: team.Name,
		Members:  make([]dto.TeamMember, len(team.Members)),
	}

	for i, member := range team.Members {
		if member == nil {
			continue
		}

		dtoTeam.Members[i] = dto.TeamMember{
			UserID:   member.ID,
			Username: member.Username,
			IsActive: member.IsActive,
		}
	}

	return dtoTeam
}

func entityUserToDTO(user *entity.User) dto.User {
	if user == nil {
		return dto.User{}
	}

	return dto.User{
		UserID:   user.ID,
		Username: user.Username,
		TeamName: user.TeamName,
		IsActive: user.IsActive,
	}
}

func entityPullRequestToDTO(pr *entity.PullRequest) dto.PullRequest {
	if pr == nil {
		return dto.PullRequest{}
	}

	reviewers := make([]string, len(pr.Reviewers))
	for i, reviewer := range pr.Reviewers {
		reviewers[i] = reviewer
	}

	var createdAtPtr *time.Time
	if !pr.CreatedAt.IsZero() {
		created := pr.CreatedAt
		createdAtPtr = &created
	}

	return dto.PullRequest{
		PullRequestID:     pr.ID,
		PullRequestName:   pr.Name,
		AuthorID:          pr.AuthorID,
		Status:            dto.PullRequestStatus(pr.Status),
		AssignedReviewers: reviewers,
		CreatedAt:         createdAtPtr,
		MergedAt:          pr.MergedAt,
	}
}

func entityPullRequestToShortDTO(pr *entity.PullRequest) dto.PullRequestShort {
	if pr == nil {
		return dto.PullRequestShort{}
	}

	return dto.PullRequestShort{
		PullRequestID:   pr.ID,
		PullRequestName: pr.Name,
		AuthorID:        pr.AuthorID,
		Status:          dto.PullRequestStatus(pr.Status),
	}
}

func entityPullRequestsToShortDTOs(prs []*entity.PullRequest) []dto.PullRequestShort {
	if len(prs) == 0 {
		return nil
	}

	items := make([]dto.PullRequestShort, 0, len(prs))
	for _, pr := range prs {
		items = append(items, entityPullRequestToShortDTO(pr))
	}

	return items
}
