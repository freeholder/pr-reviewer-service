package service

import (
	"context"
	"log/slog"

	"github.com/freeholder/pr-reviewer-service/internal/domain"
)

type UserService struct {
	logger *slog.Logger
	users  UserRepository
	prs    PullRequestRepository
}

func NewUserService(logger *slog.Logger, users UserRepository, prs PullRequestRepository) *UserService {
	return &UserService{
		logger: logger,
		users:  users,
		prs:    prs,
	}
}

func (s *UserService) SetIsActive(ctx context.Context, id domain.UserID, isActive bool) (domain.User, error) {
	user, err := s.users.SetUserActive(ctx, id, isActive)
	if err != nil {
		s.logger.Error("set user is_active", slog.String("user_id", string(id)), slog.Bool("is_active", isActive), slog.Any("err", err))
		return domain.User{}, err
	}
	return user, nil
}

func (s *UserService) ListReviewPRs(ctx context.Context, id domain.UserID) (domain.User, []domain.PullRequest, error) {
	user, err := s.users.GetUserByID(ctx, id)
	if err != nil {
		s.logger.Error("get user before list reviews", slog.String("user_id", string(id)), slog.Any("err", err))
		return domain.User{}, nil, err
	}

	prs, err := s.prs.ListByReviewer(ctx, id)
	if err != nil {
		s.logger.Error("list prs by reviewer", slog.String("user_id", string(id)), slog.Any("err", err))
		return domain.User{}, nil, err
	}

	return user, prs, nil
}
