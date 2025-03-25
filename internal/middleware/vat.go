package middleware

import (
	"context"
	"github.com/oitimon/day-ahead-prices-notificator/pkg/config"
	"net/http"
	"strconv"
)

const vatKey = "vat"

// VatMiddleware parses the "vat" query param and stores it in context
func VatMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inclVatStr := r.URL.Query().Get("vat")
		// Get default value from config.
		inclVat := r.Context().Value("config").(*config.App).Ui.IncludingVat
		if inclVatStr != "" {
			var err error
			if inclVat, err = strconv.ParseBool(inclVatStr); err != nil {
				http.Error(w, "Invalid 'vat' query parameter. Use 'true' or 'false'.", http.StatusBadRequest)
				return
			}
		}
		ctx := context.WithValue(r.Context(), vatKey, inclVat)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
