package dayahead

import (
	"context"
	"github.com/oitimon/day-ahead-prices-notificator/pkg/repository"
	"github.com/shopspring/decimal"
	"time"
)

type DayAhead interface {
	ValidateDay(day time.Time) error
	GetHtmlChart(ctx context.Context, day time.Time, opts ...repository.Option) ([]byte, error)
	GetPrices(ctx context.Context, day time.Time, opts ...repository.Option) ([]decimal.Decimal, error)
}
