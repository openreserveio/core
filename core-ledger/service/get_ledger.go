package service

import (
	"context"
	"database/sql"

	"github.com/openreserveio/core/core-ledger/model"
	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

func GetLedger(ctx context.Context, db *bun.DB, ledgerId string) (*model.Ledger, error) {

	var ledger model.Ledger
	err := db.NewSelect().Model(&ledger).Where("ledger_id = ?", ledgerId).Scan(ctx, &ledger)
	if err != nil && err != sql.ErrNoRows {
		log.Errorf("GetLedger db scan error: %s", err)
		return nil, err
	}

	return &ledger, nil

}
