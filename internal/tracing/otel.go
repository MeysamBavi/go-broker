package tracing

import (
	"context"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	traceSdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	"io"
	"log"
	"os"
	"time"
)

const (
	shutdownTimeout = time.Second * 4
)

func NewTracerProvider(config Config) (tp trace.TracerProvider, shutdown func()) {
	if !config.Enabled {
		return trace.NewNoopTracerProvider(), func() {}
	}

	var exporter traceSdk.SpanExporter
	if config.UseJaeger {
		//TODO use jaeger exporter
	} else {
		file, err := os.Create(config.OutputFile)
		if err != nil {
			log.Fatal(err)
		}
		exporter, err = newStdoutExporter(file)
		if err != nil {
			log.Fatal(err)
		}
	}

	tpp := traceSdk.NewTracerProvider(
		traceSdk.WithBatcher(exporter),
		traceSdk.WithResource(GetResource()),
	)
	shutdown = func() {
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		tpp.ForceFlush(ctx)
		if err := tpp.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}

	return tpp, shutdown
}

func GetResource() *resource.Resource {
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("go-broker"),
			semconv.ServiceVersion("v0.1.0"),
		),
	)
	if err != nil {
		panic(err)
	}

	return r
}

func newStdoutExporter(w io.Writer) (traceSdk.SpanExporter, error) {
	return stdouttrace.New(
		stdouttrace.WithWriter(w),
		stdouttrace.WithPrettyPrint(),
		stdouttrace.WithoutTimestamps(),
	)
}
