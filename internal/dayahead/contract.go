package dayahead

import (
	"context"
	"github.com/shopspring/decimal"
	"time"
)

type DayAhead interface {
	ValidateDay(day time.Time) error
	GetHtmlChart(ctx context.Context, day time.Time) ([]byte, error)
	GetPrices(ctx context.Context, day time.Time) ([]decimal.Decimal, error)
}
