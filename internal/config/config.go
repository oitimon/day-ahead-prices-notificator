package config

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

const timeLocation = "Europe/Amsterdam"
const tomorrowHourMin = 15
const LoaderDriverStub = "stub"
const LoaderDriverEnergyZero = "energyzero"
const RepositoryDriverGroupCache = "groupCache"
const MessengerDriverTelegram = "telegram"

type Api struct {
	Endpoint string
}

type GroupCache struct {
	Me     string
	Listen string
	Peers  []string
}

type Repository struct {
	Driver     string
	GroupCache GroupCache
}

type Loader struct {
	InclBtw bool
	Driver  string
	API     Api
}

type Server struct {
	Port string
}

type Telegram struct {
	Token  string
	ChatID int64
}

type Messenger struct {
	Driver   string
	Telegram Telegram
}

type Analytics struct {
	HighPrice decimal.Decimal
	LowPrice  decimal.Decimal
	Version   string
}

// App Config struct to hold environment variables
type App struct {
	Analytics       Analytics
	DataRepository  Repository
	ChartRepository Repository
	Loader          Loader
	Server          Server
	Messenger       Messenger

	locationOnce sync.Once
	location     *time.Location
}

func (cfg *App) Location() *time.Location {
	cfg.locationOnce.Do(
		func() {
			var err error
			cfg.location, err = time.LoadLocation(timeLocation)
			if err != nil {
				log.Fatal(err)
			}
		},
	)
	return cfg.location
}

func (cfg *App) TomorrowHourMin() int {
	return tomorrowHourMin
}

func (cfg *App) SelfCheck() error {

	if cfg.Analytics.HighPrice.IsZero() {
		return errors.New("ANALYTICS_HIGHPRICE not set")
	}
	if cfg.Analytics.LowPrice.IsZero() {
		return errors.New("ANALYTICS_LOWPRICE not set")
	}

	if cfg.Loader.Driver == LoaderDriverEnergyZero {
		if cfg.Loader.API.Endpoint == "" {
			return errors.New("LOADER_API_ENDPOINT not set")
		}
	} else if cfg.Loader.Driver == LoaderDriverStub {
		// nothing to check
	} else if cfg.Loader.Driver == "" {
		return errors.New("LOADER_DRIVER not set")
	} else {
		return fmt.Errorf("unknown LOADER_DRIVER: %s", cfg.Loader.Driver)
	}

	if cfg.Server.Port == "" {
		return errors.New("SERVER_PORT not set")
	}

	if cfg.Messenger.Driver == MessengerDriverTelegram {
		if cfg.Messenger.Telegram.Token == "" {
			return errors.New("MESSENGER_TELEGRAM_TOKEN not set")
		}
		if cfg.Messenger.Telegram.ChatID == 0 {
			return errors.New("MESSENGER_TELEGRAM_CHATID not set")
		}
	} else if cfg.Messenger.Driver == "" {
		return errors.New("MESSENGER_DRIVER not set")
	} else {
		return fmt.Errorf("unknown MESSENGER_DRIVER: %s", cfg.Messenger.Driver)
	}

	cfg.Location()
	return nil
}

func (a *Analytics) MinDate() time.Time {
	return time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
}
