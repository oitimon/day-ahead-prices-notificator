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

	// Check if the day is in the future
	tomorrow := time.Now().AddDate(0, 0, 1)
	tomorrow = time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, cfg.Location())
	if day.After(tomorrow) {
		http.Error(w, "Day is in the future after tomorrow", http.StatusNotFound)
		return
	} else if day.Equal(tomorrow) && time.Now().In(cfg.Location()).Hour() < cfg.TomorrowHourMin() {
		http.Error(w, "Day is tomorrow but it's too early", http.StatusNotFound)
		return
	}

	// Fetch prices
	prices, err := app.FetchPrices(&cfg.Loader, day)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Draw and send the chart
	_, err = app.ChartText(&cfg.Analytics, prices, day)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send message
	//if err = app.SendMessage(&cfg.Messenger, message); err != nil {
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//	return
	//}

	// Generate the chart as HTML
	if err = app.ChartHtml(w, &cfg.Analytics, prices, day); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
