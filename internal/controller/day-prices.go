package controller

import (
	"github.com/oitimon/day-ahead-prices-notificator/internal/app"
	"net/http"
	"time"
)

func DayPricesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cfg := ctx.Value("config").(*app.ConfigApp)
	day := ctx.Value("day").(time.Time)

	prices, err := app.FetchPrices(&cfg.Loader, day)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, price := range prices {
		println(price.String())
	}

	_, _ = w.Write([]byte(day.String()))
}
