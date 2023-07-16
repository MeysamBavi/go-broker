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
}

var (
	ErrInvalidId = errors.New("no message exists with given id")
	ErrExpired   = errors.New("this message is expired")
)

type TimeProvider interface {
	GetCurrentTime() time.Time
}

type timeProvider struct{}

func GetDefaultTimeProvider() TimeProvider {
	return timeProvider{}
}

func (tp timeProvider) GetCurrentTime() time.Time {
	return time.Now()
}
