package controller

import (
	"github.com/oitimon/day-ahead-prices-notificator/internal/dayahead"
	"html/template"
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
	html, err := da.GetHtmlChart(ctx, day)
	if err != nil {
		log.Println(err)
		http.Error(w, "server chart error", http.StatusInternalServerError)
		return
	}

	// Use html/template to render the HTML safely
	tmpl, err := template.New("chart").Parse(string(html))
	if err != nil {
		log.Println(err)
		http.Error(w, "template parsing error", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		log.Println(err)
		http.Error(w, "template execution error", http.StatusInternalServerError)
		return
	}
}
