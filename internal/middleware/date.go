package middleware

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/oitimon/day-ahead-prices-notificator/internal/app"
	"net/http"
	"strconv"
	"time"
)

// Middleware to add date values to context
func DateMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		year, err := strconv.Atoi(chi.URLParam(r, "year"))
		if err != nil {
			http.Error(w, "Invalid year value", http.StatusBadRequest)
			return
		}

		month, err := strconv.Atoi(chi.URLParam(r, "month"))
		if err != nil {
			http.Error(w, "Invalid month value", http.StatusBadRequest)
			return
		}

		day, err := strconv.Atoi(chi.URLParam(r, "day"))
		if err != nil {
			http.Error(w, "Invalid day value", http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "day", time.Date(year, time.Month(month), day, 0, 0, 0, 0, r.Context().Value("config").(*app.ConfigApp).Location()))))
	})
}
