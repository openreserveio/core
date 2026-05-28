package bus

import (
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
)

type BusConnection struct {
	ConnUrl string
	Conn    *nats.Conn
}

func NewBusConnection(connUrl string) (*BusConnection, error) {
	nc, err := nats.Connect(connUrl)
	if err != nil {
		log.Error("Error connecting to NATS bus:  %v", err)
		return nil, err
	}
	return &BusConnection{Conn: nc, ConnUrl: connUrl}, nil
}
