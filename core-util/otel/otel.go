package otel

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
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

type SpanToken string

var Tracer trace.Tracer
var currentSpans map[SpanToken]trace.Span

func StartSpan(ctx context.Context, spanName string) (context.Context, SpanToken) {
	newCtx, sp := Tracer.Start(ctx, spanName)

	if currentSpans == nil {
		currentSpans = make(map[SpanToken]trace.Span)
	}
	spanId := SpanToken(uuid.NewString())
	currentSpans[spanId] = sp

	return newCtx, spanId
}

func EndSpan(ctx context.Context, spanToken SpanToken) {
	currentSpans[spanToken].End()
	currentSpans[spanToken] = nil
}

func AddEvent(spanToken SpanToken, eventName string, vars ...interface{}) {
	message := fmt.Sprintf(eventName, vars...)
	currentSpans[spanToken].AddEvent(message)
}

func AddError(spanToken SpanToken, errorName string, err error, vars ...interface{}) {

	if err == nil {
		err = errors.New("NIL ERROR")
	}
	messageError := fmt.Errorf("APP ERROR: %s - %v", fmt.Sprintf(errorName, vars...), err)
	currentSpans[spanToken].RecordError(errors.New(messageError.Error()))
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
