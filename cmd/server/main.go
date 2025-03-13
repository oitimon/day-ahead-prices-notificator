package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/oitimon/day-ahead-prices-notificator/internal/app"
	"github.com/oitimon/day-ahead-prices-notificator/internal/controller"
	appMiddleware "github.com/oitimon/day-ahead-prices-notificator/internal/middleware"
	"github.com/shopspring/decimal"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const chartHtmlFilename = "/tmp/epex_nl_da_prices_chart.html"

func main() {
	cfg := &app.ConfigApp{}
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

	// Start the server.
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(appMiddleware.ConfigMiddleware(cfg))
	r.Use(middleware.Timeout(30 * time.Second))
	r.Get("/", controller.IndexHandler)
	r.Get("/api/v1/healthcheck", controller.HealthCheckHandler)
	r.With(appMiddleware.DateMiddleware).Get("/day-prices/{year}-{month}-{day}", controller.DayPricesHandler)
	log.Printf("Starting server on :%s\n", cfg.Server.Port)
	if err := http.ListenAndServe(":"+cfg.Server.Port, r); err != nil {
		log.Fatal(err)
	}
}

// createChart function generates a bar chart of prices
func createChart(prices []decimal.Decimal, day time.Time) (err error) {
	// Saving the chart image as HTML file
	log.Printf("Generating charts for: %s\n", day.Format("2006-01-02"))

	bar := charts.NewBar()
	xAxis := make([]string, len(prices))
	for i := 0; i < len(prices); i++ {
		xAxis[i] = fmt.Sprintf("%02d:00", i)
	}
	yAxis := make([]opts.BarData, len(prices))
	for i, price := range prices {
		yAxis[i] = opts.BarData{Value: price}
	}

	bar.SetXAxis(xAxis).
		AddSeries(
			"", yAxis,
			charts.WithAnimationOpts(opts.Animation{Animation: opts.Bool(false)}),
		).
		SetGlobalOptions(
			charts.WithTitleOpts(opts.Title{Title: fmt.Sprintf("EPEX NL %s", day.Format("2006-01-02")), Left: "36%"}),
			charts.WithXAxisOpts(
				opts.XAxis{
					AxisLabel: &opts.AxisLabel{
						Rotate: 90,
					},
				},
			),
		).
		SetSeriesOptions(
			charts.WithLabelOpts(
				opts.Label{
					Show:     opts.Bool(true),
					Position: "inside",
				},
			),
		)

	f, err := os.Create(chartHtmlFilename)
	if err != nil {
		return
	}
	defer f.Close()
	if err = bar.Render(f); err != nil {
		return
	}

	return
}

// sendPriceMessage sends a price chart to the Telegram bot with a message
func sendPriceMessage(cfg *app.ConfigApp, prices []decimal.Decimal, day time.Time) (err error) {
	message := fmt.Sprintf("EPEX NL Day-Ahead %s", day.Format("2006-01-02"))
	var priceMessage []string

	// Check for high/low prices
	highDetected := false
	lowDetected := false

	for _, priceEntry := range prices {
		price := decimal.NewFromFloat(priceEntry.InexactFloat64())
		if price.GreaterThanOrEqual(cfg.Analytics.HighPrice) {
			highDetected = true
		}
		if price.LessThanOrEqual(cfg.Analytics.LowPrice) {
			lowDetected = true
		}
	}

	if highDetected && lowDetected {
		priceMessage = append(priceMessage, "There are High/Low prices")
	} else if highDetected {
		priceMessage = append(priceMessage, "There are High prices")
	} else if lowDetected {
		priceMessage = append(priceMessage, "There are Low prices")
	}

	// Send Telegram message with chart
	client, err := tgbotapi.NewBotAPI(cfg.Messenger.Telegram.Token)
	if err != nil {
		log.Fatal("Error Telegram Bot creating:", err)
	}

	message += "\n" + strings.Join(priceMessage, "\n")
	log.Printf("Sending messages to Telegram: %s\n", strings.Replace(message, "\n", " ", -1))

	if _, err = client.Send(tgbotapi.NewMessage(cfg.Messenger.Telegram.ChatID, message)); err != nil {
		return
	}

	return err
}

// sendTelegramMessage sends a general message to the Telegram bot
func sendTelegramMessage(cfg *app.ConfigTelegram, message string) {
	log.Printf("Sending message to Telegram: %s\n", message)
	client, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		log.Printf("Failed to initialize Telegram bot: %v\n", err)
		return
	}
	msg := tgbotapi.NewMessage(cfg.ChatID, message)
	if _, err = client.Send(msg); err != nil {
		log.Printf("Error sending message to Telegram: %v\n", err)
		return
	}
}
