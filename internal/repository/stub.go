package repository

import (
	"context"
	"github.com/shopspring/decimal"
	"time"
)

type Stub struct {
}

func NewStub() Data {
	return &Stub{}
}

func (*Stub) IsFinal() bool {
	return true
}

func (*Stub) Get(_ context.Context, startDate time.Time) ([]decimal.Decimal, error) {
	return []decimal.Decimal{
		decimal.NewFromFloat(0.15),
		decimal.NewFromFloat(0.13),
		decimal.NewFromFloat(0.12),
		decimal.NewFromFloat(0.11),
		decimal.NewFromFloat(0.11),
		decimal.NewFromFloat(0.11),
		decimal.NewFromFloat(0.12),
		decimal.NewFromFloat(0.12),
		decimal.NewFromFloat(0.12),
		decimal.NewFromFloat(0.11),
		decimal.NewFromFloat(0.08),
		decimal.NewFromFloat(0.06),
		decimal.NewFromFloat(0.04),
		decimal.NewFromFloat(0),
		decimal.NewFromFloat(-0.02),
		decimal.NewFromFloat(0.06),
		decimal.NewFromFloat(0.1),
		decimal.NewFromFloat(0.15),
		decimal.NewFromFloat(0.17),
		decimal.NewFromFloat(0.18),
		decimal.NewFromFloat(0.16),
		decimal.NewFromFloat(0.15),
		decimal.NewFromFloat(0.15),
		decimal.NewFromFloat(0.13),
	}, nil
}
