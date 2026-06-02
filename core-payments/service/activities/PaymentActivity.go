package activities

import "github.com/uptrace/bun"

type PaymentActivity struct {
	PaymentsDB *bun.DB
}

func NewPaymentActivity(paymentsDB *bun.DB) *PaymentActivity {
	return &PaymentActivity{
		PaymentsDB: paymentsDB,
	}
}
