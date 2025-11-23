package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/freeholder/pr-reviewer-service/internal/domain"
)

type PRRepo struct {
	db *sql.DB
}

func NewPRRepo(db *sql.DB) *PRRepo {
	return &PRRepo{db: db}
}

func (r *PRRepo) Create(ctx context.Context, pr domain.PullRequest) (err error) {
	if err := pr.Validate(); err != nil {
		return err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	_, err = tx.ExecContext(ctx, "INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status) VALUES ($1, $2, $3, $4)", string(pr.ID), pr.Name, string(pr.AuthorID), string(pr.Status))
	if err != nil {
		if isUnique(err) {
			return domain.NewDomainError(domain.ErrPRExists, "pull request already exists")
		}
		return fmt.Errorf("insert pull_request: %w", err)
	}

	if len(pr.AssignedReviewers) > 0 {
		const insertReviewer = "INSERT INTO pull_request_reviewers (pull_request_id, reviewer_id) VALUES ($1, $2)"

		for _, reviewerID := range pr.AssignedReviewers {
			_, err = tx.ExecContext(ctx, insertReviewer, string(pr.ID), string(reviewerID))
			if err != nil {
				return fmt.Errorf("insert reviewer %s: %w", reviewerID, err)
			}
		}
	}
	return nil
}

func (r *PRRepo) GetByID(ctx context.Context, id domain.PullRequestID) (domain.PullRequest, error) {
	const query = "SELECT p.pull_request_id, p.pull_request_name, p.author_id, p.status, p.created_at, p.merged_at, r.reviewer_id FROM pull_requests p LEFT JOIN pull_request_reviewers r ON r.pull_request_id = p.pull_request_id WHERE p.pull_request_id = $1"

	rows, err := r.db.QueryContext(ctx, query, string(id))
	if err != nil {
		return domain.PullRequest{}, fmt.Errorf("get pull_request: %w", err)
	}
	defer rows.Close()

	var pr domain.PullRequest
	var reviewers []domain.UserID
	found := false

	for rows.Next() {
		var (
			prID, name, authorID, status string
			createdAt                    time.Time
			mergedAt                     sql.NullTime
			reviewerID                   sql.NullString
		)

		if err := rows.Scan(&prID, &name, &authorID, &status, &createdAt, &mergedAt, &reviewerID); err != nil {
			return domain.PullRequest{}, fmt.Errorf("scan pull_request row: %w", err)
		}

		if !found {
			pr.ID = domain.PullRequestID(prID)
			pr.Name = name
			pr.AuthorID = domain.UserID(authorID)
			pr.Status = domain.PRStatus(status)

			pr.CreatedAt = &createdAt

			if mergedAt.Valid {
				t := mergedAt.Time
				pr.MergedAt = &t
			}

			found = true
		}

		if reviewerID.Valid {
			reviewers = append(reviewers, domain.UserID(reviewerID.String))
		}
	}

	if err := rows.Err(); err != nil {
		return domain.PullRequest{}, fmt.Errorf("iterate pull_request rows: %w", err)
	}

	if !found {
		return domain.PullRequest{}, domain.NewDomainError(domain.ErrNotFound, "pull request not found")
	}

	pr.AssignedReviewers = reviewers

	return pr, nil
}

func (r *PRRepo) SetMerged(ctx context.Context, id domain.PullRequestID, mergedAt time.Time) (domain.PullRequest, error) {
	const query = "UPDATE pull_requests SET status = 'MERGED', merged_at = COALESCE(merged_at, $2) WHERE pull_request_id = $1 RETURNING pull_request_id, pull_request_name, author_id, status, created_at, merged_at"
	var (
		prID, name, authorID, status string
		createdAt                    time.Time
		merged                       sql.NullTime
	)

	err := r.db.QueryRowContext(ctx, query, string(id), mergedAt).Scan(&prID, &name, &authorID, &status, &createdAt, &merged)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.PullRequest{}, domain.NewDomainError(domain.ErrNotFound, "pull request not found")
		}
		return domain.PullRequest{}, fmt.Errorf("set merged: %w", err)
	}

	pr := domain.PullRequest{
		ID:       domain.PullRequestID(prID),
		Name:     name,
		AuthorID: domain.UserID(authorID),
		Status:   domain.PRStatus(status),
	}

	pr.CreatedAt = &createdAt
	if merged.Valid {
		t := merged.Time
		pr.MergedAt = &t
	}

	reviewers, err := r.loadReviewers(ctx, pr.ID)
	if err != nil {
		return domain.PullRequest{}, err
	}
	pr.AssignedReviewers = reviewers

	return pr, nil

}

func (r *PRRepo) loadReviewers(ctx context.Context, id domain.PullRequestID) ([]domain.UserID, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT reviewer_id FROM pull_request_reviewers WHERE pull_request_id = $1", string(id))
	if err != nil {
		return nil, fmt.Errorf("load reviewers: %w", err)
	}
	defer rows.Close()

	var reviewers []domain.UserID
	for rows.Next() {
		var rid string
		if err := rows.Scan(&rid); err != nil {
			return nil, fmt.Errorf("scan reviewer: %w", err)
		}
		reviewers = append(reviewers, domain.UserID(rid))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate reviewers: %w", err)
	}

	return reviewers, nil
}

func (r *PRRepo) ReplaceReviewer(ctx context.Context, prID domain.PullRequestID, oldReviewerID, newReviewerID domain.UserID) (pr domain.PullRequest, err error) {
	tx, err := r.db.BeginTx(ctx, nil)

	if err != nil {
		return domain.PullRequest{}, fmt.Errorf("begin tx: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	res, err := tx.ExecContext(ctx, "DELETE FROM pull_request_reviewers WHERE pull_request_id = $1 AND reviewer_id = $2", string(prID), string(oldReviewerID))

	if err != nil {
		return domain.PullRequest{}, fmt.Errorf("delete old reviewer: %w", err)
	}

	affected, err := res.RowsAffected()

	if err != nil {
		return domain.PullRequest{}, fmt.Errorf("rows affected: %w", err)
	}
	if affected == 0 {
		return domain.PullRequest{}, domain.NewDomainError(domain.ErrNotAssigned, "reviewer is not assigned to this pull request")
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO pull_request_reviewers (pull_request_id, reviewer_id)VALUES ($1, $2)", string(prID), string(newReviewerID))

	if err != nil {
		return domain.PullRequest{}, fmt.Errorf("insert new reviewer: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return domain.PullRequest{}, fmt.Errorf("commit tx: %w", err)
	}

	pr, err = r.GetByID(ctx, prID)
	if err != nil {
		return domain.PullRequest{}, err
	}

	return pr, nil
}

func (r *PRRepo) ListByReviewer(ctx context.Context, reviewerID domain.UserID) ([]domain.PullRequest, error) {
	const query = "SELECT p.pull_request_id, p.pull_request_name, p.author_id, p.status, p.created_at, p.merged_at FROM pull_requests p JOIN pull_request_reviewers r ON r.pull_request_id = p.pull_request_id WHERE r.reviewer_id = $1 ORDER BY p.created_at DESC, p.pull_request_id"

	rows, err := r.db.QueryContext(ctx, query, string(reviewerID))
	if err != nil {
		return nil, fmt.Errorf("list PRs by reviewer: %w", err)
	}
	defer rows.Close()

	var result []domain.PullRequest

	for rows.Next() {
		var (
			id, name, authorID, status string
			createdAt                  time.Time
			mergedAt                   sql.NullTime
		)

		if err := rows.Scan(&id, &name, &authorID, &status, &createdAt, &mergedAt); err != nil {
			return nil, fmt.Errorf("scan PR: %w", err)
		}

		pr := domain.PullRequest{
			ID:       domain.PullRequestID(id),
			Name:     name,
			AuthorID: domain.UserID(authorID),
			Status:   domain.PRStatus(status),
		}

		pr.CreatedAt = &createdAt
		if mergedAt.Valid {
			t := mergedAt.Time
			pr.MergedAt = &t
		}

		result = append(result, pr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate PRs by reviewer: %w", err)
	}

	return result, nil
}

func (r *PRRepo) GetOpenPRIDsByReviewer(ctx context.Context, reviewerID domain.UserID) ([]domain.PullRequestID, error) {
	const query = "SELECT pr.pull_request_id FROM pull_requests pr JOIN pull_request_reviewers prr ON prr.pull_request_id = pr.pull_request_id WHERE prr.reviewer_id = $1 AND pr.status = 'OPEN';"
	rows, err := r.db.QueryContext(ctx, query, string(reviewerID))
	if err != nil {
		return nil, fmt.Errorf("get open prs by reviewer: %w", err)
	}
	defer rows.Close()

	var result []domain.PullRequestID

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan pr id: %w", err)
		}
		result = append(result, domain.PullRequestID(id))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate open prs by reviewer: %w", err)
	}

	return result, nil
}
