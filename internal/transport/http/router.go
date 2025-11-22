package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(h *Handler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// r.Route("/users", func(r chi.Router) {
	// 	r.Post("/setIsActive", h.UserSetIsActive)
	// 	r.Get("/getReview", h.UserGetReview)
	// })
	return r
}
