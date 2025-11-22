package http

import (
	"errors"
	"net/http"

	"github.com/freeholder/pr-reviewer-service/internal/domain"
)

func (h *Handler) writeError(w http.ResponseWriter, err error) {
	var verr *domain.ValidationError
	if errors.As(err, &verr) {
		h.writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": map[string]any{
				"code":    "BAD_REQUEST",
				"message": verr.Message,
			},
		})
		return
	}
	var derr *domain.DomainError
	if errors.As(err, &derr) {
		status := httpStatusFromDomainCode(derr.Code)
		h.writeJSON(w, status, map[string]any{
			"error": map[string]any{
				"code":    string(derr.Code),
				"message": derr.Message,
			},
		})
		return
	}
	h.logger.Error("internal error", "err", err)
	h.writeJSON(w, http.StatusInternalServerError, map[string]any{
		"error": map[string]any{
			"code":    "INTERNAL_ERROR",
			"message": "internal server error",
		},
	})
}
func httpStatusFromDomainCode(code domain.ErrorCode) int {
	switch code {
	case domain.ErrTeamExists:
		return http.StatusBadRequest // 400
	case domain.ErrPRExists:
		return http.StatusConflict // 409
	case domain.ErrPRMerged, domain.ErrNotAssigned, domain.ErrNoCandidate:
		return http.StatusConflict // 409
	case domain.ErrNotFound:
		return http.StatusNotFound // 404
	default:
		return http.StatusInternalServerError
	}
}
