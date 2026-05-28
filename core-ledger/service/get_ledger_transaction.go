package service

import (
	"context"
	"database/sql"

	"github.com/openreserveio/core/core-ledger/model"
	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

func GetLedgerTransaction(ctx context.Context, db *bun.DB, ledgerId string, ledgerTransactionId string) (*model.LedgerTransaction, error) {

	var ledgerTx model.LedgerTransaction

	// Get the base transaction
	err := db.NewSelect().Model(&ledgerTx).Where("ledger_transaction_id = ? and ledger_id = ?", ledgerTransactionId, ledgerId).Scan(ctx, &ledgerTx)
	if err != nil && err != sql.ErrNoRows {
		log.Errorf("GetLedgerTransaction error: %v", err)
		return nil, err
	}
	if err == sql.ErrNoRows {
		return nil, nil
	}

	// Get the entries (debits and credits!)
	var entries []model.LedgerTransactionEntry
	var debits []model.LedgerTransactionEntry
	var credits []model.LedgerTransactionEntry
	err = db.NewSelect().Model(&model.LedgerTransactionEntry{}).Where("ledger_transaction_id = ?", ledgerTransactionId).Scan(ctx, &entries)
	if err != nil && err != sql.ErrNoRows {
		log.Errorf("Get LedgerTransactionEntry error: %v", err)
		return nil, err
	}
	if err == sql.ErrNoRows {
		log.Warnf("No entries found for ledger transaction id %s.  This is unusual.", ledgerTransactionId)
	}

	for _, entry := range entries {
		switch entry.TransactionEntryType {

		case model.TX_ENTRY_TYPE_DEBIT:
			debits = append(debits, entry)

		case model.TX_ENTRY_TYPE_CREDIT:
			credits = append(credits, entry)
		}
	}

	ledgerTx.Debits = debits
	ledgerTx.Credits = credits

	return &ledgerTx, nil

}
