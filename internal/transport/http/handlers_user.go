package http

import (
	"encoding/json"
	"net/http"

	"github.com/freeholder/pr-reviewer-service/internal/domain"
)

type setIsActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

type userResponse struct {
	User userDTO `json:"user"`
}

type userReviewResponse struct {
	UserID       string                `json:"user_id"`
	PullRequests []pullRequestShortDTO `json:"pull_requests"`
}

func (h *Handler) UserSetIsActive(w http.ResponseWriter, r *http.Request) {
	var req setIsActiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, domain.NewValidationError("body", "invalid JSON body"))
		return
	}

	if req.UserID == "" {
		h.writeError(w, domain.NewValidationError("user_id", "must not be empty"))
		return
	}

	user, err := h.userService.SetIsActive(r.Context(), domain.UserID(req.UserID), req.IsActive)
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, http.StatusOK, userResponse{User: userToDTO(user)})
}

func (h *Handler) UserGetReview(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		h.writeError(w, domain.NewValidationError("user_id", "must not be empty"))
		return
	}

	user, prs, err := h.userService.ListReviewPRs(r.Context(), domain.UserID(userID))
	if err != nil {
		h.writeError(w, err)
		return
	}

	resp := userReviewResponse{
		UserID:       string(user.ID),
		PullRequests: make([]pullRequestShortDTO, 0, len(prs)),
	}

	for _, pr := range prs {
		resp.PullRequests = append(resp.PullRequests, prToShortDTO(pr))
	}

	h.writeJSON(w, http.StatusOK, resp)
}
