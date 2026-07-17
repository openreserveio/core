package activities

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/openreserveio/core/core-payments/pmtmodel"
	"github.com/openreserveio/core/core-util/bus"
	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

func (act *PaymentActivity) UpdatePaymentStatus(ctx context.Context, payment pmtmodel.Payment, paymentStatus string) error {

	var previousPaymentStatus string

	previousPaymentStatus = payment.CurrentPaymentStatus
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

	paymentStatusUpdateNotification := pmtmodel.PaymentStatusUpdateNotification{
		PaymentID:        payment.ID,
		NotificationDate: time.Now().String(),
		PreviousStatus:   previousPaymentStatus,
		CurrentStatus:    payment.CurrentPaymentStatus,
		AdditionalInfo:   "",
	}

	bus.Send(ctx, act.BusConn, fmt.Sprintf("%s.%s", "payments_service", "payment_status_update_notification"), &paymentStatusUpdateNotification)

	return nil
}
