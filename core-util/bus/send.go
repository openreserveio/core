package bus

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/openreserveio/core/core-util/otel"
	log "github.com/sirupsen/logrus"
)

func Send(ctx context.Context, busConn *BusConnection, subject string, data interface{}) error {

	ctx, st := otel.StartSpan(ctx, "bus.Send")
	defer otel.EndSpan(ctx, st)

	// Encode Request data
	dataBytes := Encode(ctx, data)

	// new bus message
	busMessage := BusMessage{
		ID:      uuid.NewString(),
		Data:    dataBytes,
		IsReply: false,
	}
	bmRaw := Encode(ctx, &busMessage)

	// Create a NATS message and inject headers for OTEL
	natsMsg := nats.NewMsg(subject)
	otel.InjectNatsHeaders(ctx, natsMsg)
	natsMsg.Data = bmRaw

	if err := busConn.Conn.PublishMsg(natsMsg); err != nil {
		otel.AddError(st, "Error sending: %v", err)
		log.Error("Error sending: %v", err)
		return err
	}

	return nil

}

func SendForReply(ctx context.Context, busConn *BusConnection, timeout time.Duration, subject string, messageData interface{}, replyMessage interface{}) error {

	ctx, st := otel.StartSpan(ctx, "bus.SendForReply")
	defer otel.EndSpan(ctx, st)

	// Encode Request data
	dataBytes := Encode(ctx, messageData)

	// new bus message
	busMessage := BusMessage{
		ID:      uuid.NewString(),
		Data:    dataBytes,
		IsReply: true,
	}
	bmRaw := Encode(ctx, &busMessage)

	otel.AddEvent(st, "Sending message for reply: %v", busMessage)
	var err error
	var rep *nats.Msg
	if rep, err = busConn.Conn.Request(subject, bmRaw, timeout); err != nil {
		otel.AddError(st, "Error sending for reply: %v", err)
		log.Error("Error sending for reply: %v", err)
		return err
	}

	// Decode the reply
	var replyBusMessage BusMessage
	if err = Decode(ctx, rep.Data, &replyBusMessage); err != nil {
		otel.AddError(st, "Error decoding reply: %v", err)
		log.Error("Error decoding reply: %v", err)
		return err
	}

	// If reply msg is nil, no need to decode the data
	if replyMessage == nil || replyBusMessage.Data == nil {
		otel.AddEvent(st, "Reply message was nil!")
		return nil
	}

	// Decode the data in the reply
	if err = Decode(ctx, replyBusMessage.Data, replyMessage); err != nil {
		otel.AddError(st, "Error decoding reply data: %v", err)
		log.Error("Error decoding reply data: %v", err)
		return err
	}

	return nil

}
