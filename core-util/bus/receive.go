package bus

import (
	"errors"
	"net/http"

	"github.com/nats-io/nats.go/micro"
	"github.com/openreserveio/core/core-util/otel"
	log "github.com/sirupsen/logrus"
)

func Receive(request micro.Request, target interface{}) error {

	// Extract Context
	ctx := otel.ExtractNatsContext(request)
	ctx, st := otel.StartSpan(ctx, "bus.Receive")
	defer otel.EndSpan(ctx, st)

	// Decode Request Data
	otel.AddEvent(st, "Decoding Request Data")
	var busMessage BusMessage
	if err := Decode(ctx, request.Data(), &busMessage); err != nil {
		otel.AddError(st, "Error decoding request data", err)
		log.Error("Error decoding request data: %v", err)
		return err
	}

	if busMessage.ReplyStatus >= http.StatusBadRequest && busMessage.ReplyStatus < http.StatusInternalServerError {
		otel.AddError(st, "Error received in request: %v", nil, busMessage.ReplyStatusDetail)
		log.Error("Error received in request: %v", busMessage.ReplyStatusDetail)
		return BadRequest(busMessage.ReplyStatusDetail, nil)
	}

	if busMessage.ReplyStatus >= http.StatusInternalServerError {
		otel.AddError(st, "System Failure Error received in request: %v", nil, busMessage.ReplyStatusDetail)
		log.Error("System Failure Error received in request: %v", busMessage.ReplyStatusDetail)
		return BadRequest(busMessage.ReplyStatusDetail, errors.New(busMessage.ReplyStatusDetail))
	}

	// Decode the data in the message
	otel.AddEvent(st, "Decoding Message Data")
	if err := Decode(ctx, busMessage.Data, target); err != nil {
		otel.AddError(st, "Error decoding message data", err)
		log.Error("Error decoding message data: %v", err)
		return err
	}

	return nil

}
