package controller

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/oitimon/day-ahead-prices-notificator/pkg/config"
	"github.com/oitimon/day-ahead-prices-notificator/pkg/dayahead"
	"github.com/oitimon/day-ahead-prices-notificator/pkg/repository"
)

func NowPriceHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	da := ctx.Value("dayahead").(dayahead.DayAhead)
	cfg := ctx.Value("config").(*config.App)
	vat := ctx.Value("vat").(bool)

	now := time.Now().In(cfg.Location())
	day := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, cfg.Location())

	prices, err := da.GetPrices(ctx, day, repository.WithVat(vat))
	if err != nil {
		http.Error(w, "failed to get prices", http.StatusInternalServerError)
		return
	}

	hour := now.Hour()
	if hour >= len(prices) {
		http.Error(w, "price not available for current hour", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	price, _ := prices[hour].Float64()
	_ = json.NewEncoder(w).Encode(map[string]float64{"price": price})
}
