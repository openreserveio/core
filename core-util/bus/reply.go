package bus

import (
	"github.com/google/uuid"
	"github.com/nats-io/nats.go/micro"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func Reply(request micro.Request, status int, responseData interface{}, err error) error {

	// Encode Response data
	responseBytes := Encode(responseData)

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

	bmRaw := Encode(&busMessage)

	if err := request.Respond(bmRaw); err != nil {
		log.Error("Error responding: %v", err)
		return err
	}

	return nil

}

func ReplyOK(request micro.Request, responseData interface{}) error {

	// Encode Response data
	responseBytes := Encode(responseData)

	// new bus message for error
	busMessage := BusMessage{
		ID:                uuid.NewString(),
		Data:              responseBytes,
		IsReply:           true,
		ReplyStatus:       http.StatusOK,
		ReplyStatusDetail: "OK",
	}
	bmRaw := Encode(&busMessage)

	if err := request.Respond(bmRaw); err != nil {
		log.Error("Error responding: %v", err)
		return err
	}

	return nil

}

func ReplyWithNotFound(request micro.Request) error {

	// new bus message for error
	errorBusMessage := BusMessage{
		ID:                uuid.NewString(),
		IsReply:           true,
		ReplyStatus:       http.StatusNotFound,
		ReplyStatusDetail: "Not Found",
	}
	ebmRaw := Encode(&errorBusMessage)

	if err := request.Respond(ebmRaw); err != nil {
		log.Error("Error responding with error: %v", err)
		return err
	}

	return nil

}

func ReplyWithSystemError(request micro.Request, sysErr error) error {

	// new bus message for error
	errorBusMessage := BusMessage{
		ID:                uuid.NewString(),
		Data:              request.Data(),
		IsReply:           true,
		ReplyStatus:       http.StatusInternalServerError,
		ReplyStatusDetail: sysErr.Error(),
	}
	ebmRaw := Encode(&errorBusMessage)

	if err := request.Respond(ebmRaw); err != nil {
		log.Error("Error responding with error: %v", err)
		return err
	}

	return nil
}

func ReplyWithBadRequestError(request micro.Request, badReqErr error) error {

	// new bus message for error
	errorBusMessage := BusMessage{
		ID:                uuid.NewString(),
		IsReply:           true,
		ReplyStatus:       http.StatusBadRequest,
		Data:              request.Data(),
		ReplyStatusDetail: badReqErr.Error(),
	}
	ebmRaw := Encode(&errorBusMessage)

	if err := request.Respond(ebmRaw); err != nil {
		log.Error("Error responding with error: %v", err)
		return err
	}

	return nil
}
