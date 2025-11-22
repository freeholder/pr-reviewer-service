package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/freeholder/pr-reviewer-service/internal/domain"
)

type TeamRepo struct {
	db *sql.DB
}

func NewTeamRepo(db *sql.DB) *TeamRepo {
	return &TeamRepo{db: db}
}

func (r *TeamRepo) CreateTeam(ctx context.Context, team domain.Team) error {
	if err := team.Validate(); err != nil {
		return err
	}

	_, err := r.db.ExecContext(ctx,
		"INSERT INTO teams (team_name) VALUES ($1)",
		string(team.Name))

	if err != nil {
		if isUnique(err) {
			return domain.NewDomainError(domain.ErrTeamExists, "team already exists")
		}
		return fmt.Errorf("insert team: %w", err)
	}
	return nil
}

func (r *TeamRepo) GetTeamByName(ctx context.Context, name domain.TeamName) (domain.Team, error) {
	var teamName string
	err := r.db.QueryRowContext(ctx, "SELECT team_name FROM teams WHERE team_name = $1", string(name)).Scan(&teamName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Team{}, domain.NewDomainError(domain.ErrNotFound, "team not found")
		}
		return domain.Team{}, fmt.Errorf("get team: %w", err)
	}

	rows, err := r.db.QueryContext(ctx,
		"SELECT user_id, username, is_active FROM users WHERE team_name = $1 ORDER BY user_id", string(name))
	if err != nil {
		return domain.Team{}, fmt.Errorf("get team members: %w", err)
	}

	defer rows.Close()

	var members []domain.User
	for rows.Next() {
		var u domain.User
		var userID, username string
		var isActive bool

		if err := rows.Scan(&userID, &username, &isActive); err != nil {
			return domain.Team{}, fmt.Errorf("scan team member: %w", err)
		}

		u.ID = domain.UserID(userID)
		u.Username = username
		u.TeamName = name
		u.IsActive = isActive

		members = append(members, u)
	}
	if err := rows.Err(); err != nil {
		return domain.Team{}, fmt.Errorf("iterate team members: %w", err)
	}

	return domain.Team{
		Name:    name,
		Members: members,
	}, nil
}
