package notifications

import (
	"context"

	"github.com/openreserveio/core/core-util/bus"
)

type PaymentStatusNotifier struct {
	BusConn *bus.BusConnection
}

func NewPaymentStatusNotifier(ctx context.Context, busUrl string) *PaymentStatusNotifier {

	psn := PaymentStatusNotifier{}



}
