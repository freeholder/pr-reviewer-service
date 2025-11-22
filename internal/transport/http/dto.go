package http

import (
	"time"

	"github.com/freeholder/pr-reviewer-service/internal/domain"
)

type teamMemberDTO struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type teamDTO struct {
	TeamName string          `json:"team_name"`
	Members  []teamMemberDTO `json:"members"`
}

func teamFromDTO(dto teamDTO) domain.Team {
	members := make([]domain.User, 0, len(dto.Members))
	for _, m := range dto.Members {
		members = append(members, domain.User{
			ID:       domain.UserID(m.UserID),
			Username: m.Username,
			TeamName: domain.TeamName(dto.TeamName),
			IsActive: m.IsActive,
		})
	}

	return domain.Team{
		Name:    domain.TeamName(dto.TeamName),
		Members: members,
	}
}

func teamToDTO(t domain.Team) teamDTO {
	members := make([]teamMemberDTO, 0, len(t.Members))
	for _, m := range t.Members {
		members = append(members, teamMemberDTO{
			UserID:   string(m.ID),
			Username: m.Username,
			IsActive: m.IsActive,
		})
	}

	return teamDTO{
		TeamName: string(t.Name),
		Members:  members,
	}
}

type userDTO struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

func userToDTO(u domain.User) userDTO {
	return userDTO{
		UserID:   string(u.ID),
		Username: u.Username,
		TeamName: string(u.TeamName),
		IsActive: u.IsActive,
	}
}

type pullRequestDTO struct {
	PullRequestID     string     `json:"pull_request_id"`
	PullRequestName   string     `json:"pull_request_name"`
	AuthorID          string     `json:"author_id"`
	Status            string     `json:"status"`
	AssignedReviewers []string   `json:"assigned_reviewers"`
	CreatedAt         *time.Time `json:"createdAt,omitempty"`
	MergedAt          *time.Time `json:"mergedAt,omitempty"`
}

func prToDTO(pr domain.PullRequest) pullRequestDTO {
	reviewers := make([]string, 0, len(pr.AssignedReviewers))
	for _, id := range pr.AssignedReviewers {
		reviewers = append(reviewers, string(id))
	}

	return pullRequestDTO{
		PullRequestID:     string(pr.ID),
		PullRequestName:   pr.Name,
		AuthorID:          string(pr.AuthorID),
		Status:            string(pr.Status),
		AssignedReviewers: reviewers,
		CreatedAt:         pr.CreatedAt,
		MergedAt:          pr.MergedAt,
	}
}

type pullRequestShortDTO struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
	Status          string `json:"status"`
}

func prToShortDTO(pr domain.PullRequest) pullRequestShortDTO {
	return pullRequestShortDTO{
		PullRequestID:   string(pr.ID),
		PullRequestName: pr.Name,
		AuthorID:        string(pr.AuthorID),
		Status:          string(pr.Status),
	}
}
