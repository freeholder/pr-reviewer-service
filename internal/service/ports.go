package service

import (
	"context"
	"time"

	"github.com/freeholder/pr-reviewer-service/internal/domain"
)

type TeamRepository interface {
	CreateTeam(ctx context.Context, team domain.Team) error
	GetTeamByName(ctx context.Context, name domain.TeamName) (domain.Team, error)
}

type UserRepository interface {
	UpdateUsers(ctx context.Context, users []domain.User) error
	GetUserByID(ctx context.Context, id domain.UserID) (domain.User, error)
	SetUserActive(ctx context.Context, id domain.UserID, isActive bool) (domain.User, error)
	GetActiveTeamMembersExcept(ctx context.Context, teamName domain.TeamName, exclude []domain.UserID) ([]domain.User, error)
}

type PullRequestRepository interface {
	Create(ctx context.Context, pr domain.PullRequest) error
	GetByID(ctx context.Context, id domain.PullRequestID) (domain.PullRequest, error)
	SetMerged(ctx context.Context, id domain.PullRequestID, mergedAt time.Time) (domain.PullRequest, error)
	ReplaceReviewer(ctx context.Context, prID domain.PullRequestID, oldReviewerID, newReviewerID domain.UserID) (domain.PullRequest, error)
	ListByReviewer(ctx context.Context, reviewerID domain.UserID) ([]domain.PullRequest, error)
	GetOpenPRIDsByReviewer(ctx context.Context, reviewerID domain.UserID) ([]domain.PullRequestID, error)
}
