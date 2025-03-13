package app

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
const loaderDriverStub = "stub"
const loaderDriverEnergyZero = "energyzero"
const messengerDriverTelegram = "telegram"

type ConfigAPI struct {
	Endpoint string
}

type ConfigLoader struct {
	InclBtw bool
	Driver  string
	API     ConfigAPI
}

type ConfigServer struct {
	Port string
}

type ConfigTelegram struct {
	Token  string
	ChatID int64
}

type ConfigMessenger struct {
	Driver   string
	Telegram ConfigTelegram
}

type ConfigAnalytics struct {
	HighPrice decimal.Decimal
	LowPrice  decimal.Decimal
	Version   string
}

// Config struct to hold environment variables
type ConfigApp struct {
	Analytics ConfigAnalytics
	Loader    ConfigLoader
	Server    ConfigServer
	Messenger ConfigMessenger

	locationOnce sync.Once
	location     *time.Location
}

func (cfg *ConfigApp) Location() *time.Location {
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

func (cfg *ConfigApp) TomorrowHourMin() int {
	return tomorrowHourMin
}

func (cfg *ConfigApp) SelfCheck() error {

	if cfg.Analytics.HighPrice.IsZero() {
		return errors.New("ANALYTICS_HIGHPRICE not set")
	}
	if cfg.Analytics.LowPrice.IsZero() {
		return errors.New("ANALYTICS_LOWPRICE not set")
	}

	if cfg.Loader.Driver == loaderDriverEnergyZero {
		if cfg.Loader.API.Endpoint == "" {
			return errors.New("LOADER_API_ENDPOINT not set")
		}
	} else if cfg.Loader.Driver == loaderDriverStub {
		// nothing to check
	} else if cfg.Loader.Driver == "" {
		return errors.New("LOADER_DRIVER not set")
	} else {
		return fmt.Errorf("unknown LOADER_DRIVER: %s", cfg.Loader.Driver)
	}

	if cfg.Server.Port == "" {
		return errors.New("SERVER_PORT not set")
	}

	if cfg.Messenger.Driver == messengerDriverTelegram {
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
