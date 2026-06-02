package activities

import (
	"context"
	"encoding/xml"
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/moov-io/fednow20022/gen/pacs_008_001_08"
	"github.com/openreserveio/core/core-payments/pmtmodel"
	log "github.com/sirupsen/logrus"
)

func (act *PaymentActivity) ValidatePaymentInstruction(ctx context.Context, payment pmtmodel.Payment) (pmtmodel.Payment, error) {

	convertedPayment, err := convertFednowMessageToPayment(&payment)

	if convertedPayment.ID != "" {
		log.Infof("----> Payment is validated")

		// Store Update to Payment
		_, err = act.PaymentsDB.NewUpdate().
			Model(convertedPayment).
			Where("id = ?", convertedPayment.ID).
			Exec(ctx)
		if err != nil {
			return pmtmodel.Payment{}, err
		}

		return *convertedPayment, nil
	}

	return pmtmodel.Payment{}, errors.New("PaymentID should not be empty")

}

func convertFednowMessageToPayment(pmtMessage *pmtmodel.Payment) (*pmtmodel.Payment, error) {

	var fedNowMessage pacs_008_001_08.Document
	err := xml.Unmarshal(pmtMessage.PaymentMessage, &fedNowMessage)
	if err != nil {
		log.Errorf("Error unmarshalling FedNow Message: %v", err)
		return nil, err
	}

	sourceAmountFloat, err := strconv.ParseFloat(string(fedNowMessage.FIToFICstmrCdtTrf.CdtTrfTxInf[0].IntrBkSttlmAmt.Text), 64)
	if err != nil {
		log.Errorf("Error converting FedNow Message: %v", err)
		return nil, err
	}
	sourceAmount := int64(sourceAmountFloat * 100)

	settlementDate := time.Time(*fedNowMessage.FIToFICstmrCdtTrf.CdtTrfTxInf[0].IntrBkSttlmDt)

	pmt := pmtmodel.Payment{
		ID:                          pmtMessage.ID,
		PaymentNetworkID:            pmtmodel.PAYMENT_NETWORK_US_FEDNOW,
		ServiceSpecificID:           uuid.NewString(),
		NetworkIdentifier:           string(fedNowMessage.FIToFICstmrCdtTrf.CdtTrfTxInf[0].PmtId.EndToEndId),
		SourceCurrency:              string(fedNowMessage.FIToFICstmrCdtTrf.CdtTrfTxInf[0].IntrBkSttlmAmt.Ccy),
		TargetCurrency:              string(fedNowMessage.FIToFICstmrCdtTrf.CdtTrfTxInf[0].IntrBkSttlmAmt.Ccy),
		SourceAmount:                sourceAmount,
		TargetAmount:                sourceAmount,
		PaymentMessage:              pmtMessage.PaymentMessage,
		UltimateOriginatorEntityID:  "",
		UltimateBeneficiaryEntityID: "",
		CreateDate:                  time.Now(),
		EffectiveDate:               settlementDate,
		ModifyDate:                  time.Now(),
		IsBatch:                     false,
		ParentPaymentID:             "",
	}

	return &pmt, nil

}
