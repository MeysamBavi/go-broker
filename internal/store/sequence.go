package store

import (
	"context"
	"github.com/MeysamBavi/go-broker/internal/tracing"
	"go.opentelemetry.io/otel/trace"
)

type Sequence interface {
	CreateNewId(ctx context.Context, subject string) (int32, error)
	Load(ctx context.Context, subject string, lastId int32) error
}

type sequenceWithTracing struct {
	trace.TracerProvider
	core Sequence
}

func SequenceWithTracing(core Sequence, tracerProvider trace.TracerProvider) Sequence {
	return &sequenceWithTracing{
		TracerProvider: tracerProvider,
		core:           core,
	}
}

func (s *sequenceWithTracing) tracer() trace.Tracer {
	return s.Tracer(packageName + ".Sequence")
}

func (s *sequenceWithTracing) CreateNewId(ctx context.Context, subject string) (int32, error) {
	ctx, span := s.tracer().Start(ctx, "CreateNewId")
	defer span.End()

	span.SetAttributes(tracing.Subject(subject))

	id, err := s.core.CreateNewId(ctx, subject)

	span.SetAttributes(tracing.MessageAssignedId(int(id)))
	tracing.SetStatusAndError(span, err)

	return id, err
}

func (s *sequenceWithTracing) Load(ctx context.Context, subject string, lastId int32) error {
	ctx, span := s.tracer().Start(ctx, "Load")
	defer span.End()

	span.SetAttributes(tracing.Subject(subject))
	span.SetAttributes(tracing.MessageId(int(lastId)))

	err := s.core.Load(ctx, subject, lastId)

	tracing.SetStatusAndError(span, err)

	return err
}
