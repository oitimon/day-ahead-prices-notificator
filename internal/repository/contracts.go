package repository

import (
	"github.com/shopspring/decimal"
	"time"
)

type Data interface {
	Get(startDate time.Time) ([]decimal.Decimal, error)
}

type Bytes interface {
	Get(startDate time.Time) ([]byte, error)
}
