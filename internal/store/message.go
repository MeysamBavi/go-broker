package store

import (
	"context"
	"errors"
	"github.com/MeysamBavi/go-broker/internal/tracing"
	"github.com/MeysamBavi/go-broker/pkg/broker"
	"go.opentelemetry.io/otel/trace"
	"time"
)

const (
	packageName = "/internal/store"
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

type withTracing struct {
	core           Message
	tracerProvider trace.TracerProvider
}

func MessageWithTracing(coreMessage Message, provider trace.TracerProvider) Message {
	return &withTracing{
		core:           coreMessage,
		tracerProvider: provider,
	}
}

func (w *withTracing) tracer() trace.Tracer {
	return w.tracerProvider.Tracer(packageName + ".Message")
}

func (w *withTracing) SaveMessage(ctx context.Context, subject string, message *broker.Message) error {
	ctx, span := w.tracer().Start(ctx, "SaveMessage")
	defer span.End()

	span.SetAttributes(tracing.Subject(subject))

	err := w.core.SaveMessage(ctx, subject, message)

	span.SetAttributes(tracing.MessageAssignedId(message.Id))

	tracing.SetStatusAndError(span, err)

	return err
}

func (w *withTracing) GetMessage(ctx context.Context, subject string, id int) (*broker.Message, error) {
	ctx, span := w.tracer().Start(ctx, "GetMessage")
	defer span.End()

	span.SetAttributes(tracing.Subject(subject))
	span.SetAttributes(tracing.MessageId(id))

	m, err := w.core.GetMessage(ctx, subject, id)

	tracing.SetStatusAndError(span, err)

	return m, err
}
