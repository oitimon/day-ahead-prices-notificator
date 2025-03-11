package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/oitimon/day-ahead-prices-notificator/pkg/models"
	"github.com/shopspring/decimal"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

const fetchHttpTimeout = 10 * time.Second

// fetchPrices function downloads and parses the prices from the driver
func FetchPrices(cfg *ConfigLoader, startDate time.Time) ([]decimal.Decimal, error) {
	switch cfg.Driver {
	case loaderDriverEnergyZero:
		return fetchAsEnergyZero(cfg, startDate)
	case loaderDriverStub:
		return generateStub()
	default:
		return nil, errors.New("unknown loader driver")
	}
}

func fetchAsEnergyZero(cfg *ConfigLoader, startDate time.Time) (res []decimal.Decimal, err error) {
	endDate := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 23, 59, 59, 0, startDate.Location())
	url := fmt.Sprintf(
		"%s/energyprices?fromDate=%s&tillDate=%s&interval=4&usageType=1&inclBtw=%s", cfg.API.Endpoint,
		startDate.In(time.UTC).Format("2006-01-02T15:04:05.000Z"),
		endDate.In(time.UTC).Format("2006-01-02T15:04:05.000Z"), strconv.FormatBool(cfg.InclBtw),
	)

	data := models.PriceData{}
	if err = fetchByUrl(url, &data); err != nil {
		return
	}
	if len(data.Prices) == 0 {
		err = errors.New("no prices available")
		return
	}

	res = data.PricesDecimal()
	return
}

func fetchByUrl(url string, data *models.PriceData) (err error) {
	log.Printf("Fetching prices from %s\n", url)

	ctx, cancel := context.WithTimeout(context.Background(), fetchHttpTimeout)
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

func generateStub() ([]decimal.Decimal, error) {
	return []decimal.Decimal{
		decimal.NewFromFloat(0.15),
		decimal.NewFromFloat(0.13),
		decimal.NewFromFloat(0.12),
		decimal.NewFromFloat(0.11),
		decimal.NewFromFloat(0.11),
		decimal.NewFromFloat(0.11),
		decimal.NewFromFloat(0.12),
		decimal.NewFromFloat(0.12),
		decimal.NewFromFloat(0.12),
		decimal.NewFromFloat(0.11),
		decimal.NewFromFloat(0.08),
		decimal.NewFromFloat(0.06),
		decimal.NewFromFloat(0.04),
		decimal.NewFromFloat(0),
		decimal.NewFromFloat(0),
		decimal.NewFromFloat(0.06),
		decimal.NewFromFloat(0.1),
		decimal.NewFromFloat(0.15),
		decimal.NewFromFloat(0.17),
		decimal.NewFromFloat(0.18),
		decimal.NewFromFloat(0.16),
		decimal.NewFromFloat(0.15),
		decimal.NewFromFloat(0.15),
		decimal.NewFromFloat(0.13),
	}, nil
}
