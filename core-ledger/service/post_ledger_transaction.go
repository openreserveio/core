package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/openreserveio/core/core-ledger/model"
	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

func PostLedgerTransaction(ctx context.Context, db *bun.DB, ledgerId string, debits []model.LedgerTransactionEntry, credits []model.LedgerTransactionEntry, metadata map[string]interface{}) (*model.LedgerTransaction, []*model.AccountBalance, error) {

	err := validateEntries(debits, credits)
	if err != nil {
		return nil, nil, err
	}

	ledgerTx := model.LedgerTransaction{
		ID:              uuid.NewString(),
		LedgerID:        ledgerId,
		Metadata:        metadata,
		Debits:          debits,
		Credits:         credits,
		TransactionDate: time.Now(),
	}

	var updatedBalances []*model.AccountBalance
	err = db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {

		// Create Ledger Transaction
		err = persistLedgerTransaction(ctx, tx, &ledgerTx)
		if err != nil {
			return err
		}

		// Create Entries
		err = persistLedgerTransactionEntries(ctx, tx, &ledgerTx)
		if err != nil {
			return err
		}

		// Recalculate Balances
		updatedBalances, err = recalculateBalances(ctx, tx, &ledgerTx)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		log.Errorf("Error posting ledger transaction: %v", err)
		return nil, nil, err
	}

	return &ledgerTx, updatedBalances, nil

}

func validateEntries(debits []model.LedgerTransactionEntry, credits []model.LedgerTransactionEntry) error {

	// Debits must equal Credits
	var totalDebits int64 = 0
	var totalCredits int64 = 0

	for _, entry := range debits {
		totalDebits += entry.Amount
	}

	for _, entry := range credits {
		totalCredits += entry.Amount
	}

	if totalDebits != totalCredits {
		return errors.New("total debits does not match total credits")
	}

	return nil

}

func persistLedgerTransaction(ctx context.Context, tx bun.Tx, transaction *model.LedgerTransaction) error {

	_, err := tx.NewInsert().Model(transaction).Exec(ctx)
	if err != nil {
		log.Errorf("Error inserting ledger transaction into database: %v", err)
		return err
	}

	return nil
}

func persistLedgerTransactionEntries(ctx context.Context, tx bun.Tx, transaction *model.LedgerTransaction) error {

	for _, entry := range transaction.Debits {
		entry.LedgerTransactionID = transaction.ID
		entry.TransactionEntryType = model.TX_ENTRY_TYPE_DEBIT
		_, err := tx.NewInsert().Model(&entry).Exec(ctx)
		if err != nil {
			log.Errorf("Error inserting ledger transaction entry into database: %v", err)
			return err
		}
	}

	for _, entry := range transaction.Credits {
		entry.LedgerTransactionID = transaction.ID
		entry.TransactionEntryType = model.TX_ENTRY_TYPE_CREDIT
		_, err := tx.NewInsert().Model(&entry).Exec(ctx)
		if err != nil {
			log.Errorf("Error inserting ledger transaction entry into database: %v", err)
			return err
		}
	}

	return nil

}

func recalculateBalances(ctx context.Context, db bun.Tx, transaction *model.LedgerTransaction) ([]*model.AccountBalance, error) {

	// get current balances, if any
	var updatedBalances []*model.AccountBalance

	for _, entry := range transaction.Debits {

		recalcedBalances, err := recalculateAccountBalances(ctx, db, transaction.LedgerID, model.TX_ENTRY_TYPE_DEBIT, entry.AccountID, entry.Amount, transaction.ID)
		if err != nil {
			return nil, err
		}
		updatedBalances = append(updatedBalances, recalcedBalances...)

	}

	for _, entry := range transaction.Credits {

		recalcedBalances, err := recalculateAccountBalances(ctx, db, transaction.LedgerID, model.TX_ENTRY_TYPE_CREDIT, entry.AccountID, entry.Amount, transaction.ID)
		if err != nil {
			return nil, err
		}
		updatedBalances = append(updatedBalances, recalcedBalances...)

	}

	return updatedBalances, nil

}

func recalculateAccountBalances(ctx context.Context, db bun.Tx, ledgerId string, txDebitOrCredit string, accountId string, txAmount int64, txId string) ([]*model.AccountBalance, error) {

	// updated balances
	var updatedBalances []*model.AccountBalance

	// get account and account balance
	var ledgerAccount model.Account
	err := db.NewSelect().Model(&model.Account{}).Where("account_id = ? and ledger_id = ?", accountId, ledgerId).Scan(ctx, &ledgerAccount)
	if err != nil && err != sql.ErrNoRows {
		log.Errorf("Error scanning ledger (%s) account (%s): %v", ledgerId, accountId, err)
		return nil, err
	}
	if err == sql.ErrNoRows {
		log.Errorf("Account (%s) not found in ledger (%s)", accountId, ledgerId)
		return nil, fmt.Errorf("account (%s) not found in ledger (%s)", accountId, ledgerId)
	}

	var accountBalance model.AccountBalance
	err = db.NewSelect().Model(&model.AccountBalance{}).Where("account_id = ?", accountId).Order("balance_date DESC").Scan(ctx, &accountBalance)
	if err == sql.ErrNoRows {

		// This is a new account balance! Start fresh!
		accountBalance.AccountID = accountId
		accountBalance.Balance = 0

	}

	// adjust balance based on class of account
	if txDebitOrCredit == model.TX_ENTRY_TYPE_DEBIT {

		switch ledgerAccount.Class {
		case model.ACCOUNT_CLASS_ASSET:
			accountBalance.Balance += txAmount

		case model.ACCOUNT_CLASS_LIABILITY:
			accountBalance.Balance -= txAmount

		case model.ACCOUNT_CLASS_EQUITY:
			accountBalance.Balance -= txAmount

		case model.ACCOUNT_CLASS_INCOME:
			accountBalance.Balance -= txAmount

		case model.ACCOUNT_CLASS_EXPENSE:
			accountBalance.Balance += txAmount
		}

	} else if txDebitOrCredit == model.TX_ENTRY_TYPE_CREDIT {

		switch ledgerAccount.Class {
		case model.ACCOUNT_CLASS_ASSET:
			accountBalance.Balance -= txAmount

		case model.ACCOUNT_CLASS_LIABILITY:
			accountBalance.Balance += txAmount

		case model.ACCOUNT_CLASS_EQUITY:
			accountBalance.Balance += txAmount

		case model.ACCOUNT_CLASS_INCOME:
			accountBalance.Balance += txAmount

		case model.ACCOUNT_CLASS_EXPENSE:
			accountBalance.Balance -= txAmount
		}

	} else {

		return nil, fmt.Errorf("invalid transaction type, must be DEBIT or CREDIT: %s", txDebitOrCredit)

	}

	// adjust dates and latest TXs
	accountBalance.BalanceDate = time.Now()
	accountBalance.BalanceAsOfTransactionID = txId

	// Insert into balances table
	_, err = db.NewInsert().Model(&accountBalance).Exec(ctx)
	if err != nil {
		log.Errorf("Error inserting ledger account balance: %v", err)
		return nil, err
	}

	// If there are parent accounts, recalculate their balances
	updatedBalances = append(updatedBalances, &accountBalance)
	if ledgerAccount.ParentAccountID.String != "" {
		updatesBalancesParents, err := recalculateAccountBalances(ctx, db, ledgerId, txDebitOrCredit, ledgerAccount.ParentAccountID.String, txAmount, txId)
		if err != nil {
			return nil, err
		}
		updatedBalances = append(updatedBalances, updatesBalancesParents...)
	}

	return updatedBalances, nil

}
