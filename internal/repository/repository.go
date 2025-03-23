package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/oitimon/day-ahead-prices-notificator/internal/config"
	"strings"
)

func NewDataRepository(ctx context.Context, cfg *config.Repository) (dr Data, err error) {
	drivers := strings.Split(cfg.Driver, ",")

	for i := len(drivers) - 1; i >= 0; i-- {
		driver := strings.TrimSpace(drivers[i])
		var repo Data
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
		if dr == nil && !repo.IsFinal() {
			err = fmt.Errorf("driver %s is not final", driver)
		} else if dr != nil && repo.IsFinal() {
			err = fmt.Errorf("driver %s is final", driver)
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
