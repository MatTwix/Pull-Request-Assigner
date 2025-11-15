package v1

import (
	"net/http"

	"github.com/MatTwix/Pull-Request-Assigner/internal/service"
	"github.com/MatTwix/Pull-Request-Assigner/pkg/logger"
)

type AuthMiddleware struct {
	authService service.Auth
	log         logger.Logger
}

func (a *AuthMiddleware) APIKeyMiddleware(requireAdmin bool) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("X-Api-Key")
			if apiKey == "" {
				newErrorResponse(w, http.StatusUnauthorized, CodeUserAuthError, "missing API key")
				a.log.Warn("missing API key", map[string]any{"path": r.URL.Path})
				return
			}

			if requireAdmin {
				if !a.authService.IsAdminKey(apiKey) {
					a.log.Warn("invalid admin API key", map[string]any{"key": apiKey, "path": r.URL.Path})
					newErrorResponse(w, http.StatusUnauthorized, CodeAdminAuthError, ErrAdminAuthError.Error())
					return
				}
			} else {
				if !a.authService.IsAdminKey(apiKey) && !a.authService.IsUserKey(apiKey) {
					a.log.Warn("invalid API key", map[string]any{"key": apiKey, "path": r.URL.Path})
					newErrorResponse(w, http.StatusUnauthorized, CodeUserAuthError, ErrUserAuthError.Error())
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
