package activities

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/openreserveio/core/core-payments/pmtmodel"
	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

func ValidatePaymentInstruction(ctx context.Context, db *bun.DB, payment pmtmodel.Payment) (pmtmodel.Payment, error) {

	err := json.Unmarshal(payment.PaymentMessage, &payment)
	if err != nil {
		return pmtmodel.Payment{}, err
	}

	if payment.ID != "" {
		log.Infof("----> Payment is validated")

		// Store Update to Payment
		_, err = db.NewUpdate().
			Model(&payment).
			Exec(ctx)
		if err != nil {
			return pmtmodel.Payment{}, err
		}

		return payment, nil
	}

	return pmtmodel.Payment{}, errors.New("PaymentID should not be empty")

}
