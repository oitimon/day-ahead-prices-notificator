package chart

import (
	"github.com/shopspring/decimal"
	"time"
)

type Chart interface {
	HtmlChart(prices []decimal.Decimal, day time.Time) ([]byte, error)
	TextChart(prices []decimal.Decimal, day time.Time) (string, error)
}
