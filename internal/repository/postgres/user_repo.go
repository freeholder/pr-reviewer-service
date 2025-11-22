package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/freeholder/pr-reviewer-service/internal/domain"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) UpdateUsers(ctx context.Context, users []domain.User) error {
	if len(users) == 0 {
		return nil
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

	const query = "INSERT INTO users (user_id, username, team_name, is_active) VALUES ($1, $2, $3, $4) ON CONFLICT (user_id) DO UPDATE SET username = EXCLUDED.username, team_name = EXCLUDED.team_name, is_active = EXCLUDED.is_active;"

	for _, u := range users {
		if err := u.Validate(); err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, query, string(u.ID), u.Username, string(u.TeamName), u.IsActive)

		if err != nil {
			return fmt.Errorf("update user %s: %w", u.ID, err)
		}
	}
	return nil
}

func (r *UserRepo) GetUserByID(ctx context.Context, id domain.UserID) (domain.User, error) {
	var u domain.User
	var teamName string

	err := r.db.QueryRowContext(ctx, "SELECT user_id, username, team_name, is_active FROM users WHERE user_id = $1", string(id)).Scan(&id, &u.Username, &teamName, &u.IsActive)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, domain.NewDomainError(domain.ErrNotFound, "user not found")
		}
		return domain.User{}, fmt.Errorf("get user by id: %w", err)
	}
	u.ID = id
	u.TeamName = domain.TeamName(teamName)
	return u, nil
}

func (r *UserRepo) SetUserActive(ctx context.Context, id domain.UserID, isActive bool) (domain.User, error) {
	var u domain.User
	var userID, username, teamName string

	err := r.db.QueryRowContext(ctx, "UPDATE users SET is_active = $1 WHERE user_id = $2 RETURNING user_id, username, team_name, is_active", isActive, string(id)).Scan(&userID, &username, &teamName, &u.IsActive)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, domain.NewDomainError(domain.ErrNotFound, "user not founds")
		}
		return domain.User{}, fmt.Errorf("set user active: %w", err)
	}

	u.ID = domain.UserID(userID)
	u.Username = username
	u.TeamName = domain.TeamName(teamName)

	return u, nil
}

func (r *UserRepo) GetActiveTeamMembersExcept(ctx context.Context, teamName domain.TeamName, exclude []domain.UserID) ([]domain.User, error) {
	query := "SELECT user_id, username, team_name, is_active FROM users WHERE team_name = $1 AND is_active = TRUE"

	args := []any{string(teamName)}

	for i, id := range exclude {
		query += fmt.Sprintf(" AND user_id <> $%d", i+2)
		args = append(args, string(id))
	}

	rows, err := r.db.QueryContext(ctx, query, args...)

	if err != nil {
		return nil, fmt.Errorf("get active team members: %w", err)
	}
	defer rows.Close()

	var result []domain.User

	for rows.Next() {
		var u domain.User
		var userID, username, tn string

		if err := rows.Scan(&userID, &username, &tn, &u.IsActive); err != nil {
			return nil, fmt.Errorf("scan active member: %w", err)
		}

		u.ID = domain.UserID(userID)
		u.Username = username
		u.TeamName = domain.TeamName(tn)

		result = append(result, u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate active members: %w", err)
	}
	return result, nil
}
