package controller

import (
	"github.com/oitimon/day-ahead-prices-notificator/pkg/dayahead"
	"github.com/oitimon/day-ahead-prices-notificator/pkg/repository"
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

	//@todo development in progress
	//if err := da.SendMessage(ctx, day, repository.WithVat(true)); err != nil {
	//	log.Println(err)
	//	http.Error(w, "server chart error", http.StatusInternalServerError)
	//	return
	//}

	// Generate the chart as HTML
	html, err := da.GetHtmlChart(ctx, day, repository.WithVat(ctx.Value("vat").(bool)))
	if err != nil {
		log.Println(err)
		http.Error(w, "HTML chart rendering failed", http.StatusInternalServerError)
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
