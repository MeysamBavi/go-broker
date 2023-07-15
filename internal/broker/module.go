package broker

import (
	"context"
	"fmt"
	"github.com/MeysamBavi/go-broker/internal/store"
	"github.com/MeysamBavi/go-broker/pkg/broker"
	"sync"
	"time"
)

type Module struct {
	msgStore store.Message
	closed   bool
}

func NewModule() broker.Broker {
	return &Module{
		msgStore: store.NewInMemoryMessage(timeProvider{}),
		closed:   false,
	}
}

type timeProvider struct{}

func (tp timeProvider) GetCurrentTime() time.Time {
	return time.Now()
}

func (m *Module) Close() error {
	m.closed = true
	return nil
}

func (m *Module) Publish(ctx context.Context, subject string, msg broker.Message) (int, error) {
	if m.closed {
		return 0, broker.ErrUnavailable
	}

	err := m.msgStore.SaveMessage(ctx, subject, &msg)
	if err != nil {
		return 0, fmt.Errorf("unexpected error while saving message: %w", err)
	}

	return msg.Id, nil
}

func (m *Module) Subscribe(ctx context.Context, subject string) (<-chan broker.Message, error) {
	if m.closed {
		return nil, broker.ErrUnavailable
	}

	var wg sync.WaitGroup
	ch := make(chan broker.Message)
	wg.Add(1)
	go func() {
		ctx := context.Background()
		message, err := m.msgStore.GetNextMessage(ctx, subject, nil, func() {
			wg.Done()
		})
		for {
			if err != nil {
				// TODO: Log error
				fmt.Println(err)
				close(ch)
				return
			}
			ch <- *message
			message, err = m.msgStore.GetNextMessage(ctx, subject, &message.Id, nil)
		}
	}()

	wg.Wait()
	return ch, nil
}

func (m *Module) Fetch(ctx context.Context, subject string, id int) (broker.Message, error) {
	var emptyResult broker.Message
	if m.closed {
		return emptyResult, broker.ErrUnavailable
	}

	msg, err := m.msgStore.GetMessage(ctx, subject, id)

	if err == store.ErrInvalidId {
		return emptyResult, broker.ErrInvalidID
	}

	if err == store.ErrExpired {
		return emptyResult, broker.ErrExpiredID
	}

	if err != nil {
		return emptyResult, fmt.Errorf("unexpected error while getting message: %w", err)
	}

	return *msg, nil
}
