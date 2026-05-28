package service

import (
	"context"
	"database/sql"

	"github.com/openreserveio/core/core-ledger/model"
	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

func GetAccount(ctx context.Context, db *bun.DB, ledgerId string, accountId string, code string) (*model.Account, error) {

	var account model.Account
	if accountId != "" {
		err := db.NewSelect().Model(&model.Account{}).Where("account_id = ? and ledger_id = ?", accountId, ledgerId).Scan(ctx, &account)
		if err != nil && err != sql.ErrNoRows {
			log.Errorf("Error while retrieving account from EntityDB by account ID:  %v", err)
			return nil, err
		}
	} else if code != "" {
		err := db.NewSelect().Model(&model.Account{}).Where("ledger_id = ? and code = ?", ledgerId, code).Scan(ctx, &account)
		if err != nil && err != sql.ErrNoRows {
			log.Errorf("Error while retrieving account from EntityDB by Code:  %v", err)
			return nil, err
		}
	}

	return &account, nil

}
