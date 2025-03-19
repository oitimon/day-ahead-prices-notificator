package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/oitimon/day-ahead-prices-notificator/internal/config"
	"github.com/oitimon/day-ahead-prices-notificator/internal/controller"
	"github.com/oitimon/day-ahead-prices-notificator/internal/dayahead"
	appMiddleware "github.com/oitimon/day-ahead-prices-notificator/internal/middleware"
	"log"
	"net/http"
	"os"
	"time"
)

const chartHtmlFilename = "/tmp/epex_nl_da_prices_chart.html"

func main() {
	cfg := &config.App{}
	if _, err := os.Stat(".env"); err == nil {
		// We load and parse the .env file only if it exists,
		// otherwise we rely on the environment variables.
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}
	if err := envconfig.Process("", cfg); err != nil {
		log.Fatalf("Error processing environment variables: %v", err)
	}
	if err := cfg.SelfCheck(); err != nil {
		log.Fatalf("Error checking configuration: %v", err)
	}

	// Load version from the file.
	data, err := os.ReadFile("VERSION")
	if err != nil {
		log.Fatal("Error opening VERSION file")
	}
	cfg.Analytics.Version = string(data)

	// Prepare service(s).
	da, err := dayahead.NewDayAhead(cfg)
	if err != nil {
		log.Fatalf("Error creating DayAhead service: %v", err)
	}

	// Start the server.
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(appMiddleware.ConfigMiddleware(cfg))
	r.Use(appMiddleware.DayAheadMiddleware(da))
	r.Use(middleware.Timeout(30 * time.Second))
	r.Get("/", controller.IndexHandler)
	r.Get("/api/v1/healthcheck", controller.HealthCheckHandler)
	r.With(appMiddleware.DateMiddleware).Get("/day-prices/{year}-{month}-{day}", controller.DayPricesHandler)
	log.Printf("Starting server on :%s\n", cfg.Server.Port)
	if err := http.ListenAndServe(":"+cfg.Server.Port, r); err != nil {
		log.Fatal(err)
	}
}
