package dayahead

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/oitimon/day-ahead-prices-notificator/pkg/chart"
	"github.com/oitimon/day-ahead-prices-notificator/pkg/config"
	"github.com/oitimon/day-ahead-prices-notificator/pkg/messenger"
	"github.com/oitimon/day-ahead-prices-notificator/pkg/repository"
	"github.com/shopspring/decimal"
	"strconv"
	"time"
)

const msgDayFormat = "Monday, 02 January 2006"

type Service struct {
	cfg             *config.App
	dataRepository  repository.Data
	chartRepository repository.Bytes
	chart           chart.Chart
	msgr            messenger.Messenger
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

	// Prepare Chart and Messenger.
	s.chart = chart.NewChart(&cfg.Ui)
	if s.msgr, err = messenger.NewMessenger(&cfg.Messenger); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Service) GetHtmlChart(ctx context.Context, day time.Time, opts ...repository.Option) (html []byte, err error) {
	options := repository.NewOptions(opts...)
	prices, err := s.dataRepository.Get(ctx, day, opts...)
	if err != nil {
		return
	}
	if html, err = s.chart.HtmlChart(prices, day); err != nil {
		return
	}

	var before, after string
	dayBefore := day.AddDate(0, 0, -1)
	if s.ValidateDay(dayBefore) == nil {
		before = fmt.Sprintf(`<a href="/day-prices/%s?vat=%v">Before</a>`, dayBefore.Format("2006-01-02"), options.WithVat)
	}
	dayAfter := day.AddDate(0, 0, 1)
	if s.ValidateDay(dayAfter) == nil {
		after = fmt.Sprintf(`<a href="/day-prices/%s?vat=%v">After</a>`, day.AddDate(0, 0, 1).Format("2006-01-02"), options.WithVat)
	}

	html = bytes.Replace(html, []byte("<div class=\"container\">"), []byte(
		fmt.Sprintf(`<div class="container">%s&nbsp;%s</div><div class="container">`,
			before, after),
	), 1)

	return
}

func (s *Service) GetTextChart(ctx context.Context, day time.Time, opts ...repository.Option) (string, error) {
	prices, err := s.dataRepository.Get(ctx, day, opts...)
	if err != nil {
		return "", err
	}
	return s.chart.TextChart(prices, day)
}

func (s *Service) SendMessage(ctx context.Context, day time.Time, opts ...repository.Option) error {
	msg, err := s.GetTextChart(ctx, day, opts...)
	if err != nil {
		return err
	}
	//@todo: use a proper URL from config
	url := fmt.Sprintf("http://my-url/day-prices/%s?vat=%s",
		day.Format("2006-01-02"), strconv.FormatBool(repository.NewOptions(opts...).WithVat))
	msg = fmt.Sprintf("EPEX NL [%s](%s)\n\n%s", day.Format(msgDayFormat), url, msg)
	return s.msgr.SendMessage(ctx, msg)
}

func (s *Service) GetPrices(ctx context.Context, startDate time.Time, opts ...repository.Option) ([]decimal.Decimal, error) {
	return s.dataRepository.Get(ctx, startDate, opts...)
}

func (s *Service) ValidateDay(day time.Time) error {
	if day.Before(s.cfg.Ui.Analytics.MinDate(s.cfg.Location())) {
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
