package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/openreserveio/core/core-gl/generated/glmodel"
	glmodelint "github.com/openreserveio/core/core-gl/glmodel"
	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

func CreateEntity(ctx context.Context, db *bun.DB, ledgerEntity *glmodel.LedgerEntity) error {

	mailingAddress := glmodelint.EntityAddress{
		ID:         uuid.NewString(),
		RawAddress: ledgerEntity.MailingAddress.RawAddress,
		CreateDate: time.Now(),
		ParsedDate: time.Now(),
	}

	businessAddress := glmodelint.EntityAddress{
		ID:         uuid.NewString(),
		RawAddress: ledgerEntity.BusinessAddress.RawAddress,
		CreateDate: time.Now(),
		ParsedDate: time.Now(),
	}

	entityName := glmodelint.EntityName{
		ID:                  uuid.NewString(),
		IndividualSurName:   ledgerEntity.EntityName.IndividualSurName,
		IndividualGivenName: ledgerEntity.EntityName.IndividualGivenName,
		CreateDate:          time.Now(),
	}

	entity := glmodelint.Entity{
		ID:                   uuid.NewString(),
		EntityType:           ledgerEntity.EntityType.String(),
		EntityNameID:         entityName.ID,
		MailingAddressID:     mailingAddress.ID,
		BusinessAddressID:    businessAddress.ID,
		LatestVerificationID: "",
		SourceType:           "",
		SourceID:             "",
		Remapped:             false,
		RemappedToEntityID:   "",
		CreateDate:           time.Now(),
		UpdateDate:           time.Now(),
	}

	// Store Components then entity
	db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {

		_, err := tx.NewInsert().Model(&mailingAddress).Exec(ctx)
		if err != nil {
			log.Errorf("Error creating mailing address: %v", err)
			return err
		}

		_, err = tx.NewInsert().Model(&businessAddress).Exec(ctx)
		if err != nil {
			log.Errorf("Error creating business address: %v", err)
			return err
		}

		_, err = tx.NewInsert().Model(&entityName).Exec(ctx)
		if err != nil {
			log.Errorf("Error creating entity name: %v", err)
			return err
		}

		_, err = tx.NewInsert().Model(&entity).Exec(ctx)
		if err != nil {
			log.Errorf("Error creating entity: %v", err)
			return err
		}

		return nil
	})

	ledgerEntity.EntityId = entity.ID
	ledgerEntity.EntityName.EntityNameId = entityName.ID
	ledgerEntity.MailingAddress.EntityAddressId = mailingAddress.ID
	ledgerEntity.BusinessAddress.EntityAddressId = businessAddress.ID

	return nil
}
