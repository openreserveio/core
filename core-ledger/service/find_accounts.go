package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/openreserveio/core/core-ledger/model"
	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

func FindAllAccountsInLedger(ctx context.Context, db *bun.DB, ledgerId string) ([]*model.Account, error) {

	var accounts []*model.Account
	err := db.NewSelect().Model(&accounts).Where("ledger_id = ?", ledgerId).Scan(ctx, &accounts)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		log.Errorf("Error scanning accounts (FindAllAccountsInLedger): %v", err)
		return nil, err
	}

	return accounts, nil

}

func FindAccountsByClass(ctx context.Context, db *bun.DB, ledgerId string, accountClass string) ([]*model.Account, error) {

	var accounts []*model.Account
	err := db.NewSelect().Model(&accounts).Where("ledger_id = ? AND class = ?", ledgerId, accountClass).Scan(ctx, &accounts)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		log.Errorf("Error scanning accounts (FindAllAccountsInLedger): %v", err)
		return nil, err
	}

	return accounts, nil
}

func FindAccountsByMetadata(ctx context.Context, db *bun.DB, ledgerId string, metadataCriteria map[string]string) ([]*model.Account, error) {

	if len(metadataCriteria) == 0 {
		log.Infof("Zero Metadata Criteria, returning empty slice")
		return nil, nil
	}

	// construct a query where we look for JSON metadata that matches the criteria
	var metadataQuery string
	for key, value := range metadataCriteria {
		metadataQuery = metadataQuery + fmt.Sprintf("\"%s\": \"%s\", ", key, value)
	}
	// remove the trailing comma
	metadataQuery = metadataQuery[:len(metadataQuery)-2]
	metadataQuery = fmt.Sprintf("metadata @> '{%s}'", metadataQuery)

	var accounts []*model.Account
	err := db.NewSelect().Model(&accounts).Where("ledger_id = ? and "+metadataQuery, ledgerId).Scan(ctx, &accounts)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		log.Errorf("Error scanning accounts (FindByMetadata): %v", err)
		return nil, err
	}

	return accounts, nil

}
