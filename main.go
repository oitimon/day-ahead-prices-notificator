package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	//"github.com/chromedp/chromedp"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/shopspring/decimal"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const chartHtmlFilename = "/tmp/epex_nl_da_prices_chart.html"
const chartPngFilename = "/tmp/epex_nl_da_prices_chart.png"

// const htmlChartWidth = 820
// const htmlChartHeight = 520
const timeLocation = "Europe/Amsterdam"
const fetchTimeout = 10 * time.Second

// Config struct to hold environment variables
type Config struct {
	APIEndpoint    string          `envconfig:"API_ENDPOINT"`
	InclBtw        string          `envconfig:"INCL_BTW"`
	TelegramToken  string          `envconfig:"TELEGRAM_TOKEN"`
	TelegramChatID int64           `envconfig:"TELEGRAM_CHAT_ID"`
	HighPrice      decimal.Decimal `envconfig:"HIGH_PRICE"`
	LowPrice       decimal.Decimal `envconfig:"LOW_PRICE"`

	WebPort string `envconfig:"WEB_PORT"`

	locationOnce sync.Once
	location     *time.Location `envconfig:"-"`
}

func (cfg *Config) Location() *time.Location {
	cfg.locationOnce.Do(
		func() {
			var err error
			cfg.location, err = time.LoadLocation(timeLocation)
			if err != nil {
				log.Fatal(err)
			}
		},
	)
	return cfg.location
}

func (cfg *Config) SelfCheck() {
	if cfg.APIEndpoint == "" {
		log.Fatal("API_ENDPOINT not set")
	}
	if cfg.InclBtw == "" {
		log.Fatal("INCL_BTW not set")
	}
	if cfg.TelegramToken == "" {
		log.Fatal("TELEGRAM_TOKEN not set")
	}
	if cfg.TelegramChatID == 0 {
		log.Fatal("TELEGRAM_CHAT_ID not set")
	}
	if cfg.HighPrice.IsZero() {
		log.Fatal("HIGH_PRICE not set")
	}
	if cfg.LowPrice.IsZero() {
		log.Fatal("LOW_PRICE not set")
	}
	if cfg.WebPort == "" {
		log.Fatal("WEB_PORT not set")
	}
	cfg.Location()
}

// PriceData struct to parse the response JSON
type PriceData struct {
	sync.Once

	Prices        []PriceDataEntry `json:"Prices"`
	pricesFloat64 []float64
}

type PriceDataEntry struct {
	Price       float64 `json:"price"`
	ReadingDate string  `json:"readingDate"`
}

func (data *PriceData) PricesFloat64() []float64 {
	data.Do(
		func() {
			data.pricesFloat64 = make([]float64, len(data.Prices))
			for i, priceEntry := range data.Prices {
				data.pricesFloat64[i] = priceEntry.Price
			}
		},
	)
	return data.pricesFloat64
}

func main() {
	cfg := &Config{}
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
	cfg.SelfCheck()

	// Get start and end dates for tomorrow in Amsterdam time
	//now := time.Now()
	//startDate := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, cfg.Location())
	//
	//fetchAndSendPrices(cfg, startDate)

	// Start the server.
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Get("/", homeHandler)
	r.Get("/api/v1/healthcheck", healthCheckHandler)
	log.Printf("Starting server on :%s\n", cfg.WebPort)
	if err := http.ListenAndServe(":"+cfg.WebPort, r); err != nil {
		log.Fatal(err)
	}
}

func fetchAndSendPrices(cfg *Config, startDate time.Time) {
	endDate := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 23, 59, 59, 0, cfg.Location())

	// Create URL with dynamic dates and InclBtw parameter
	url := fmt.Sprintf(
		"%s/energyprices?fromDate=%s&tillDate=%s&interval=4&usageType=1&inclBtw=%s", cfg.APIEndpoint,
		startDate.In(time.UTC).Format("2006-01-02T15:04:05.000Z"),
		endDate.In(time.UTC).Format("2006-01-02T15:04:05.000Z"), cfg.InclBtw,
	)

	// Step 1: Download JSON
	priceData, err := fetchPrices(url)
	if err != nil {
		//sendTelegramMessage(cfg, "Error generating "+startDate.Format("2006-01-02"))
		log.Fatal(err)
		return
	}

	// Step 2: Create chart
	r, err := createChart(&priceData, startDate)
	if err != nil {
		//sendTelegramMessage(cfg, "Error generating "+startDate.Format("2006-01-02"))
		log.Fatal(err)
		return
	}

	// Step 3: Send Telegram message
	if err = sendPriceMessage(cfg, &priceData, startDate, r); err != nil {
		//sendTelegramMessage(cfg, "Error generating "+startDate.Format("2006-01-02"))
		log.Fatal(err)
		return
	}
}

// fetchPrices function downloads and parses the price JSON
func fetchPrices(url string) (data PriceData, err error) {
	log.Printf("Fetching prices from %s\n", url)

	ctx, cancel := context.WithTimeout(context.Background(), fetchTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		err = fmt.Errorf("failed to create request: %w", err)
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		err = fmt.Errorf("failed to fetch data from API: %w", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = errors.New("Failed to fetch data from API, status code: " + resp.Status)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("failed to read response body: %w", err)
		return
	}
	if err = json.Unmarshal(body, &data); err != nil {
		return
	}

	if len(data.Prices) == 0 {
		err = errors.New("no prices available")
		return
	}

	return
}

// createChart function generates a bar chart of prices
func createChart(priceData *PriceData, day time.Time) (r io.Reader, err error) {
	// Saving the chart image as HTML file
	log.Printf("Generating charts for: %s\n", day.Format("2006-01-02"))
	prices := priceData.PricesFloat64()

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

	// Saving the chart image as PNG file
	//ctx, cancel := chromedp.NewContext(context.Background())
	//defer cancel()
	//var buf []byte
	//if err = chromedp.Run(
	//	ctx,
	//	chromedp.Navigate("file://"+chartHtmlFilename),
	//	chromedp.WaitVisible("body"),
	//	chromedp.Sleep(1),
	//	chromedp.EmulateViewport(htmlChartWidth, htmlChartHeight),
	//	chromedp.CaptureScreenshot(&buf),
	//); err != nil {
	//	return
	//}
	var buf []byte
	buf = []byte("PNG test")

	r = bytes.NewReader(buf)
	return
}

// sendPriceMessage sends a price chart to the Telegram bot with a message
func sendPriceMessage(cfg *Config, priceData *PriceData, day time.Time, r io.Reader) (err error) {
	message := fmt.Sprintf("EPEX NL Day-Ahead %s", day.Format("2006-01-02"))
	var priceMessage []string

	// Check for high/low prices
	highDetected := false
	lowDetected := false

	for _, priceEntry := range priceData.Prices {
		price := decimal.NewFromFloat(priceEntry.Price)
		if price.GreaterThanOrEqual(cfg.HighPrice) {
			highDetected = true
		}
		if price.LessThanOrEqual(cfg.LowPrice) {
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
	client, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		log.Fatal("Error Telegram Bot creating:", err)
	}

	message += "\n" + strings.Join(priceMessage, "\n")
	log.Printf("Sending messages to Telegram: %s\n", strings.Replace(message, "\n", " ", -1))

	if _, err = client.Send(tgbotapi.NewMessage(cfg.TelegramChatID, message)); err != nil {
		return
	}

	_, err = client.Send(
		tgbotapi.NewPhoto(
			cfg.TelegramChatID, tgbotapi.FileReader{
				Name:   chartPngFilename,
				Reader: r,
			},
		),
	)

	return err
}

// sendTelegramMessage sends a general message to the Telegram bot
func sendTelegramMessage(cfg *Config, message string) {
	log.Printf("Sending message to Telegram: %s\n", message)
	client, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		log.Printf("Failed to initialize Telegram bot: %v\n", err)
		return
	}
	msg := tgbotapi.NewMessage(cfg.TelegramChatID, message)
	if _, err = client.Send(msg); err != nil {
		log.Printf("Error sending message to Telegram: %v\n", err)
		return
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("Welcome to DA price notificator!"))
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("healthy"))
}
