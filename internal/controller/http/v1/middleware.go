package v1

import (
	"net/http"
	"time"

	"github.com/MatTwix/Pull-Request-Assigner/pkg/logger"
	"github.com/go-chi/chi/v5/middleware"
)

const metricsRoutePath = "/metrics"

func loggingMiddleware(log logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == metricsRoutePath {
				next.ServeHTTP(w, r)
				return
			}

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			start := time.Now()
			next.ServeHTTP(ww, r)
			duration := time.Since(start)

			log.Info("request completed", map[string]any{
				"method":      r.Method,
				"path":        r.URL.Path,
				"status":      ww.Status(),
				"duration_ms": duration.Milliseconds(),
				"remote_ip":   r.RemoteAddr,
			})
		})
	}
}
