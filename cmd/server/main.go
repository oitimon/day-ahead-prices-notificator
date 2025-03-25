package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/oitimon/day-ahead-prices-notificator/internal/controller"
	appMiddleware "github.com/oitimon/day-ahead-prices-notificator/internal/middleware"
	"github.com/oitimon/day-ahead-prices-notificator/pkg/config"
	"github.com/oitimon/day-ahead-prices-notificator/pkg/dayahead"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const chartHtmlFilename = "/tmp/epex_nl_da_prices_chart.html"

func main() {
	cfg := loadConfig()

	// Creating the context.
	ctx, cancel := context.WithCancel(context.Background())

	// Load version from the file.
	data, err := os.ReadFile("VERSION")
	if err != nil {
		log.Fatal("Error opening VERSION file")
	}
	cfg.Ui.Version = string(data)

	// Prepare service(s).
	da, err := dayahead.NewDayAhead(ctx, cfg)
	if err != nil {
		log.Fatalf("Error creating DayAhead service: %v", err)
	}

	// Prepare the router.
	r := router(cfg, da)

	// Prepare the server.
	srv := server(ctx, &cfg.Server, r)

	// Wait signal.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-sigChan:
		log.Printf("Received signal: %v\n", sig)
		cancel()
		// Wait all server's handlers close connections and finish.
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down server: %v", err)
		}
	case <-ctx.Done():
		// Was canceled by server, we don't have to wait it.
		// Can be some jon here (sending buffered logs, etc).
	}
}

func loadConfig() *config.App {
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

	return cfg
}

func router(cfg *config.App, da dayahead.DayAhead) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(10 * time.Second))

	r.Use(appMiddleware.ConfigMiddleware(cfg))
	r.Use(appMiddleware.DayAheadMiddleware(da))

	r.Get("/", controller.IndexHandler)
	r.Get("/api/v1/healthcheck", controller.HealthCheckHandler)
	r.With(appMiddleware.DateMiddleware).With(appMiddleware.VatMiddleware).
		Get("/day-prices/{day}", controller.DayPricesHandler)

	return r
}

func server(ctx context.Context, cfg *config.Server, r chi.Router) *http.Server {
	ctx, cancel := context.WithCancel(ctx)
	// Some default settings for the server.
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		ReadHeaderTimeout: 15 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
		Handler:           r,
	}

	// Static files.
	fs := http.FileServer(http.Dir("static"))
	r.Handle("/favicon.ico", fs) // only ico file for now

	log.Printf("Starting server on :%s\n", cfg.Port)
	go func() {
		err := srv.ListenAndServe()
		log.Println("Server error:", err)
		// We cancel if Server generates fatal in the process.
		cancel()
	}()

	return srv
}
