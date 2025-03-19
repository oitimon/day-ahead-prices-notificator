package repository

import (
	"errors"
	"github.com/oitimon/day-ahead-prices-notificator/internal/config"
	"github.com/oitimon/day-ahead-prices-notificator/internal/loader"
)

func NewDataRepository(cfg *config.Repository, ldr loader.Loader) (dr Data, err error) {
	switch cfg.Driver {
	case config.RepositoryDriverGroupCache:
		dr = NewGroupCache(&cfg.GroupCache, ldr).Data()
	default:
		err = errors.New("unknown data repository driver")
	}
	return
}

func NewBytesRepository(cfg *config.Repository) (br Bytes, err error) {
	switch cfg.Driver {
	case config.RepositoryDriverGroupCache:
		br = NewGroupCache(&cfg.GroupCache, nil).Bytes()
	default:
		err = errors.New("unknown bytes repository driver")
	}
	return
}
