package fednow

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/openreserveio/core/core-payments/pmtmodel"
	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

func (f *FednowInboundPaymentWorkflow) CreateStatusUpdateFlowStep(status string) func(data []byte, option map[string][]string) ([]byte, error) {

	return func(data []byte, option map[string][]string) ([]byte, error) {

		log.Infof("StatusUpdate: %s", string(data))
		log.Infof("StatusUpdate Options: %v", option)

		var payment pmtmodel.Payment
		err := json.Unmarshal(data, &payment)
		if err != nil {
			log.Errorf("Error unmarshalling status update flow step: %v", err)
			return nil, err
		}

		payment.CurrentPaymentStatus = status
		_, err = f.PaymentsDB.NewUpdate().Model(&payment).Where("id = ?", payment.ID).Exec(context.Background())
		if err != nil {
			log.Errorf("Error updating status update flow step: %v", err)
			return nil, err
		}

		paymentStatusHistory := pmtmodel.PaymentStatusHistory{
			BaseModel:     bun.BaseModel{},
			ID:            uuid.NewString(),
			PaymentID:     payment.ID,
			PaymentStatus: status,
			StatusDetail:  "Updated as part of payment processing flow",
			CreateDate:    time.Now(),
		}
		_, err = f.PaymentsDB.NewInsert().Model(&paymentStatusHistory).Exec(context.Background())
		if err != nil {
			log.Errorf("Error inserting status update flow step: %v", err)
			return nil, err
		}

		data, err = json.Marshal(payment)
		if err != nil {
			log.Errorf("Error marshalling payment status update flow step: %v", err)
			return nil, err
		}

		return data, nil
	}

}
