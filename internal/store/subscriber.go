package store

import (
	"context"
	"github.com/MeysamBavi/go-broker/internal/tracing"
	"github.com/MeysamBavi/go-broker/pkg/broker"
	"go.opentelemetry.io/otel/trace"
)

type Subscriber interface {
	AddSubscriber(ctx context.Context, subject string, callBack OnPublishFunc)
	Publish(ctx context.Context, subject string, message *broker.Message)
}

type OnPublishFunc func(message *broker.Message)

type subscriberWithTracing struct {
	core           Subscriber
	tracerProvider trace.TracerProvider
}

func SubscriberWithTracing(core Subscriber, tracerProvider trace.TracerProvider) Subscriber {
	return &subscriberWithTracing{
		core:           core,
		tracerProvider: tracerProvider,
	}
}

func (s *subscriberWithTracing) tracer() trace.Tracer {
	return s.tracerProvider.Tracer(packageName + ".Subscriber")
}

func (s *subscriberWithTracing) AddSubscriber(ctx context.Context, subject string, callBack OnPublishFunc) {
	ctx, span := s.tracer().Start(ctx, "AddSubscriber")
	defer span.End()

	span.SetAttributes(tracing.Subject(subject))

	s.core.AddSubscriber(ctx, subject, callBack)
}

func (s *subscriberWithTracing) Publish(ctx context.Context, subject string, message *broker.Message) {
	ctx, span := s.tracer().Start(ctx, "Publish")
	defer span.End()

	span.SetAttributes(tracing.Subject(subject))
	span.SetAttributes(tracing.MessageId(message.Id))

	s.core.Publish(ctx, subject, message)
}
