package messenger

import "context"

type Messenger interface {
	SendMessage(ctx context.Context, message string) (err error)
}
