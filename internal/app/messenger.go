package app

import (
	"errors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
)

var ErrUnknownMessengerDriver = errors.New("unknown messenger driver")

func SendMessage(cfg *ConfigMessenger, message string) (err error) {
	switch cfg.Driver {
	case messengerDriverTelegram:
		err = sendTelegram(&cfg.Telegram, message)
	default:
		err = ErrUnknownMessengerDriver
	}

	return
}

func sendTelegram(cfg *ConfigTelegram, message string) (err error) {
	client, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		err = errors.New("error creating Telegram Bot: " + err.Error())
		return
	}

	log.Printf("Sending messages to Telegram: %s\n", strings.Replace(message, "\n", " ", -1))

	msg := tgbotapi.NewMessage(cfg.ChatID, message)
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	if _, err = client.Send(msg); err != nil {
		err = errors.New("error sending Telegram message: " + err.Error())
		return
	}

	return
}
