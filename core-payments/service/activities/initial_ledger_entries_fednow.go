package activities

import (
	"context"
	"fmt"

	"github.com/openreserveio/core/core-payments/generated/glmodel"
	"github.com/openreserveio/core/core-payments/pmtmodel"
)

func (act *PaymentActivity) FednowInitialLedgerEntries(ctx context.Context, payment pmtmodel.Payment, accountingConfig pmtmodel.AccountingConfig) error {

	// Initial ledger entry for fednow and fedwires, RTGS which realtime credit the settlement account
	// Debit the settlement account, credit the clearing account
	resp, err := act.CoreGLClient.PostTransaction(ctx, &glmodel.PostTransactionRequest{
		LedgerId:        accountingConfig.LedgerID,
		TransactionType: glmodel.PostTransactionRequest_US_PAYMENT_FEDNOW,
		JournalEntry: &glmodel.JournalEntry{
			Debits: []*glmodel.JournalEntryItem{
				&glmodel.JournalEntryItem{
					AccountCode: accountingConfig.FednowSettlementAccountCode,
					Amount:      payment.TargetAmount,
					Note:        fmt.Sprintf("FEDNOW PMT %s", payment.ID),
				},
			},
			Credits: []*glmodel.JournalEntryItem{
				&glmodel.JournalEntryItem{
					AccountCode: accountingConfig.FednowClearingAccountCode,
					Amount:      payment.TargetAmount,
					Note:        fmt.Sprintf("FEDNOW PMT %s", payment.ID),
				},
			},
			Purpose:        fmt.Sprintf("FEDNOW PMT %s", payment.ID),
			PosterEntityId: "",
			TaggedEntityId: nil,
		},
		CorePayment: &glmodel.CorePayment{
			Lifecycle: glmodel.CorePayment_INITIAL_POSTING,
			PaymentId: payment.ID,
		},
		UltimateOriginator:  nil,
		UltimateBeneficiary: nil,
	})

	if err != nil {
		return err
	}
	if resp.Status.Code != 200 {
		return fmt.Errorf("Error posting initial ledger entry: %v", resp.Status.StatusMessage)
	}

	return nil
}
