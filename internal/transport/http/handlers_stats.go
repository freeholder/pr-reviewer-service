package http

import "net/http"

type reviewerStatDTO struct {
	UserID        string `json:"user_id"`
	Username      string `json:"username"`
	AssignedCount int64  `json:"assigned_count"`
}

func (h *Handler) GetReviewerStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stats, err := h.statsService.GetReviewerStats(ctx)
	if err != nil {
		h.writeError(w, err)
		return
	}

	resp := struct {
		ReviewerStats []reviewerStatDTO `json:"reviewer_stats"`
	}{
		ReviewerStats: make([]reviewerStatDTO, 0, len(stats)),
	}

	for _, s := range stats {
		resp.ReviewerStats = append(resp.ReviewerStats, reviewerStatDTO{
			UserID:        string(s.UserID),
			Username:      s.Username,
			AssignedCount: s.AssignedCount,
		})
	}

	h.writeJSON(w, http.StatusOK, resp)
}
