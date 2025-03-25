package middleware

import (
	"context"
	"github.com/oitimon/day-ahead-prices-notificator/pkg/config"
	"net/http"
)

// ConfigMiddleware Middleware to add config to context
func ConfigMiddleware(cfg *config.App) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "config", cfg)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
