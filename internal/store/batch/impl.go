package batch

import (
	"context"
	"fmt"
	"github.com/MeysamBavi/go-broker/internal/tracing"
	"github.com/MeysamBavi/go-broker/pkg/broker"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"time"
)

const packageName = "internal/store/batch"

type impl struct {
	writer     Writer
	config     Config
	itemStream chan *Item
}

func NewHandler(config Config, writer Writer, tp trace.TracerProvider) Handler {
	h := &impl{
		writer:     writer,
		config:     config,
		itemStream: make(chan *Item, 1),
	}
	go h.flusher(tp.Tracer(packageName + ".Handler"))

	return h
}

func (h *impl) flusher(tracer trace.Tracer) {
	ctx, span := tracer.Start(context.Background(), "flusher")
	defer span.End()

	ticker := time.NewTicker(h.config.Timeout)
	buffer := make([]*Item, 0, h.config.Size)

	flush := func(causedByTimeout bool) {
		if len(buffer) == 0 {
			return
		}

		ctx, span := tracer.Start(ctx, "flush")
		defer span.End()

		span.SetAttributes(attribute.Int("batchSize", len(buffer)))
		span.SetAttributes(attribute.Bool("causedByTimeout", causedByTimeout))

		err := h.writer(ctx, buffer)

		tracing.SetStatusAndError(span, err)

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
			flush(true)
		case item := <-h.itemStream:
			buffer = append(buffer, item)
			if len(buffer) == cap(buffer) {
				flush(false)
			}
		}
	}
}

func (h *impl) AddAndWait(ctx context.Context, subject string, message *broker.Message) error {
	item := Item{
		Subject: subject,
		Message: message,
		resolve: make(chan struct{}),
	}
	h.itemStream <- &item

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-item.resolve:
	}

	return item.Err
}
