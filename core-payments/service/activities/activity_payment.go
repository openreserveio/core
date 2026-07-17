package activities

import (
	"github.com/openreserveio/core/core-payments/generated/glmodel"
	"github.com/openreserveio/core/core-util/bus"
	"github.com/uptrace/bun"
)

type PaymentActivity struct {
	PaymentsDB   *bun.DB
	EntityDB     *bun.DB
	CoreGLClient glmodel.GeneralLedgerServiceClient
	BusConn      *bus.BusConnection
}

func NewPaymentActivity(paymentsDB *bun.DB, entityDB *bun.DB, coreGLClient glmodel.GeneralLedgerServiceClient, busConn *bus.BusConnection) *PaymentActivity {

	return &PaymentActivity{
		PaymentsDB:   paymentsDB,
		EntityDB:     entityDB,
		CoreGLClient: coreGLClient,
		BusConn:      busConn,
	}
}
