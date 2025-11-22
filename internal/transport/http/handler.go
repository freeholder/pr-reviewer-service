package http

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/freeholder/pr-reviewer-service/internal/service"
)

type Handler struct {
	logger       *slog.Logger
	teamService  *service.TeamService
	userService  *service.UserService
	prService    *service.PRService
	statsService *service.StatsService
}

func NewHandler(logger *slog.Logger, teamService *service.TeamService, userService *service.UserService, prService *service.PRService, statsService *service.StatsService) *Handler {
	return &Handler{
		logger:       logger,
		teamService:  teamService,
		userService:  userService,
		prService:    prService,
		statsService: statsService,
	}

}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if data == nil {
		return
	}

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("write json response", slog.Any("err", err))
	}
}
