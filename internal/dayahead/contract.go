package dayahead

import (
	"github.com/shopspring/decimal"
	"time"
)

type DayAhead interface {
	ValidateDay(day time.Time) error
	GetHtmlChart(day time.Time) ([]byte, error)
	GetPrices(day time.Time) ([]decimal.Decimal, error)
}
