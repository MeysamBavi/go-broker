package batch

import (
	"context"
	"github.com/MeysamBavi/go-broker/pkg/broker"
	"time"
)

type Item struct {
	Subject string
	Message *broker.Message
	Err     error
	resolve chan struct{}
}

type Writer func(ctx context.Context, values []*Item) error

type Config struct {
	Timeout time.Duration `config:"timeout"`
	Size    int           `config:"size"`
}

type Handler interface {
	AddAndWait(ctx context.Context, subject string, message *broker.Message) error
}
