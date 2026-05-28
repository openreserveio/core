package service

import (
	"context"
	"fmt"

	"github.com/openreserveio/core/core-gl/bus"
	"github.com/openreserveio/core/core-gl/generated/glmodel"
	"github.com/openreserveio/core/core-gl/generated/model"
	glmodelint "github.com/openreserveio/core/core-gl/glmodel"
	log "github.com/sirupsen/logrus"
)

type FednowResult struct {
	TransactionID string
	Status        string
}

func PostFedNowPayment(ctx context.Context, busConn *bus.BusConnection, coreLedgerClient model.CoreLedgerServiceClient, ledgerId string, entry *glmodel.JournalEntry, paymentInfo *glmodel.CorePayment) (*FednowResult, error) {

	log.Infof("Posting Fednow Payment in GL!")
	metadata := map[string]string{
		glmodelint.MD_KEY_TRANSACTION_TYPE:      glmodelint.MD_TRANSACTION_TYPE_PAYMENT,
		glmodelint.MD_KEY_PAYMENT_CHANNEL:       glmodelint.MD_PAYMENT_CHANNEL_US_FEDNOW,
		glmodelint.MD_KEY_TRANSACTION_REFERENCE: paymentInfo.PaymentId,
		glmodelint.MD_KEY_PURPOSE:               entry.Purpose,
	}

	switch paymentInfo.Lifecycle {

	case glmodel.CorePayment_INITIAL_POSTING:
		metadata[glmodelint.MD_KEY_PAYMENT_LIFECYCLE] = glmodelint.MD_PAYMENT_LIFECYCLE_INITIAL_POSTING

	case glmodel.CorePayment_CLEARING_POSTING:
		metadata[glmodelint.MD_KEY_PAYMENT_LIFECYCLE] = glmodelint.MD_PAYMENT_LIFECYCLE_CLEARING_POSTING

	case glmodel.CorePayment_CUSTOMER_ACCOUNT_POSTING:
		metadata[glmodelint.MD_KEY_PAYMENT_LIFECYCLE] = glmodelint.MD_PAYMENT_LIFECYCLE_CUSTOMER_ACCOUNT_POSTING

	}

	var debits []*model.PostLedgerTransactionRequest_Entry
	var credits []*model.PostLedgerTransactionRequest_Entry

	for _, dbt := range entry.Debits {

		// get account ID
		res, err := coreLedgerClient.GetLedgerAccount(ctx, &model.GetLedgerAccountRequest{Code: dbt.AccountCode, LedgerId: ledgerId})
		if err != nil {
			log.Errorf("Error looking up account ID for account code %s: %v", dbt.AccountCode, err)
			return nil, err
		}
		if res.Status.Code != 200 {
			log.Errorf("Could not get account ID from account code %s: %v", dbt.AccountCode, res.Status.StatusMessage)
			return nil, fmt.Errorf(res.Status.StatusMessage)
		}

		entry := model.PostLedgerTransactionRequest_Entry{
			AccountId: res.AccountId,
			Amount:    dbt.Amount,
			Currency:  "",
			Metadata:  map[string]string{glmodelint.MD_KEY_NOTE: dbt.Note},
		}
		debits = append(debits, &entry)
	}

	for _, cdt := range entry.Credits {

		// get account ID
		res, err := coreLedgerClient.GetLedgerAccount(ctx, &model.GetLedgerAccountRequest{Code: cdt.AccountCode, LedgerId: ledgerId})
		if err != nil {
			log.Errorf("Error looking up account ID for account code %s: %v", cdt.AccountCode, err)
			return nil, err
		}
		if res.Status.Code != 200 {
			log.Errorf("Could not get account ID from account code %s: %v", cdt.AccountCode, res.Status.StatusMessage)
			return nil, fmt.Errorf(res.Status.StatusMessage)
		}

		entry := model.PostLedgerTransactionRequest_Entry{
			AccountId: res.AccountId,
			Amount:    cdt.Amount,
			Currency:  "",
			Metadata:  map[string]string{glmodelint.MD_KEY_NOTE: cdt.Note},
		}
		credits = append(credits, &entry)
	}

	request := model.PostLedgerTransactionRequest{
		LedgerId: ledgerId,
		Debits:   debits,
		Credits:  credits,
		Metadata: metadata,
	}
	resp, err := coreLedgerClient.PostLedgerTransaction(ctx, &request)
	if err != nil {
		log.Errorf("Error posting transaction: %v", err)
		return nil, err
	}

	if resp.Status.Code != 200 {
		log.Errorf("Problem occurred while posting transaction: %v", resp.Status.StatusMessage)
		return nil, fmt.Errorf(resp.Status.StatusMessage)
	}

	fednowResult := FednowResult{
		TransactionID: resp.LedgerTransactionId,
		Status:        STATUS_POSTED,
	}
	return &fednowResult, nil

}
