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

	r.Route("/team", func(r chi.Router) {
		r.Post("/add", h.TeamAdd)
		r.Get("/get", h.TeamGet)
	})

	r.Route("/users", func(r chi.Router) {
		r.Post("/setIsActive", h.UserSetIsActive)
		r.Get("/getReview", h.UserGetReview)
	})

	r.Route("/pullRequest", func(r chi.Router) {
		r.Post("/create", h.PRCreate)
		r.Post("/merge", h.PRMerge)
		r.Post("/reassign", h.PRReassign)
	})

	r.Get("/health", h.Health)

	return r
}
