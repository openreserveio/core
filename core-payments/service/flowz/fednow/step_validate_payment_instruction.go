package fednow

import (
	"encoding/json"
	"errors"

	"github.com/openreserveio/core/core-payments/pmtmodel"
	log "github.com/sirupsen/logrus"
)

func (f *FednowInboundPaymentWorkflow) ValidatePaymentInstruction(data []byte, option map[string][]string) ([]byte, error) {

	log.Infof("ValidatePaymentInstruction: %s", string(data))
	log.Infof("ValidatePaymentInstruction Options: %v", option)

	var payment pmtmodel.Payment
	err := json.Unmarshal(data, &payment)
	if err != nil {
		return nil, err
	}

	if payment.ID != "" {
		log.Infof("----> Payment is validated")
		return data, nil
	}

	return data, errors.New("PaymentID should not be empty")

}
