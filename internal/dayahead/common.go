package dayahead

import (
	"errors"
	"fmt"
	"github.com/oitimon/day-ahead-prices-notificator/internal/chart"
	"github.com/oitimon/day-ahead-prices-notificator/internal/config"
	"github.com/oitimon/day-ahead-prices-notificator/internal/loader"
	"github.com/oitimon/day-ahead-prices-notificator/internal/repository"
	"github.com/shopspring/decimal"
	"time"
)

type Common struct {
	cfg             *config.App
	dataRepository  repository.Data
	bytesRepository repository.Bytes
	chart           chart.Chart
}

func NewDayAhead(cfg *config.App) (DayAhead, error) {
	da := &Common{
		cfg: cfg,
	}
	var err error

	// Prepare Loader and Repositories.
	ldr := loader.NewLoader(&cfg.Loader)
	da.dataRepository, err = repository.NewDataRepository(&cfg.DataRepository, ldr)
	if err != nil {
		err = fmt.Errorf("Error creating data repository: %v", err)
	}
	da.bytesRepository, err = repository.NewBytesRepository(&cfg.DataRepository)
	if err != nil {
		err = fmt.Errorf("Error creating bytes repository: %v", err)
	}

	// Prepare Chart.
	da.chart = chart.NewChart(&cfg.Analytics)

	return da, err
}

func (da *Common) GetHtmlChart(day time.Time) (html []byte, err error) {
	prices, err := da.dataRepository.Get(day)
	if err != nil {
		return
	}
	return da.chart.HtmlChart(prices, day)
}

func (da *Common) GetPrices(startDate time.Time) ([]decimal.Decimal, error) {
	return da.dataRepository.Get(startDate)
}

func (da *Common) ValidateDay(day time.Time) error {
	if day.Before(da.cfg.Analytics.MinDate()) {
		return errors.New("day is too old")
	}

	tomorrow := time.Now().AddDate(0, 0, 1)
	tomorrow = time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, da.cfg.Location())
	if day.After(tomorrow) {
		return errors.New("day is in the future after tomorrow")
	} else if day.Equal(tomorrow) && time.Now().In(da.cfg.Location()).Hour() < da.cfg.TomorrowHourMin() {
		return errors.New("day is tomorrow but it's too early")
	}
	return nil
}
