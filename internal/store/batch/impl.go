package batch

import (
	"context"
	"fmt"
	"github.com/MeysamBavi/go-broker/pkg/broker"
	"time"
)

type impl struct {
	writer     Writer
	config     Config
	itemStream chan *Item
}

func NewHandler(config Config, writer Writer) Handler {
	h := &impl{
		writer:     writer,
		config:     config,
		itemStream: make(chan *Item, 1),
	}
	go h.flusher()

	return h
}

func (h *impl) flusher() {
	ticker := time.NewTicker(h.config.Timeout)
	buffer := make([]*Item, 0, h.config.Size)
	flush := func() {
		if len(buffer) == 0 {
			return
		}

		err := h.writer(buffer)
		for _, item := range buffer {
			if item.Err == nil {
				item.Err = err
			} else if err != nil {
				item.Err = fmt.Errorf("%s: %w", err.Error(), item.Err)
			}
			close(item.resolve)
		}
		buffer = buffer[:0]
		ticker.Reset(h.config.Timeout)
	}

	for {
		select {
		case <-ticker.C:
			flush()
		case item := <-h.itemStream:
			buffer = append(buffer, item)
			if len(buffer) == cap(buffer) {
				flush()
			}
		}
	}
}

func (h *impl) AddAndWait(ctx context.Context, subject string, message *broker.Message) error {
	item := Item{
		Subject: subject,
		Message: message,
		resolve: make(chan struct{}),
		Ctx:     ctx,
	}
	h.itemStream <- &item

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-item.resolve:
	}

	return item.Err
}
