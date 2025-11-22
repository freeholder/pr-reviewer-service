package http

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type Handler struct {
	logger *slog.Logger
}

func NewHandler(logger *slog.Logger) *Handler {
	return &Handler{
		logger: logger,
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
