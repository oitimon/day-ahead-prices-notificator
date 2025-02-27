package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/chromedp/chromedp"
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
const htmlChartWidth = 820
const htmlChartHeight = 520
const timeLocation = "Europe/Amsterdam"

// Config struct to hold environment variables
type Config struct {
	APIEndpoint    string          `envconfig:"API_ENDPOINT"`
	InclBtw        string          `envconfig:"INCL_BTW"`
	TelegramToken  string          `envconfig:"TELEGRAM_TOKEN"`
	TelegramChatID int64           `envconfig:"TELEGRAM_CHAT_ID"`
	HighPrice      decimal.Decimal `envconfig:"HIGH_PRICE"`
	LowPrice       decimal.Decimal `envconfig:"LOW_PRICE"`
}

// PriceData struct to parse the response JSON
type PriceData struct {
	sync.Once

	Prices []struct {
		Price       float64 `json:"price"`
		ReadingDate string  `json:"readingDate"`
	} `json:"Prices"`
	pricesFloat64 []float64
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
	var cfg Config
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	if err = envconfig.Process("", &cfg); err != nil {
		log.Fatalf("Error processing environment variables: %v", err)
	}
	if cfg.TelegramToken == "" || cfg.TelegramChatID == 0 {
		log.Fatal("TELEGRAM_TOKEN or TELEGRAM_CHAT_ID not set")
	}

	// Get start and end dates for tomorrow in Amsterdam time
	location, _ := time.LoadLocation(timeLocation)
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, location)
	endDate := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 23, 59, 59, 0, location)

	// Create URL with dynamic dates and InclBtw parameter
	url := fmt.Sprintf(
		"%s/energyprices?fromDate=%s&tillDate=%s&interval=4&usageType=1&inclBtw=%s", cfg.APIEndpoint,
		startDate.In(time.UTC).Format("2006-01-02T15:04:05.000Z"),
		endDate.In(time.UTC).Format("2006-01-02T15:04:05.000Z"), cfg.InclBtw,
	)

	// Step 1: Download JSON
	priceData, err := fetchPrices(url)
	if err != nil {
		sendTelegramMessage(cfg, "Error generating "+startDate.Format("2006-01-02"))
		log.Fatal(err)
		return
	}

	// Step 2: Create chart
	r, err := createChart(&priceData, startDate)
	if err != nil {
		sendTelegramMessage(cfg, "Error generating "+startDate.Format("2006-01-02"))
		log.Fatal(err)
		return
	}

	// Step 3: Send Telegram message
	if err = sendPriceMessage(cfg, &priceData, startDate, r); err != nil {
		sendTelegramMessage(cfg, "Error generating "+startDate.Format("2006-01-02"))
		log.Fatal(err)
		return
	}
}

// fetchPrices function downloads and parses the price JSON
func fetchPrices(url string) (data PriceData, err error) {
	log.Printf("Fetching prices from %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = errors.New("Failed to fetch data from API, status code: " + resp.Status)
		return
	}

	body, _ := io.ReadAll(resp.Body)
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
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()
	var buf []byte
	if err = chromedp.Run(
		ctx,
		chromedp.Navigate("file://"+chartHtmlFilename),
		chromedp.WaitVisible("body"),
		chromedp.Sleep(1),
		chromedp.EmulateViewport(htmlChartWidth, htmlChartHeight),
		chromedp.CaptureScreenshot(&buf),
	); err != nil {
		return
	}

	r = bytes.NewReader(buf)
	return
}

// sendPriceMessage sends a price chart to the Telegram bot with a message
func sendPriceMessage(cfg Config, priceData *PriceData, day time.Time, r io.Reader) (err error) {
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
func sendTelegramMessage(cfg Config, message string) {
	log.Printf("Sending message to Telegram: %s\n", message)
	client, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		log.Fatal("Error Telegram Bot creating:", err)
	}
	msg := tgbotapi.NewMessage(cfg.TelegramChatID, message)
	if _, err = client.Send(msg); err != nil {
		log.Printf("Error sending message to Telegram: %v\n", err)
	}
}
