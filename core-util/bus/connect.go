package bus

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/openreserveio/core/core-util/otel"
	log "github.com/sirupsen/logrus"
)

type BusConnection struct {
	ConnUrl string
	Conn    *nats.Conn
}

func NewBusConnection(ctx context.Context, connUrl string) (*BusConnection, error) {

	otel.StartSpan(ctx, "bus.NewBusConnection")
	defer otel.EndSpan(ctx)

	nc, err := nats.Connect(connUrl)
	if err != nil {
		otel.AddError("Error connecting to NATS bus", err)
		log.Error("Error connecting to NATS bus:  %v", err)
		return nil, err
	}

	otel.AddEvent("Connected to NATS bus")
	return &BusConnection{Conn: nc, ConnUrl: connUrl}, nil
}
