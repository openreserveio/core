package otel

import (
	"context"
	"fmt"

	"go.opentelemetry.io/contrib/exporters/autoexport"
	ot "go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	EXPORTER_TYPE_CONSOLE = "console"
	EXPORTER_TYPE_JAEGER  = "jaeger"
	EXPORTER_TYPE_OTLP    = "otlp"
)

var Tracer trace.Tracer
var currentSpan trace.Span

func StartSpan(ctx context.Context, spanName string) context.Context {
	newCtx, sp := Tracer.Start(ctx, spanName)
	currentSpan = sp
	return newCtx
}

func EndSpan(ctx context.Context) {
	currentSpan.End()
	currentSpan = nil
}

func AddEvent(eventName string, vars ...interface{}) {
	message := fmt.Sprintf(eventName, vars...)
	currentSpan.AddEvent(message)
}

func NewExporter(ctx context.Context, exporterType string) sdktrace.SpanExporter {

	switch exporterType {
	case EXPORTER_TYPE_CONSOLE:
		panic("Not Implemented")
	case EXPORTER_TYPE_JAEGER:
		panic("Not Implemented")
	case EXPORTER_TYPE_OTLP:
		spanExporter, err := autoexport.NewSpanExporter(ctx)
		if err != nil {
			panic("Failed to create OTLP span exporter")
		}
		return spanExporter
	}

	return nil

}

func NewTracerProvider(serviceName string, exp sdktrace.SpanExporter) {

	// Ensure default SDK resources and the required service name are set.
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		),
	)

	if err != nil {
		panic(err)
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(r),
	)

	ot.SetTracerProvider(tracerProvider)
	Tracer = tracerProvider.Tracer(serviceName)

}
