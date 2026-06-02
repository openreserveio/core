package activities

import (
	"context"

	"github.com/google/uuid"
	"github.com/openreserveio/core/core-payments/pmtmodel"
	"github.com/uptrace/bun"
)

func StoreRawPaymentInstruction(ctx context.Context, db *bun.DB, rawPaymentInstruction []byte) (pmtmodel.Payment, error) {

	// Stores the raw payment instruction for processing later in the workflow
	var payment pmtmodel.Payment
	payment.ID = uuid.NewString()
	payment.PaymentMessage = rawPaymentInstruction
	_, err := db.NewInsert().Model(&payment).Exec(ctx.(context.Context))
	if err != nil {
		return pmtmodel.Payment{}, err
	}

	return payment, nil
}
