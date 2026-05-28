package bus

import (
	"errors"
	"github.com/nats-io/nats.go/micro"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func Receive(request micro.Request, target interface{}) error {

	// Decode Request Data
	var busMessage BusMessage
	if err := Decode(request.Data(), &busMessage); err != nil {
		log.Error("Error decoding request data: %v", err)
		return err
	}

	if busMessage.ReplyStatus >= http.StatusBadRequest && busMessage.ReplyStatus < http.StatusInternalServerError {
		log.Error("Error received in request: %v", busMessage.ReplyStatusDetail)
		return BadRequest(busMessage.ReplyStatusDetail, nil)
	}

	if busMessage.ReplyStatus >= http.StatusInternalServerError {
		log.Error("System Failure Error received in request: %v", busMessage.ReplyStatusDetail)
		return BadRequest(busMessage.ReplyStatusDetail, errors.New(busMessage.ReplyStatusDetail))
	}

	// Decode the data in the message
	if err := Decode(busMessage.Data, target); err != nil {
		log.Error("Error decoding message data: %v", err)
		return err
	}

	return nil

}
