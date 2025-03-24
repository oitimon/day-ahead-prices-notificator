package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/oitimon/day-ahead-prices-notificator/internal/config"
	"strings"
)

type Options struct {
	WithVat bool
}

type Option func(*Options)

func NewDataRepository(ctx context.Context, cfg *config.Repository) (dr Data, err error) {
	drivers := strings.Split(cfg.Driver, ",")

	for i := len(drivers) - 1; i >= 0; i-- {
		driver := strings.TrimSpace(drivers[i])
		var repo Data
		if repo, err = func() (repo Data, err error) {
			switch driver {
			case config.RepositoryDriverGroupCache:
				repo = NewGroupCache(&cfg.GroupCache, dr).Data()
			case config.RepositoryDriverInflux:
				if repo, err = NewInflux(ctx, &cfg.Influx, dr); err != nil {
					return
				}
			case config.RepositoryDriverEnergyzero:
				repo = NewEnergyzero(&cfg.Energyzero)
			case config.RepositoryDriverStub:
				repo = NewStub()
			default:
				err = errors.New("unknown data repository driver")
				return
			}
			return
		}(); err != nil {
			return
		}

		if dr == nil && !repo.IsFinal() {
			err = fmt.Errorf("driver %s cannot be started because it is not final", driver)
			return
		} else if dr != nil && repo.IsFinal() {
			err = fmt.Errorf("driver %s cannot be chained because it is final", driver)
			return
		}

		dr = repo
	}

	return
}

func NewBytesRepository(ctx context.Context, cfg *config.Repository) (br Bytes, err error) {
	switch cfg.Driver {
	case config.RepositoryDriverGroupCache:
		br = NewGroupCache(&cfg.GroupCache, nil).Bytes()
	default:
		err = errors.New("unknown bytes repository driver")
	}
	return
}

func WithVat(includingVat bool) Option {
	return func(o *Options) {
		o.WithVat = includingVat
	}
}

func NewOptions(opts ...Option) (o *Options) {
	o = &Options{}
	for _, opt := range opts {
		opt(o)
	}
	return
}
