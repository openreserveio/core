package fednow

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/openreserveio/core/core-payments/generated/glmodel"
	"github.com/openreserveio/core/core-payments/pmtmodel"
	log "github.com/sirupsen/logrus"
)

func (f *FednowInboundPaymentWorkflow) InitialLedgerPostings(data []byte, option map[string][]string) ([]byte, error) {

	log.Infof("InitialLedgerPostings: %s", string(data))
	log.Infof("InitialLedgerPostings Options: %v", option)

	var payment pmtmodel.Payment
	err := json.Unmarshal(data, &payment)
	if err != nil {
		return nil, err
	}

	pmt := glmodel.CorePayment{
		Lifecycle: glmodel.CorePayment_INITIAL_POSTING,
		PaymentId: payment.ID,
	}

	postTransactionRequest := glmodel.PostTransactionRequest{
		LedgerId:            payment.LedgerID,
		TransactionType:     glmodel.PostTransactionRequest_US_PAYMENT_FEDNOW,
		CorePayment:         &pmt,
		UltimateOriginator:  nil,
		UltimateBeneficiary: nil,
		JournalEntry: &glmodel.JournalEntry{
			Debits: []*glmodel.JournalEntryItem{
				&glmodel.JournalEntryItem{
					AccountCode:    f.AccountingConfig.FednowSettlementAccountCode,
					Amount:         payment.TargetAmount,
					Note:           fmt.Sprintf("Initial Posting for FedNow Payment ID: %s", payment.ID),
					TaggedEntityId: nil,
				},
			},
			Credits: []*glmodel.JournalEntryItem{
				&glmodel.JournalEntryItem{
					AccountCode:    f.AccountingConfig.FednowSettlementInProgressAccountCode,
					Amount:         payment.TargetAmount,
					Note:           fmt.Sprintf("Initial Posting for FedNow Payment ID: %s", payment.ID),
					TaggedEntityId: nil,
				},
			},
			Purpose:        fmt.Sprintf("FedNow Inbound Payment ID: %s", payment.ID),
			PosterEntityId: "",
			TaggedEntityId: nil,
		},
	}

	log.Infof("Posting transaction to GL")
	postTxResponse, err := f.CoreGL.PostTransaction(context.Background(), &postTransactionRequest)
	if err != nil {
		log.Errorf("Error posting transaction to GL: %v", err)
		return nil, err
	}

	if postTxResponse.Status.Code != http.StatusOK {
		return nil, fmt.Errorf("Problem posting transaction to GL: %s", postTxResponse.Status.StatusMessage)
	}

	log.Infof("Transaction posted to GL")
	return data, nil
}
