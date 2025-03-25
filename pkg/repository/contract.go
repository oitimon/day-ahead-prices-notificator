package repository

import (
	"context"
	"github.com/shopspring/decimal"
	"time"
)

type Data interface {
	Get(ctx context.Context, startDate time.Time, opts ...Option) ([]decimal.Decimal, error)
	IsFinal() bool
}

type Bytes interface {
	Get(ctx context.Context, startDate time.Time, opts ...Option) ([]byte, error)
	IsFinal() bool
}
