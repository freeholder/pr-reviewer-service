package service

import (
	"context"

	"github.com/freeholder/pr-reviewer-service/internal/domain"
)

type StatsRepository interface {
	GetReviewerStats(ctx context.Context) ([]domain.ReviewerStats, error)
}

type StatsService struct {
	statsRepo StatsRepository
}

func NewStatsService(statsRepo StatsRepository) *StatsService {
	return &StatsService{statsRepo: statsRepo}
}

func (s *StatsService) GetReviewerStats(ctx context.Context) ([]domain.ReviewerStats, error) {
	return s.statsRepo.GetReviewerStats(ctx)
}
