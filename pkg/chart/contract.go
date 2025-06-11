package chart

import (
	"github.com/shopspring/decimal"
	"time"
)

type Chart interface {
	HtmlChart(prices []decimal.Decimal, day time.Time, format int) ([]byte, error)
	TextChart(prices []decimal.Decimal, day time.Time) (string, error)
}
