package activities

import (
	"context"

	"github.com/google/uuid"
	"github.com/openreserveio/core/core-payments/pmtmodel"
)

func (act *PaymentActivity) StoreRawPaymentInstruction(ctx context.Context, rawPaymentInstruction []byte) (pmtmodel.Payment, error) {

	// Stores the raw payment instruction for processing later in the workflow
	var payment pmtmodel.Payment
	payment.ID = uuid.NewString()
	payment.PaymentMessage = rawPaymentInstruction
	payment.CurrentPaymentStatus = pmtmodel.PAYMENT_STATUS_INSTRUCTION_RECEIVED
	// _, err := db.NewInsert().Model(&payment).Exec(ctx)
	insertQuery := act.PaymentsDB.NewInsert()
	insertQuery = insertQuery.Model(&payment)
	_, err := insertQuery.Exec(ctx)
	if err != nil {
		return pmtmodel.Payment{}, err
	}

	return payment, nil
}
