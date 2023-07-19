package broker

import (
	"context"
	"github.com/MeysamBavi/go-broker/internal/tracing"
	"github.com/MeysamBavi/go-broker/pkg/broker"
	"go.opentelemetry.io/otel/trace"
)

const (
	packageName = "/internal/broker"
)

func WithTracing(core broker.Broker, provider trace.TracerProvider) broker.Broker {
	return &withTracing{
		TracerProvider: provider,
		core:           core,
	}
}

type withTracing struct {
	trace.TracerProvider
	core broker.Broker
}

func (w *withTracing) tracer() trace.Tracer {
	return w.Tracer(packageName + ".Module")
}

func (w *withTracing) Close() error {
	return w.core.Close()
}

func (w *withTracing) Publish(ctx context.Context, subject string, msg broker.Message) (int, error) {
	ctx, span := w.tracer().Start(ctx, "Publish")
	defer span.End()

	span.SetAttributes(tracing.Subject(subject))

	id, err := w.core.Publish(ctx, subject, msg)

	tracing.SetStatusAndError(span, err)

	return id, err
}

func (w *withTracing) Subscribe(ctx context.Context, subject string) (<-chan broker.Message, error) {
	ctx, span := w.tracer().Start(ctx, "Subscribe")
	defer span.End()

	span.SetAttributes(tracing.Subject(subject))

	ch, err := w.core.Subscribe(ctx, subject)

	tracing.SetStatusAndError(span, err)

	return ch, err
}

func (w *withTracing) Fetch(ctx context.Context, subject string, id int) (broker.Message, error) {
	ctx, span := w.tracer().Start(ctx, "Fetch")
	defer span.End()

	span.SetAttributes(tracing.Subject(subject))
	span.SetAttributes(tracing.MessageId(id))

	msg, err := w.core.Fetch(ctx, subject, id)

	tracing.SetStatusAndError(span, err)

	return msg, err
}
