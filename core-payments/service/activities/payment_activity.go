package activities

import (
	"github.com/openreserveio/core/core-payments/generated/glmodel"
	"github.com/uptrace/bun"
)

type PaymentActivity struct {
	PaymentsDB   *bun.DB
	EntityDB     *bun.DB
	CoreGLClient glmodel.GeneralLedgerServiceClient
}

func NewPaymentActivity(paymentsDB *bun.DB, entityDB *bun.DB, coreGLClient glmodel.GeneralLedgerServiceClient) *PaymentActivity {
	return &PaymentActivity{
		PaymentsDB:   paymentsDB,
		EntityDB:     entityDB,
		CoreGLClient: coreGLClient,
	}
}
