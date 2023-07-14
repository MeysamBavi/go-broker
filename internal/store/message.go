package store

import (
	"context"
	"errors"
	"github.com/MeysamBavi/go-broker/pkg/broker"
	"time"
)

type Message interface {
	SaveMessage(ctx context.Context, subject string, message *broker.Message) error
	GetMessage(ctx context.Context, subject string, id int) (*broker.Message, error)
	// GetNextMessage blocks until the next message on the subject is available and then returns.
	// If previousId is not nil, the next message of this id will be returned.
	// If beforeBlockCallBack is not nil, it will be called when the goroutine gets blocked or when the function returns. It will be called exactly once
	GetNextMessage(ctx context.Context, subject string, previousId *int, beforeBlockCallBack func()) (*broker.Message, error)
}

var (
	ErrInvalidId = errors.New("no message exists with given id")
	ErrExpired   = errors.New("this message is expired")
)

type TimeProvider interface {
	GetCurrentTime() time.Time
}
