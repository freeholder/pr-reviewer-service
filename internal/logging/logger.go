package logging

import (
	"log/slog"
	"os"
)

func NewLogger() *slog.Logger {
	handler := slog.NewJSONHandler(os.Stderr, nil)
	return slog.New(handler)
}
