package bus

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go/micro"
	"github.com/openreserveio/core/core-util/otel"
	log "github.com/sirupsen/logrus"
)

func Reply(ctx context.Context, request micro.Request, status int, responseData interface{}, err error) error {

	otel.StartSpan(ctx, "bus.Reply")
	defer otel.EndSpan(ctx)

	// Encode Response data
	responseBytes := Encode(ctx, responseData)

	// new bus message for error
	busMessage := BusMessage{
		ID:          uuid.NewString(),
		Data:        responseBytes,
		IsReply:     true,
		ReplyStatus: status,
	}
	if err != nil {
		busMessage.ReplyStatusDetail = err.Error()
	}

	bmRaw := Encode(ctx, &busMessage)

	otel.AddEvent("Responding to request")
	if err := request.Respond(bmRaw); err != nil {
		otel.AddError("Error responding to request", err)
		log.Error("Error responding: %v", err)
		return err
	}

	return nil

}

func ReplyOK(ctx context.Context, request micro.Request, responseData interface{}) error {

	otel.StartSpan(ctx, "bus.ReplyOK")
	defer otel.EndSpan(ctx)

	// Encode Response data
	responseBytes := Encode(ctx, responseData)

	// new bus message for error
	busMessage := BusMessage{
		ID:                uuid.NewString(),
		Data:              responseBytes,
		IsReply:           true,
		ReplyStatus:       http.StatusOK,
		ReplyStatusDetail: "OK",
	}
	bmRaw := Encode(ctx, &busMessage)

	if err := request.Respond(bmRaw); err != nil {
		otel.AddError("Error responding to request", err)
		log.Error("Error responding: %v", err)
		return err
	}

	return nil

}

func ReplyWithNotFound(ctx context.Context, request micro.Request) error {

	otel.StartSpan(ctx, "bus.ReplyWithNotFound")
	defer otel.EndSpan(ctx)

	// new bus message for error
	errorBusMessage := BusMessage{
		ID:                uuid.NewString(),
		IsReply:           true,
		ReplyStatus:       http.StatusNotFound,
		ReplyStatusDetail: "Not Found",
	}
	ebmRaw := Encode(ctx, &errorBusMessage)

	if err := request.Respond(ebmRaw); err != nil {
		otel.AddError("Error responding to request", err)
		log.Error("Error responding with error: %v", err)
		return err
	}

	return nil

}

func ReplyWithSystemError(ctx context.Context, request micro.Request, sysErr error) error {

	otel.StartSpan(ctx, "bus.ReplyWithSystemError")
	defer otel.EndSpan(ctx)

	// new bus message for error
	errorBusMessage := BusMessage{
		ID:                uuid.NewString(),
		Data:              request.Data(),
		IsReply:           true,
		ReplyStatus:       http.StatusInternalServerError,
		ReplyStatusDetail: sysErr.Error(),
	}
	ebmRaw := Encode(ctx, &errorBusMessage)

	if err := request.Respond(ebmRaw); err != nil {
		log.Error("Error responding with error: %v", err)
		return err
	}

	return nil
}

func ReplyWithBadRequestError(ctx context.Context, request micro.Request, badReqErr error) error {

	otel.StartSpan(ctx, "bus.ReplyWithBadRequestError")
	defer otel.EndSpan(ctx)

	// new bus message for error
	errorBusMessage := BusMessage{
		ID:                uuid.NewString(),
		IsReply:           true,
		ReplyStatus:       http.StatusBadRequest,
		Data:              request.Data(),
		ReplyStatusDetail: badReqErr.Error(),
	}
	ebmRaw := Encode(ctx, &errorBusMessage)

	if err := request.Respond(ebmRaw); err != nil {
		otel.AddError("Error responding to request", err)
		log.Error("Error responding with error: %v", err)
		return err
	}

	return nil
}
