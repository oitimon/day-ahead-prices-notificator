package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/oitimon/day-ahead-prices-notificator/pkg/config"
	"github.com/oitimon/day-ahead-prices-notificator/pkg/repository/energyzero"
	"github.com/shopspring/decimal"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

const fetchHttpTimeout = 10 * time.Second

type Energyzero struct {
	cfg *config.Energyzero
}

func NewEnergyzero(cfg *config.Energyzero) Data {
	return &Energyzero{
		cfg: cfg,
	}
}

func (e *Energyzero) IsFinal() bool {
	return true
}

func (e *Energyzero) Get(ctx context.Context, startDate time.Time, opts ...Option) (prices []decimal.Decimal, err error) {
	options := NewOptions(opts...)
	endDate := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 23, 59, 59, 0, startDate.Location())
	url := fmt.Sprintf(
		"%s/energyprices?fromDate=%s&tillDate=%s&interval=4&usageType=1&inclBtw=%s", e.cfg.API.Endpoint,
		startDate.In(time.UTC).Format("2006-01-02T15:04:05.000Z"),
		endDate.In(time.UTC).Format("2006-01-02T15:04:05.000Z"), strconv.FormatBool(options.WithVat),
	)

	data := energyzero.PriceData{}
	if err = e.fetchByUrl(ctx, url, &data); err != nil {
		return
	}
	if len(data.Prices) == 0 {
		err = errors.New("no prices available")
		return
	}

	prices = data.PricesDecimal()
	return
}

func (e *Energyzero) fetchByUrl(ctx context.Context, url string, data *energyzero.PriceData) (err error) {
	log.Printf("Fetching prices from %s\n", url)

	ctx, cancel := context.WithTimeout(ctx, fetchHttpTimeout)
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
