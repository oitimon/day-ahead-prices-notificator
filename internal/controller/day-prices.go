package controller

import (
	"github.com/oitimon/day-ahead-prices-notificator/internal/dayahead"
	"log"
	"net/http"
	"time"
)

func DayPricesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	day := ctx.Value("day").(time.Time)
	da := ctx.Value("dayahead").(dayahead.DayAhead)

	// Validate the day
	if err := da.ValidateDay(day); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Generate the chart as HTML
	html, err := da.GetHtmlChart(day)
	if err != nil {
		log.Println(err)
		http.Error(w, "server chart error", http.StatusInternalServerError)
		return
	}

	w.Write(html)
}
