package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/freeholder/pr-reviewer-service/internal/domain"
)

type TeamService struct {
	logger *slog.Logger
	teams  TeamRepository
	users  UserRepository
}

func NewTeamService(logger *slog.Logger, teams TeamRepository, users UserRepository) *TeamService {
	return &TeamService{
		logger: logger,
		teams:  teams,
		users:  users,
	}
}

func (s *TeamService) AddTeam(ctx context.Context, team domain.Team) (domain.Team, error) {
	if err := team.Validate(); err != nil {
		return domain.Team{}, err
	}

	for i := range team.Members {
		team.Members[i].TeamName = team.Name
	}

	if err := s.teams.CreateTeam(ctx, team); err != nil {
		s.logger.Error("create team", slog.String("team", string(team.Name)), slog.Any("err", err))
		return domain.Team{}, err
	}

	if err := s.users.UpdateUsers(ctx, team.Members); err != nil {
		s.logger.Error("upsert team members", slog.String("team", string(team.Name)), slog.Any("err", err))
		return domain.Team{}, fmt.Errorf("upsert team members: %w", err)
	}

	result, err := s.teams.GetTeamByName(ctx, team.Name)
	if err != nil {
		s.logger.Error("get team after create", slog.String("team", string(team.Name)), slog.Any("err", err))
		return domain.Team{}, err
	}

	return result, nil
}

func (s *TeamService) GetTeam(ctx context.Context, name domain.TeamName) (domain.Team, error) {
	team, err := s.teams.GetTeamByName(ctx, name)
	if err != nil {
		s.logger.Error("get team", slog.String("team", string(name)), slog.Any("err", err))
		return domain.Team{}, err
	}
	return team, nil
}
