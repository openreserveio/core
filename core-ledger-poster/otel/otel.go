package otel

import (
	"context"
	"errors"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
	"go.opentelemetry.io/contrib/exporters/autoexport"
	ot "go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
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

func AddError(errorName string, err error, vars ...interface{}) {
	messageError := fmt.Errorf("APP ERROR: %s - %v", fmt.Sprintf(errorName, vars...), err)
	currentSpan.RecordError(errors.New(messageError.Error()))
}

func InjectNatsHeaders(ctx context.Context, msg *nats.Msg) {

	prop := ot.GetTextMapPropagator()
	prop.Inject(ctx, natsHeaderCarrier(msg.Header))

}

func ExtractNatsContext(request micro.Request) context.Context {

	prop := ot.GetTextMapPropagator()
	ctx := prop.Extract(context.Background(), natsHeaderCarrier(request.Headers()))
	return ctx

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

type natsHeaderCarrier nats.Header

func (c natsHeaderCarrier) Get(key string) string {
	return nats.Header(c).Get(key)
}

func (c natsHeaderCarrier) Set(key string, value string) {
	nats.Header(c).Set(key, value)
}

func (c natsHeaderCarrier) Keys() []string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	return keys
}
