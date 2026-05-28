package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/openreserveio/core/core-ledger/model"
	"github.com/uptrace/bun"
)

func CreateLedger(ctx context.Context, dbBun *bun.DB, ledgerDefinition *model.Ledger) (*model.Ledger, error) {

	ledgerDefinition.ID = uuid.NewString()
	_, err := dbBun.NewInsert().Model(ledgerDefinition).Exec(ctx)
	if err != nil {
		return nil, err
	}

	return ledgerDefinition, nil

}
