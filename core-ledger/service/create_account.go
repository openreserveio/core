package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/openreserveio/core/core-ledger/model"
	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

func CreateAccount(ctx context.Context, db *bun.DB, ledgerId string, name string, code string, class string, metadata map[string]string, parentAccountId string, currency string) (*model.Account, error) {

	var parentAcctIdNullString sql.NullString
	if parentAccountId == "" {
		parentAcctIdNullString = sql.NullString{Valid: false}
	} else {
		parentAcctIdNullString = sql.NullString{String: parentAccountId, Valid: true}
	}

	acct := model.Account{
		ID:              uuid.NewString(),
		LedgerID:        ledgerId,
		Name:            name,
		Code:            code,
		Class:           class,
		Metadata:        ConvertMapStringToMapInterface(metadata),
		ParentAccountID: parentAcctIdNullString,
		Currency:        currency,
	}

	tx, _ := db.BeginTx(ctx, nil)
	_, err := tx.NewInsert().Model(&acct).Exec(ctx)
	if err != nil {
		tx.Rollback()
		log.Errorf("Unable to insert new account:  %v", err)
		return nil, err
	}

	// Create Account Balance
	acctBalance := model.AccountBalance{
		AccountID:                acct.ID,
		BalanceAsOfTransactionID: "",
		Balance:                  0,
		BalanceDate:              time.Now(),
	}
	_, err = tx.NewInsert().Model(&acctBalance).Exec(ctx)
	if err != nil {
		tx.Rollback()
		log.Errorf("Unable to insert new account:  %v", err)
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		log.Errorf("Error committing transaction: %v", err)
		return nil, err
	}

	return &acct, nil

}
