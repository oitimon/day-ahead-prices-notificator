package middleware

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/oitimon/day-ahead-prices-notificator/pkg/config"
	"net/http"
	"strings"
	"time"
)

// Middleware to add date values to context
func DateMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg := r.Context().Value("config").(*config.App)
		var date time.Time
		var err error
		switch strings.ToLower(chi.URLParam(r, "day")) {
		case "today":
			date = time.Now()
		case "yesterday":
			date = time.Now().AddDate(0, 0, -1)
		case "tomorrow":
			date = time.Now().AddDate(0, 0, 1)
		default:
			if date, err = time.Parse("2006-01-02", chi.URLParam(r, "day")); err != nil {
				http.Error(w, "Invalid date format: "+err.Error(), http.StatusBadRequest)
				return
			}
		}
		date = date.In(cfg.Location())
		date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, cfg.Location())

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "day", date)))
	})
}
