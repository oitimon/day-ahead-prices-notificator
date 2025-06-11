package messenger

import (
	"context"
	"log"
)

type Screener struct {
}

func NewScreener() Messenger {
	return &Screener{}
}

func (s *Screener) SendMessage(_ context.Context, message string) error {
	log.Printf("Screener message:\n%s\n", message)

	return nil
}
