package messenger

import (
	"fmt"
	"github.com/oitimon/day-ahead-prices-notificator/pkg/config"
)

func NewMessenger(cfg *config.Messenger) (msgr Messenger, err error) {
	switch cfg.Driver {
	case config.MessengerDriverTelegram:
		msgr, err = NewTelegram(&cfg.Telegram)
	case config.MessengerDriverScreener:
		msgr = NewScreener()
	default:
		err = fmt.Errorf("unknown messenger driver: %s", cfg.Driver)
	}

	return
}
