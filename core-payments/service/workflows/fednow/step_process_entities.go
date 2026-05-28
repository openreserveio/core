package fednow

import (
	"encoding/json"
	"encoding/xml"

	"github.com/moov-io/fednow20022/gen/pacs_008_001_08"
	"github.com/openreserveio/core/core-payments/generated/glmodel"
	"github.com/openreserveio/core/core-payments/pmtmodel"
	log "github.com/sirupsen/logrus"
)

func (f *FednowInboundPaymentWorkflow) ProcessEntities(data []byte, option map[string][]string) ([]byte, error) {

	log.Infof("ProcessEntities: %s", string(data))
	log.Infof("ProcessEntities Options: %v", option)

	// Unmarshall the original payment message
	// All of the entities we care about are in there
	var payment pmtmodel.Payment
	err := json.Unmarshal(data, &payment)
	if err != nil {
		log.Errorf("Error unmarshalling payment to process entities flow step: %v", err)
		return nil, err
	}

	var fednowMessage pacs_008_001_08.Document
	err = xml.Unmarshal([]byte(payment.PaymentMessage), &fednowMessage)
	if err != nil {
		log.Errorf("Error unmarshalling Fednow Payment Message to process entities flow step: %v", err)
		return nil, err
	}

	// Do Ultimate Beneficiary first
	if fednowMessage.FIToFICstmrCdtTrf.CdtTrfTxInf[0].UltmtCdtr.Nm == nil {
		log.Infof("No Ultimate Creditor Name")
	} else {
		ultimateBeneName := string(*fednowMessage.FIToFICstmrCdtTrf.CdtTrfTxInf[0].UltmtCdtr.Nm)
		entity := glmodel.Entity{}
		_ = ultimateBeneName
		_ = entity

		log.Infof("Ultimate Beneficiary Name: %s", ultimateBeneName)
	}

	// Ultimate originator
	if fednowMessage.FIToFICstmrCdtTrf.CdtTrfTxInf[0].UltmtDbtr.Nm == nil {
		log.Infof("No Ultimate Debtor Name")
	} else {
		ultimateOrigName := string(*fednowMessage.FIToFICstmrCdtTrf.CdtTrfTxInf[0].UltmtDbtr.Nm)
		entity := glmodel.Entity{}
		_ = ultimateOrigName
		_ = entity

		log.Infof("Ultimate Originator Name: %s", ultimateOrigName)
	}

	return nil, nil

}
