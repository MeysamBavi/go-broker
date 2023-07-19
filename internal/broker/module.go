package broker

import (
	"context"
	"fmt"
	"github.com/MeysamBavi/go-broker/internal/store"
	"github.com/MeysamBavi/go-broker/pkg/broker"
)

const (
	subscribeChannelBuffer = 72
)

type Module struct {
	msgStore    store.Message
	subscribers store.Subscriber
	closed      bool
}

func NewModule() broker.Broker {
	return &Module{
		msgStore:    store.NewInMemoryMessage(store.GetDefaultTimeProvider()),
		subscribers: store.NewInMemorySubscriber(),
		closed:      false,
	}
}

func NewModuleWithStores(message store.Message, subscriber store.Subscriber) broker.Broker {
	return &Module{
		msgStore:    message,
		subscribers: subscriber,
		closed:      false,
	}
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
	m.subscribers.Publish(ctx, subject, &msg)

	return msg.Id, nil
}

func (m *Module) Subscribe(ctx context.Context, subject string) (<-chan broker.Message, error) {
	if m.closed {
		return nil, broker.ErrUnavailable
	}

	ch := make(chan broker.Message, subscribeChannelBuffer)
	callback := func(msg *broker.Message) {
		ch <- *msg
	}
	m.subscribers.AddSubscriber(ctx, subject, callback)

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
