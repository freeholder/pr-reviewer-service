package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/freeholder/pr-reviewer-service/internal/domain"
)

type StatsRepo struct {
	db *sql.DB
}

func NewStatsRepo(db *sql.DB) *StatsRepo {
	return &StatsRepo{db: db}
}

func (r *StatsRepo) GetReviewerStats(ctx context.Context) ([]domain.ReviewerStats, error) {
	const query = "SELECT u.user_id, u.username, COUNT(prr.pull_request_id) AS assigned_count FROM users u LEFT JOIN pull_request_reviewers prr ON prr.reviewer_id = u.user_id GROUP BY u.user_id, u.username ORDER BY assigned_count DESC, u.user_id;"

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get reviewer stats: %w", err)
	}
	defer rows.Close()

	var result []domain.ReviewerStats

	for rows.Next() {
		var st domain.ReviewerStats
		if err := rows.Scan(&st.UserID, &st.Username, &st.AssignedCount); err != nil {
			return nil, fmt.Errorf("scan reviewer stat: %w", err)
		}
		result = append(result, st)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate reviewer stats: %w", err)
	}

	return result, nil
}
