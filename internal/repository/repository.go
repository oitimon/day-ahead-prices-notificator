package repository

import (
	"context"
	"errors"
	"github.com/oitimon/day-ahead-prices-notificator/internal/config"
	"github.com/oitimon/day-ahead-prices-notificator/internal/loader"
)

func NewDataRepository(ctx context.Context, cfg *config.Repository, ldr loader.Loader) (dr Data, err error) {
	switch cfg.Driver {
	case config.RepositoryDriverGroupCache:
		dr = NewGroupCache(&cfg.GroupCache, ldr).Data()
	case config.RepositoryDriverInflux:
		dr = NewInflux(ctx, &cfg.Influx, ldr).Data()
	default:
		err = errors.New("unknown data repository driver")
	}
	return
}

func NewBytesRepository(ctx context.Context, cfg *config.Repository) (br Bytes, err error) {
	switch cfg.Driver {
	case config.RepositoryDriverGroupCache:
		br = NewGroupCache(&cfg.GroupCache, nil).Bytes()
	case config.RepositoryDriverInflux:
		br = NewInflux(ctx, &cfg.Influx, nil).Bytes()
	default:
		err = errors.New("unknown bytes repository driver")
	}
	return
}
