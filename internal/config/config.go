package config

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

const timeLocation = "Europe/Amsterdam"
const tomorrowHourMin = 15
const RepositoryDriverGroupCache = "groupcache"
const RepositoryDriverInflux = "influx"
const RepositoryDriverEnergyzero = "energyzero"
const RepositoryDriverStub = "stub"
const MessengerDriverTelegram = "telegram"

type Api struct {
	Endpoint string
}

type GroupCache struct {
	Me     string
	Listen string
	Peers  []string
}

type Influx struct {
	Url     string
	Orgname string
	Bucket  string
	Token   string
}

type Energyzero struct {
	API Api
}

type Repository struct {
	Driver     string
	GroupCache GroupCache
	Influx     Influx
	Energyzero Energyzero
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
				panic(err)
			}
		},
	)
	return cfg.location
}

func (cfg *App) TomorrowHourMin() int {
	return tomorrowHourMin
}

func (cfg *App) SelfCheck() error {
	if err := cfg.Analytics.selfCheck(); err != nil {
		return err
	}
	if err := cfg.DataRepository.selfCheck("DATA"); err != nil {
		return err
	}
	if err := cfg.ChartRepository.selfCheck("CHART"); err != nil {
		return err
	}
	if err := cfg.Server.selfCheck(); err != nil {
		return err
	}
	if err := cfg.Messenger.selfCheck(); err != nil {
		return err
	}

	cfg.Location()
	return nil
}

func (a *Analytics) selfCheck() error {
	if a.HighPrice.IsZero() {
		return errors.New("ANALYTICS_HIGHPRICE not set")
	}
	if a.LowPrice.IsZero() {
		return errors.New("ANALYTICS_LOWPRICE not set")
	}
	return nil
}

func (a *Analytics) MinDate() time.Time {
	return time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
}

func (s *Server) selfCheck() error {
	if s.Port == "" {
		return errors.New("SERVER_PORT not set")
	}
	return nil
}

func (g *GroupCache) selfCheck() error {
	if g.Me == "" {
		return errors.New("GROUPCACHE_ME not set")
	}
	if g.Listen == "" {
		return errors.New("GROUPCACHE_LISTEN not set")
	}
	return nil
}

func (i *Influx) selfCheck() error {
	if i.Url == "" {
		return errors.New("INFLUX_URL not set")
	}
	if i.Orgname == "" {
		return errors.New("INFLUX_ORGNAME not set")
	}
	if i.Bucket == "" {
		return errors.New("INFLUX_BUCKET not set")
	}
	if i.Token == "" {
		return errors.New("INFLUX_TOKEN not set")
	}
	return nil
}

func (e *Energyzero) selfCheck() error {
	return e.API.selfCheck("DATAREPOSITORY_ENERGYZERO")
}

func (a *Api) selfCheck(prefix string) error {
	if a.Endpoint == "" {
		return fmt.Errorf("%s_API_ENDPOINT not set", prefix)
	}
	return nil
}

func (r *Repository) selfCheck(prefix string) error {
	drivers := strings.Split(r.Driver, ",")

	for _, driver := range drivers {
		driver = strings.TrimSpace(driver)
		if driver == RepositoryDriverGroupCache {
			if err := r.GroupCache.selfCheck(); err != nil {
				return err
			}
		} else if driver == RepositoryDriverInflux {
			if err := r.Influx.selfCheck(); err != nil {
				return err
			}
		} else if driver == RepositoryDriverEnergyzero {
			if err := r.Energyzero.selfCheck(); err != nil {
				return err
			}
		} else if driver == RepositoryDriverStub {
			// No check needed for stub
		} else if driver == "" {
			return fmt.Errorf("%sREPOSITORY_DRIVER not set", prefix)
		} else {
			return fmt.Errorf("unknown %sREPOSITORY_DRIVER: %s", prefix, driver)
		}
	}
	return nil
}

func (m *Messenger) selfCheck() error {
	if m.Driver == MessengerDriverTelegram {
		if m.Telegram.Token == "" {
			return errors.New("MESSENGER_TELEGRAM_TOKEN not set")
		}
		if m.Telegram.ChatID == 0 {
			return errors.New("MESSENGER_TELEGRAM_CHATID not set")
		}
	} else if m.Driver == "" {
		return errors.New("MESSENGER_DRIVER not set")
	} else {
		return fmt.Errorf("unknown MESSENGER_DRIVER: %s", m.Driver)
	}
	return nil
}
