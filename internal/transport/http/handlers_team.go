package http

import (
	"encoding/json"
	"net/http"

	"github.com/freeholder/pr-reviewer-service/internal/domain"
)

type teamAddResponse struct {
	Team teamDTO `json:"team"`
}

func (h *Handler) TeamAdd(w http.ResponseWriter, r *http.Request) {
	var req teamDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, domain.NewValidationError("body", "invalid JSON body"))
		return
	}

	team := teamFromDTO(req)

	result, err := h.teamService.AddTeam(r.Context(), team)
	if err != nil {
		h.writeError(w, err)
		return
	}

	resp := teamAddResponse{
		Team: teamToDTO(result),
	}

	h.writeJSON(w, http.StatusCreated, resp)
}

func (h *Handler) TeamGet(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		h.writeError(w, domain.NewValidationError("team_name", "must not be empty"))
		return
	}

	team, err := h.teamService.GetTeam(r.Context(), domain.TeamName(teamName))
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, http.StatusOK, teamToDTO(team))
}
