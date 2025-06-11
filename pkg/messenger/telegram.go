package messenger

import (
	"context"
	"errors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/oitimon/day-ahead-prices-notificator/pkg/config"
	"log"
)

type Telegram struct {
	cfg    *config.Telegram
	client *tgbotapi.BotAPI
}

func NewTelegram(cfg *config.Telegram) (Messenger, error) {
	botApi, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		return nil, errors.New("error creating Telegram Bot: " + err.Error())
	}
	return &Telegram{
		cfg:    cfg,
		client: botApi,
	}, nil
}

func (tg *Telegram) SendMessage(_ context.Context, message string) error {
	log.Printf("Sending message to Telegram: \n%s\n", message)

	msg := tgbotapi.NewMessage(tg.cfg.ChatID, tg.sanitiseMessage(message))
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	if _, err := tg.client.Send(msg); err != nil {
		return errors.New("error sending Telegram message: " + err.Error())
	}

	return nil
}

// sanitiseMessage Escape special characters for MarkdownV2
func (tg *Telegram) sanitiseMessage(message string) string {
	//message = strings.Replace(message, "-", "\\-", -1)
	//message = strings.Replace(message, ".", "\\.", -1)
	//
	//return message
	return tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, message)
}
