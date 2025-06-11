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
const MessengerDriverScreener = "screener"

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
}

type TextChart struct {
	Width int
}

type HtmlChart struct {
	Width      string
	Height     string
	TitleShift string
	Fontsize   int

	IframeWidth      string
	IframeHeight     string
	IframeTitleShift string
	IframeFontsize   int
}

type Ui struct {
	Analytics    Analytics
	TextChart    TextChart
	HtmlChart    HtmlChart
	Version      string
	IncludingVat bool
}

// App Config struct to hold environment variables
type App struct {
	Ui              Ui
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
	for _, check := range []error{
		cfg.Ui.selfCheck("UI_"),
		cfg.DataRepository.selfCheck("DATA"),
		cfg.ChartRepository.selfCheck("CHART"),
		cfg.Server.selfCheck(),
		cfg.Messenger.selfCheck(),
	} {
		if check != nil {
			return check
		}
	}

	cfg.Location()
	return nil
}

func (u *Ui) selfCheck(prefix string) error {
	for _, check := range []error{
		u.Analytics.selfCheck(prefix),
		u.TextChart.selfCheck(prefix),
		u.HtmlChart.selfCheck(prefix),
	} {
		if check != nil {
			return check
		}
	}

	return nil
}

func (a *Analytics) selfCheck(prefix string) error {
	return checkRequiredFields(map[string]any{
		"ANALYTICS_HIGHPRICE": a.HighPrice,
		"ANALYTICS_LOWPRICE":  a.LowPrice,
	}, prefix)
}

func (a *Analytics) MinDate(localtion *time.Location) time.Time {
	return time.Date(2020, 1, 1, 0, 0, 0, 0, localtion)
}

func (t *TextChart) selfCheck(prefix string) error {
	return checkRequiredFields(map[string]any{
		"TEXTCHART_WIDTH": t.Width,
	}, prefix)
}

func (t *HtmlChart) selfCheck(prefix string) error {
	if t.Width == "" {
		t.Width = "900px"
	}
	if t.Height == "" {
		t.Height = "500px"
	}
	if t.TitleShift == "" {
		t.TitleShift = "39%"
	}
	if t.Fontsize == 0 {
		t.Fontsize = 16
	}

	if t.IframeWidth == "" {
		t.IframeWidth = "350px"
	}
	if t.IframeHeight == "" {
		t.IframeHeight = "320px"
	}
	if t.IframeTitleShift == "" {
		t.IframeTitleShift = "25%"
	}
	if t.IframeFontsize == 0 {
		t.IframeFontsize = 8
	}
	return nil
}

func (s *Server) selfCheck() error {
	return checkRequiredFields(map[string]any{
		"SERVER_PORT": s.Port,
	}, "")
}

func (g *GroupCache) selfCheck() error {
	return checkRequiredFields(map[string]any{
		"GROUPCACHE_ME":     g.Me,
		"GROUPCACHE_LISTEN": g.Listen,
	}, "")
}

func (i *Influx) selfCheck() error {
	return checkRequiredFields(map[string]any{
		"INFLUX_URL":     i.Url,
		"INFLUX_ORGNAME": i.Orgname,
		"INFLUX_BUCKET":  i.Bucket,
		"INFLUX_TOKEN":   i.Token,
	}, "")
}

func (e *Energyzero) selfCheck() error {
	return e.API.selfCheck("DATAREPOSITORY_ENERGYZERO_")
}

func (a *Api) selfCheck(prefix string) error {
	return checkRequiredFields(map[string]any{
		"API_ENDPOINT": a.Endpoint,
	}, prefix)
}

func (r *Repository) selfCheck(prefix string) error {
	drivers := strings.Split(r.Driver, ",")

	for _, driver := range drivers {
		driver = strings.TrimSpace(driver)
		if check, ok := map[string]func() error{
			RepositoryDriverGroupCache: r.GroupCache.selfCheck,
			RepositoryDriverInflux:     r.Influx.selfCheck,
			RepositoryDriverEnergyzero: r.Energyzero.selfCheck,
			RepositoryDriverStub:       func() error { return nil },
		}[driver]; ok {
			if err := check(); err != nil {
				return err
			}
		} else if driver == "" {
			return fmt.Errorf("%sREPOSITORY_DRIVER not set", prefix)
		} else {
			return fmt.Errorf("unknown %sREPOSITORY_DRIVER: %s", prefix, driver)
		}
	}

	return nil
}

func (m *Messenger) selfCheck() error {
	switch m.Driver {
	case MessengerDriverTelegram:
		if err := checkRequiredFields(map[string]any{
			"MESSENGER_TELEGRAM_TOKEN":  m.Telegram.Token,
			"MESSENGER_TELEGRAM_CHATID": m.Telegram.ChatID,
		}, ""); err != nil {
			return err
		}
	case MessengerDriverScreener:
		// No required fields for Screener
	case "":
		return errors.New("MESSENGER_DRIVER not set")
	default:
		return fmt.Errorf("unknown MESSENGER_DRIVER: %s", m.Driver)
	}

	return nil
}

func checkRequiredFields(fields map[string]any, prefix string) (err error) {
	for fieldName, fieldValue := range fields {
		switch fieldValue.(type) {
		case string:
			if fieldValue.(string) == "" {
				err = fmt.Errorf("%s%s not set", prefix, fieldName)
				return
			}
		case decimal.Decimal:
			if fieldValue.(decimal.Decimal).IsZero() {
				err = fmt.Errorf("%s%s not set", prefix, fieldName)
				return
			}
		case float64:
			if fieldValue.(float64) == 0 {
				err = fmt.Errorf("%s%s not set", prefix, fieldName)
				return
			}
		case int:
			if fieldValue.(int) == 0 {
				err = fmt.Errorf("%s%s not set", prefix, fieldName)
				return
			}
		case int64:
			if fieldValue.(int64) == 0 {
				err = fmt.Errorf("%s%s not set", prefix, fieldName)
				return
			}
		default:
			err = fmt.Errorf("%s%s undefined type", prefix, fieldName)
		}
	}
	return
}
