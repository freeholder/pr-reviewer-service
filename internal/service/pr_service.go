package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/freeholder/pr-reviewer-service/internal/domain"
	"github.com/freeholder/pr-reviewer-service/internal/random"
)

type PRService struct {
	logger *slog.Logger
	users  UserRepository
	prs    PullRequestRepository
	rand   random.Randomizer
}

func NewPRService(logger *slog.Logger, users UserRepository, prs PullRequestRepository, rand random.Randomizer) *PRService {
	return &PRService{
		logger: logger,
		users:  users,
		prs:    prs,
		rand:   rand,
	}
}
func (s *PRService) Create(ctx context.Context, id domain.PullRequestID, name string, authorID domain.UserID) (domain.PullRequest, error) {
	if id == "" {
		return domain.PullRequest{}, domain.NewValidationError("pull_request_id", "must not be empty")
	}
	if name == "" {
		return domain.PullRequest{}, domain.NewValidationError("pull_request_name", "must not be empty")
	}
	if authorID == "" {
		return domain.PullRequest{}, domain.NewValidationError("author_id", "must not be empty")
	}

	author, err := s.users.GetUserByID(ctx, authorID)
	if err != nil {
		s.logger.Error("get author for pr", slog.String("author_id", string(authorID)), slog.Any("err", err))
		return domain.PullRequest{}, err
	}

	candidates, err := s.users.GetActiveTeamMembersExcept(ctx, author.TeamName, []domain.UserID{author.ID})
	if err != nil {
		s.logger.Error("get candidates for pr reviewers", slog.String("team", string(author.TeamName)), slog.Any("err", err))
		return domain.PullRequest{}, err
	}

	assigned := s.pickReviewers(candidates, 2)

	pr := domain.PullRequest{
		ID:                id,
		Name:              name,
		AuthorID:          authorID,
		Status:            domain.PRStatusOpen,
		AssignedReviewers: assigned,
	}

	if err := pr.Validate(); err != nil {
		return domain.PullRequest{}, err
	}

	if err := s.prs.Create(ctx, pr); err != nil {
		s.logger.Error("create pr", slog.String("pr_id", string(id)), slog.String("author_id", string(authorID)), slog.Any("err", err))
		return domain.PullRequest{}, err
	}

	return pr, nil
}

func (s *PRService) Merge(ctx context.Context, id domain.PullRequestID) (domain.PullRequest, error) {
	if id == "" {
		return domain.PullRequest{}, domain.NewValidationError("pull_request_id", "must not be empty")
	}

	now := time.Now().UTC()

	pr, err := s.prs.SetMerged(ctx, id, now)
	if err != nil {
		s.logger.Error("merge pr", slog.String("pr_id", string(id)), slog.Any("err", err))
		return domain.PullRequest{}, err
	}

	return pr, nil
}

func (s *PRService) ReassignReviewer(ctx context.Context, prID domain.PullRequestID, oldReviewerID domain.UserID) (domain.PullRequest, domain.UserID, error) {
	if prID == "" {
		return domain.PullRequest{}, "", domain.NewValidationError("pull_request_id", "must not be empty")
	}
	if oldReviewerID == "" {
		return domain.PullRequest{}, "", domain.NewValidationError("old_user_id", "must not be empty")
	}

	pr, err := s.prs.GetByID(ctx, prID)
	if err != nil {
		s.logger.Error("get pr before reassign", slog.String("pr_id", string(prID)), slog.Any("err", err))
		return domain.PullRequest{}, "", err
	}

	if pr.Status == domain.PRStatusMerged {
		return domain.PullRequest{}, "", domain.NewDomainError(domain.ErrPRMerged, "cannot reassign on merged PR")
	}

	found := false
	for _, rid := range pr.AssignedReviewers {
		if rid == oldReviewerID {
			found = true
			break
		}
	}
	if !found {
		return domain.PullRequest{}, "", domain.NewDomainError(domain.ErrNotAssigned, "reviewer is not assigned to this PR")
	}

	oldReviewer, err := s.users.GetUserByID(ctx, oldReviewerID)
	if err != nil {
		s.logger.Error("get old reviewer for reassign", slog.String("user_id", string(oldReviewerID)), slog.Any("err", err))
		return domain.PullRequest{}, "", err
	}

	exclude := map[domain.UserID]struct{}{
		oldReviewerID: {},
		pr.AuthorID:   {},
	}
	for _, rid := range pr.AssignedReviewers {
		exclude[rid] = struct{}{}
	}

	var excludeIDs []domain.UserID
	for id := range exclude {
		excludeIDs = append(excludeIDs, id)
	}

	candidates, err := s.users.GetActiveTeamMembersExcept(ctx, oldReviewer.TeamName, excludeIDs)
	if err != nil {
		s.logger.Error("get candidates for reassign", slog.String("team", string(oldReviewer.TeamName)), slog.Any("err", err))
		return domain.PullRequest{}, "", err
	}

	if len(candidates) == 0 {
		return domain.PullRequest{}, "", domain.NewDomainError(domain.ErrNoCandidate, "no active replacement candidate in team")
	}

	chosenIdxs := s.pickReviewers(candidates, 1)
	newReviewerID := chosenIdxs[0]

	updatedPR, err := s.prs.ReplaceReviewer(ctx, prID, oldReviewerID, newReviewerID)
	if err != nil {
		s.logger.Error("replace reviewer", slog.String("pr_id", string(prID)), slog.String("old_reviewer_id", string(oldReviewerID)), slog.String("new_reviewer_id", string(newReviewerID)), slog.Any("err", err))
		return domain.PullRequest{}, "", err
	}

	return updatedPR, newReviewerID, nil
}

func (s *PRService) pickReviewers(candidates []domain.User, limit int) []domain.UserID {
	n := len(candidates)
	if n == 0 || limit <= 0 {
		return nil
	}

	if n <= limit {
		ids := make([]domain.UserID, 0, n)
		for _, u := range candidates {
			ids = append(ids, u.ID)
		}
		return ids
	}

	used := make(map[int]struct{}, limit)
	var result []domain.UserID

	for len(result) < limit {
		i := s.rand.Intn(n)
		if _, exists := used[i]; exists {
			continue
		}
		used[i] = struct{}{}
		result = append(result, candidates[i].ID)
	}

	return result
}
