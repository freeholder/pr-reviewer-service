package http

import (
	"encoding/json"
	"net/http"

	"github.com/freeholder/pr-reviewer-service/internal/domain"
)

type createPRRequest struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
}

type mergePRRequest struct {
	PullRequestID string `json:"pull_request_id"`
}

type reassignRequest struct {
	PullRequestID string `json:"pull_request_id"`
	OldUserID     string `json:"old_user_id"`
}

type prResponse struct {
	PR pullRequestDTO `json:"pr"`
}

type reassignResponse struct {
	PR         pullRequestDTO `json:"pr"`
	ReplacedBy string         `json:"replaced_by"`
}

func (h *Handler) PRCreate(w http.ResponseWriter, r *http.Request) {
	var req createPRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, domain.NewValidationError("body", "invalid JSON body"))
		return
	}

	pr, err := h.prService.Create(r.Context(), domain.PullRequestID(req.PullRequestID), req.PullRequestName, domain.UserID(req.AuthorID))
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, http.StatusCreated, prResponse{PR: prToDTO(pr)})
}

func (h *Handler) PRMerge(w http.ResponseWriter, r *http.Request) {
	var req mergePRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, domain.NewValidationError("body", "invalid JSON body"))
		return
	}

	pr, err := h.prService.Merge(r.Context(), domain.PullRequestID(req.PullRequestID))
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, http.StatusOK, prResponse{PR: prToDTO(pr)})
}

func (h *Handler) PRReassign(w http.ResponseWriter, r *http.Request) {
	var req reassignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, domain.NewValidationError("body", "invalid JSON body"))
		return
	}

	pr, replacedBy, err := h.prService.ReassignReviewer(
		r.Context(),
		domain.PullRequestID(req.PullRequestID),
		domain.UserID(req.OldUserID),
	)
	if err != nil {
		h.writeError(w, err)
		return
	}

	resp := reassignResponse{
		PR:         prToDTO(pr),
		ReplacedBy: string(replacedBy),
	}

	h.writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
