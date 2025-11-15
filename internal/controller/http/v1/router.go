package v1

import (
	"net/http"

	"github.com/MatTwix/Pull-Request-Assigner/internal/service"
	"github.com/MatTwix/Pull-Request-Assigner/pkg/logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(r *chi.Mux, services *service.Services, logger logger.Logger) http.Handler {
	authMiddleware := &AuthMiddleware{authService: services.Auth, log: logger}

	r.Use(middleware.Recoverer)
	r.Use(loggingMiddleware(logger))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Route("/team", func(rt chi.Router) {
		team := newTeamRoutes(services.Team, logger)

		rt.Post("/add", team.add)

		rt.With(authMiddleware.APIKeyMiddleware(true)).
			Get("/get", team.get)
	})

	r.Route("/users", func(rt chi.Router) {
		user := newUserRoutes(services.User, logger)

		rt.With(authMiddleware.APIKeyMiddleware(true)).
			Post("/setIsActive", user.setIsActive)

		rt.With(authMiddleware.APIKeyMiddleware(false)).
			Get("/getReview", user.getReview)
	})

	r.Route("/pullRequest", func(rt chi.Router) {
		pr := newPullRequestRoutes(services.PullRequest, logger)
		rt.With(authMiddleware.APIKeyMiddleware(true)).
			Post("/create", pr.create)

		rt.With(authMiddleware.APIKeyMiddleware(true)).
			Post("/merge", pr.merge)

		rt.With(authMiddleware.APIKeyMiddleware(true)).
			Post("/reassign", pr.reassign)
	})

	return r
}
