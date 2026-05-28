package service

import (
	"context"
	"net/http"

	"fmt"
	"time"

	"github.com/openreserveio/core/core-gl/application"
	"github.com/openreserveio/core/core-gl/bus"
	"github.com/openreserveio/core/core-gl/generated/glmodel"
	"github.com/openreserveio/core/core-gl/generated/model"
	glmodelint "github.com/openreserveio/core/core-gl/glmodel"
	log "github.com/sirupsen/logrus"
)

func PostJournalEntry(ctx context.Context, busConn *bus.BusConnection, coreLedgerClient model.CoreLedgerServiceClient, ledgerId string, journalEntry *glmodel.JournalEntry) (string, error) {

	metadata := map[string]string{
		glmodelint.MD_TRANSACTION_TYPE_JOURNAL_ENTRY: glmodelint.MD_TRANSACTION_TYPE_JOURNAL_ENTRY,
		glmodelint.MD_KEY_TRANSACTION_REFERENCE:      "reference here 012334",
	}

	request := model.PostLedgerTransactionRequest{
		LedgerId: ledgerId,
		Debits:   nil,
		Credits:  nil,
		Metadata: metadata,
	}

	// process debits
	for _, debit := range journalEntry.Debits {

		// Lookup Account IDs
		debitAccountId, err := lookupAccountIdByAccountCode(ctx, coreLedgerClient, ledgerId, debit.AccountCode)
		if err != nil {
			return "", fmt.Errorf("Error looking up account ID for account code %s: %v", debit.AccountCode, err)
		}
		if debitAccountId == "" {
			return "", fmt.Errorf("Account code (%s) not found", debit.AccountCode)
		}

		debitEntry := model.PostLedgerTransactionRequest_Entry{
			AccountId: debitAccountId,
			Amount:    debit.Amount,
			Currency:  "",
			Metadata:  nil,
		}
		request.Debits = append(request.Debits, &debitEntry)

	}

	// process credits
	for _, credit := range journalEntry.Credits {

		// Lookup Account IDs
		creditAccountId, err := lookupAccountIdByAccountCode(ctx, coreLedgerClient, ledgerId, credit.AccountCode)
		if err != nil {
			return "", fmt.Errorf("Error looking up account ID for account code %s: %v", credit.AccountCode, err)
		}
		if creditAccountId == "" {
			return "", fmt.Errorf("Account code (%s) not found", credit.AccountCode)
		}

		creditEntry := model.PostLedgerTransactionRequest_Entry{
			AccountId: creditAccountId,
			Amount:    credit.Amount,
			Currency:  "",
			Metadata:  nil,
		}
		request.Credits = append(request.Credits, &creditEntry)

	}

	// Post the transaction - this goes to the bus
	response := model.PostLedgerTransactionResponse{}
	err := bus.SendForReply(busConn, 10*time.Second, fmt.Sprintf("%s.%s", application.SERVICE_NAME_CORE_LEDGER_POSTER, application.SERVICE_ENDPOINT_POST_TRANSACTION), &request, &response)
	if err != nil {
		log.Errorf("Error posting journal entry: %v", err)
		return "", err
	}
	if response.Status.Code != http.StatusOK {
		log.Errorf("Posting journal entry resulted in error: %v", response.Status.StatusMessage)
		return "", fmt.Errorf("Posting journal entry resulted in error: %v", response.Status.StatusMessage)
	}

	return response.LedgerTransactionId, nil

}

func lookupAccountIdByAccountCode(ctx context.Context, client model.CoreLedgerServiceClient, ledgerId string, accountCode string) (string, error) {

	resp, err := client.GetLedgerAccount(ctx, &model.GetLedgerAccountRequest{Code: accountCode, LedgerId: ledgerId})
	if err != nil {
		return "", err
	}

	return resp.GetAccountId(), nil

}
