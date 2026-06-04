package activities

import (
	"context"
	"fmt"

	"github.com/openreserveio/core/core-payments/generated/glmodel"
	"github.com/openreserveio/core/core-payments/pmtmodel"
)

func (act *SuspendPaymentActivity) FednowSuspendPaymentLedgerEntries(ctx context.Context, payment pmtmodel.Payment, accountingConfig pmtmodel.AccountingConfig) error {

	res, err := act.GLServiceClient.PostTransaction(ctx, &glmodel.PostTransactionRequest{
		LedgerId:        accountingConfig.LedgerID,
		TransactionType: glmodel.PostTransactionRequest_PAYMENT_SUSPENSE,
		JournalEntry: &glmodel.JournalEntry{
			Debits: []*glmodel.JournalEntryItem{
				&glmodel.JournalEntryItem{
					AccountCode: accountingConfig.FednowClearingAccountCode,
					Amount:      payment.TargetAmount,
					Note:        fmt.Sprintf("SUSPENDED FEDNOW PMT %s", payment.ID),
				},
			},
			Credits: []*glmodel.JournalEntryItem{
				&glmodel.JournalEntryItem{
					AccountCode: accountingConfig.FednowSuspenseAccountCode,
					Amount:      payment.TargetAmount,
					Note:        fmt.Sprintf("SUSPENDED FEDNOW PMT %s", payment.ID),
				},
			},
			Purpose:        "FedNow Payment to Suspense for Review",
			PosterEntityId: "",
			TaggedEntityId: []string{payment.UltimateBeneficiaryEntityID, payment.UltimateOriginatorEntityID},
		},
		CorePayment: &glmodel.CorePayment{
			PaymentId: payment.ID,
		},
		UltimateOriginator:  &glmodel.LedgerEntity{EntityId: payment.UltimateOriginatorEntityID},
		UltimateBeneficiary: &glmodel.LedgerEntity{EntityId: payment.UltimateBeneficiaryEntityID},
	})
	if err != nil {
		return err
	}
	if res.Status.Code != 200 {
		return fmt.Errorf("Error posting suspension ledger entry: %v", res.Status.StatusMessage)
	}

	return nil

}
