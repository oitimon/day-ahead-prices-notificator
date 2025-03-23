package repository

import (
	"context"
	"github.com/shopspring/decimal"
	"time"
)

type Data interface {
	Get(ctx context.Context, startDate time.Time) ([]decimal.Decimal, error)
	IsFinal() bool
}

type Bytes interface {
	Get(ctx context.Context, startDate time.Time) ([]byte, error)
	IsFinal() bool
}
