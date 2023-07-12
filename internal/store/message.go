package store

import (
	"context"
	"errors"
	"github.com/MeysamBavi/go-broker/pkg/broker"
)

type Message interface {
	SaveMessage(ctx context.Context, subject string, message *broker.Message) error
	GetMessage(ctx context.Context, subject string, id int) (*broker.Message, error)
}

var (
	ErrInvalidId = errors.New("no message exists with given id")
	ErrExpired   = errors.New("this message is expired")
)
