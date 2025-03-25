package middleware

import (
	"context"
	"github.com/oitimon/day-ahead-prices-notificator/pkg/dayahead"
	"net/http"
)

// DayAheadMiddleware to add Day Ahead Service to context
func DayAheadMiddleware(s dayahead.DayAhead) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "dayahead", s)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
