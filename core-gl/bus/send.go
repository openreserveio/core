package bus

import (
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	"time"
)

func Send(busConn *BusConnection, subject string, data interface{}) error {

	// Encode Request data
	dataBytes := Encode(data)

	// new bus message
	busMessage := BusMessage{
		ID:      uuid.NewString(),
		Data:    dataBytes,
		IsReply: false,
	}
	bmRaw := Encode(&busMessage)

	if err := busConn.Conn.Publish(subject, bmRaw); err != nil {
		log.Error("Error sending: %v", err)
		return err
	}

	return nil

}

func SendForReply(busConn *BusConnection, timeout time.Duration, subject string, messageData interface{}, replyMessage interface{}) error {

	// Encode Request data
	dataBytes := Encode(messageData)

	// new bus message
	busMessage := BusMessage{
		ID:      uuid.NewString(),
		Data:    dataBytes,
		IsReply: true,
	}
	bmRaw := Encode(&busMessage)

	var err error
	var rep *nats.Msg
	if rep, err = busConn.Conn.Request(subject, bmRaw, timeout); err != nil {
		log.Error("Error sending for reply: %v", err)
		return err
	}

	// Decode the reply
	var replyBusMessage BusMessage
	if err = Decode(rep.Data, &replyBusMessage); err != nil {
		log.Error("Error decoding reply: %v", err)
		return err
	}

	// If reply msg is nil, no need to decode the data
	if replyMessage == nil || replyBusMessage.Data == nil {
		return nil
	}

	// Decode the data in the reply
	if err = Decode(replyBusMessage.Data, replyMessage); err != nil {
		log.Error("Error decoding reply data: %v", err)
		return err
	}

	return nil

}
