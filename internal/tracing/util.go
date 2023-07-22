package tracing

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	subjectKey    = "subject"
	idKey         = "id"
	assignedIdKey = "assignedId"
)

func SetStatusAndError(span trace.Span, err error) {
	span.RecordError(err)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
	}
}

func Subject(val string) attribute.KeyValue {
	return attribute.String(subjectKey, val)
}

func MessageId(id int) attribute.KeyValue {
	return attribute.Int(idKey, id)
}

func MessageAssignedId(id int) attribute.KeyValue {
	return attribute.Int(assignedIdKey, id)
}
