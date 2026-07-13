package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/openreserveio/core/core-ledger/model"
	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

type BalanceOpts struct {
	AsOfDatetime *time.Time
	ForAccountID string
}

func GetBalance(ctx context.Context, db *bun.DB, opts BalanceOpts) (*model.AccountBalance, error) {

	var accountBalance *model.AccountBalance
	var err error

	switch opts.AsOfDatetime {

	case nil:
		accountBalance, err = getLatestBalance(ctx, db, opts.ForAccountID)

	default:
		accountBalance, err = getBalanceAsOfDate(ctx, db, opts.ForAccountID, *opts.AsOfDatetime)

	}

	if err != nil {
		log.Errorf("Error getting account balance: %v", err)
		return nil, err
	}

	return accountBalance, nil

}

func getLatestBalance(ctx context.Context, db *bun.DB, forAccountID string) (*model.AccountBalance, error) {

	var accountBalance model.AccountBalance
	err := db.NewSelect().
		Model(&accountBalance).
		Where("account_id = ?", forAccountID).
		Order("balance_date DESC").
		Limit(1).
		Scan(ctx, &accountBalance)
	if err != nil {
		log.Errorf("Unable to get latest balance for account ID %s: %v", forAccountID, err)
		return nil, err
	}
	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &accountBalance, nil

}

func getBalanceAsOfDate(ctx context.Context, db *bun.DB, forAccountID string, asOfDatetime time.Time) (*model.AccountBalance, error) {

	return getLatestBalance(ctx, db, forAccountID)
	// return nil, errors.New("UNIMPLEMENTED")

}
