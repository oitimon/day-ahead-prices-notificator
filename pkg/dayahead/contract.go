package dayahead

import (
	"context"
	"github.com/oitimon/day-ahead-prices-notificator/pkg/repository"
	"github.com/shopspring/decimal"
	"time"
)

type DayAhead interface {
	GetHtmlChart(ctx context.Context, day time.Time, opts ...repository.Option) ([]byte, error)
	GetHaIframeChart(ctx context.Context, day time.Time, opts ...repository.Option) ([]byte, error)
	GetTextChart(ctx context.Context, day time.Time, opts ...repository.Option) (string, error)
	SendMessage(ctx context.Context, day time.Time, opts ...repository.Option) error
	GetPrices(ctx context.Context, day time.Time, opts ...repository.Option) ([]decimal.Decimal, error)
	ValidateDay(day time.Time) error
}
