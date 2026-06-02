package activities

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/openreserveio/core/core-payments/pmtmodel"
	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

func (act *PaymentActivity) UpdatePaymentStatus(ctx context.Context, payment pmtmodel.Payment, paymentStatus string) error {

	payment.CurrentPaymentStatus = paymentStatus
	_, err := act.PaymentsDB.NewUpdate().Model(&payment).Where("id = ?", payment.ID).Exec(ctx)
	if err != nil {
		log.Errorf("Error updating status update step: %v", err)
		return err
	}

	paymentStatusHistory := pmtmodel.PaymentStatusHistory{
		BaseModel:     bun.BaseModel{},
		ID:            uuid.NewString(),
		PaymentID:     payment.ID,
		PaymentStatus: paymentStatus,
		StatusDetail:  "Updated as part of payment processing flow",
		CreateDate:    time.Now(),
	}
	_, err = act.PaymentsDB.NewInsert().Model(&paymentStatusHistory).Exec(context.Background())
	if err != nil {
		log.Errorf("Error inserting status update step: %v", err)
		return err
	}

	return nil
}
