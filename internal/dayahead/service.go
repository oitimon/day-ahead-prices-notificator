package dayahead

import (
	"context"
	"errors"
	"fmt"
	"github.com/oitimon/day-ahead-prices-notificator/internal/chart"
	"github.com/oitimon/day-ahead-prices-notificator/internal/config"
	"github.com/oitimon/day-ahead-prices-notificator/internal/repository"
	"github.com/shopspring/decimal"
	"time"
)

type Service struct {
	cfg             *config.App
	dataRepository  repository.Data
	chartRepository repository.Bytes
	chart           chart.Chart
}

func NewDayAhead(ctx context.Context, cfg *config.App) (DayAhead, error) {
	s := &Service{
		cfg: cfg,
	}
	var err error

	// Prepare Loader and Repositories.
	s.dataRepository, err = repository.NewDataRepository(ctx, &cfg.DataRepository)
	if err != nil {
		return nil, fmt.Errorf("Error creating data repository: %v", err)
	}
	s.chartRepository, err = repository.NewBytesRepository(ctx, &cfg.ChartRepository)
	if err != nil {
		return nil, fmt.Errorf("Error creating bytes repository: %v", err)
	}

	// Prepare Chart.
	s.chart = chart.NewChart(&cfg.Analytics)

	return s, nil
}

func (s *Service) GetHtmlChart(ctx context.Context, day time.Time) (html []byte, err error) {
	prices, err := s.dataRepository.Get(ctx, day)
	if err != nil {
		return
	}
	return s.chart.HtmlChart(prices, day)
}

func (s *Service) GetPrices(ctx context.Context, startDate time.Time) ([]decimal.Decimal, error) {
	return s.dataRepository.Get(ctx, startDate)
}

func (s *Service) ValidateDay(day time.Time) error {
	if day.Before(s.cfg.Analytics.MinDate()) {
		return errors.New("day is too old")
	}

	now := time.Now().In(s.cfg.Location())
	tomorrow := now.AddDate(0, 0, 1)
	tomorrow = time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, s.cfg.Location())
	if day.After(tomorrow) {
		return errors.New("day is in the future after tomorrow")
	} else if day.Equal(tomorrow) && now.Hour() < s.cfg.TomorrowHourMin() {
		return errors.New("day is tomorrow but it's too early")
	}
	return nil
}
