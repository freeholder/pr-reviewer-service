package http

import (
	"encoding/json"
	"net/http"

	"github.com/freeholder/pr-reviewer-service/internal/domain"
)

type teamAddResponse struct {
	Team teamDTO `json:"team"`
}

type bulkDeactivateRequest struct {
	TeamName string   `json:"team_name"`
	UserIDs  []string `json:"user_ids"`
}

type bulkNotReassignedDTO struct {
	PullRequestID string `json:"pull_request_id"`
	UserID        string `json:"user_id"`
	Reason        string `json:"reason"`
}

type bulkDeactivateResponse struct {
	TeamName           string                 `json:"team_name"`
	DeactivatedUserIDs []string               `json:"deactivated_user_ids"`
	ReassignedCount    int                    `json:"reassigned_count"`
	NotReassigned      []bulkNotReassignedDTO `json:"not_reassigned"`
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

func (h *Handler) BulkDeactivateTeamMembers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req bulkDeactivateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, domain.NewValidationError("body", "invalid json"))
		return
	}

	if req.TeamName == "" {
		h.writeError(w, domain.NewValidationError("team_name", "must not be empty"))
		return
	}

	if len(req.UserIDs) == 0 {
		h.writeError(w, domain.NewValidationError("user_ids", "must not be empty"))
		return
	}

	userIDs := make([]domain.UserID, 0, len(req.UserIDs))
	for _, id := range req.UserIDs {
		userIDs = append(userIDs, domain.UserID(id))
	}

	result, err := h.prService.BulkDeactivateAndReassign(ctx, domain.TeamName(req.TeamName), userIDs)
	if err != nil {
		h.writeError(w, err)
		return
	}

	resp := bulkDeactivateResponse{
		TeamName:           string(result.TeamName),
		DeactivatedUserIDs: make([]string, 0, len(result.DeactivatedUserIDs)),
		ReassignedCount:    result.ReassignedCount,
		NotReassigned:      make([]bulkNotReassignedDTO, 0, len(result.NotReassigned)),
	}

	for _, id := range result.DeactivatedUserIDs {
		resp.DeactivatedUserIDs = append(resp.DeactivatedUserIDs, string(id))
	}

	for _, nr := range result.NotReassigned {
		resp.NotReassigned = append(resp.NotReassigned, bulkNotReassignedDTO{
			PullRequestID: string(nr.PullRequestID),
			UserID:        string(nr.UserID),
			Reason:        nr.Reason,
		})
	}

	h.writeJSON(w, http.StatusOK, resp)
}
