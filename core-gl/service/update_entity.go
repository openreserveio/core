package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/openreserveio/core/core-gl/generated/glmodel"
	glmodelint "github.com/openreserveio/core/core-gl/glmodel"
	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

func UpdateEntity(ctx context.Context, db *bun.DB, ledgerEntity *glmodel.LedgerEntity) (*glmodel.LedgerEntity, error) {

	var existing glmodelint.Entity
	err := db.NewSelect().Model(&existing).Where("id = ?", ledgerEntity.EntityId).Scan(ctx, &existing)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		log.Errorf("UpdateEntity db scan error: %s", err)
		return nil, err
	}

	now := time.Now()

	err = db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {

		if ledgerEntity.EntityName != nil {
			_, err := tx.NewUpdate().
				TableExpr("entity_name").
				Set("individual_given_name = ?", ledgerEntity.EntityName.IndividualGivenName).
				Set("individual_sur_name = ?", ledgerEntity.EntityName.IndividualSurName).
				Set("individual_middle_name = ?", ledgerEntity.EntityName.IndividualMiddleName).
				Set("us_sole_proprietorship_name = ?", ledgerEntity.EntityName.UsSoleProprietorshipName).
				Set("us_llc_name = ?", ledgerEntity.EntityName.UsLLCName).
				Set("us_corporation_name = ?", ledgerEntity.EntityName.UsCorporateName).
				Set("us_government_agency_name = ?", ledgerEntity.EntityName.UsGovernmentAgencyName).
				Set("foreign_corporation_name = ?", ledgerEntity.EntityName.ForeignCorporationName).
				Where("id = ?", existing.EntityNameID).
				Exec(ctx)
			if err != nil {
				log.Errorf("UpdateEntity entity_name update error: %s", err)
				return err
			}
		}

		if ledgerEntity.MailingAddress != nil {
			mailingAddr := protoToEntityAddress(existing.MailingAddressID, ledgerEntity.MailingAddress)
			_, err := tx.NewUpdate().Model(&mailingAddr).WherePK().Exec(ctx)
			if err != nil {
				log.Errorf("UpdateEntity mailing_address update error: %s", err)
				return err
			}
		}

		if ledgerEntity.BusinessAddress != nil {
			businessAddr := protoToEntityAddress(existing.BusinessAddressID, ledgerEntity.BusinessAddress)
			_, err := tx.NewUpdate().Model(&businessAddr).WherePK().Exec(ctx)
			if err != nil {
				log.Errorf("UpdateEntity business_address update error: %s", err)
				return err
			}
		}

		_, err = tx.NewUpdate().
			TableExpr("entity").
			Set("entity_type = ?", ledgerEntity.EntityType.String()).
			Set("update_date = ?", now).
			Where("id = ?", existing.ID).
			Exec(ctx)
		if err != nil {
			log.Errorf("UpdateEntity entity update error: %s", err)
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return GetEntity(ctx, db, existing.ID)
}

func protoToEntityAddress(id string, src *glmodel.LedgerEntityAddress) glmodelint.EntityAddress {
	return glmodelint.EntityAddress{
		ID:            id,
		RawAddress:    src.RawAddress,
		House:         src.House,
		Category:      src.Category,
		Near:          src.Near,
		HouseNumber:   src.HouseNumber,
		Road:          src.Road,
		Unit:          src.Unit,
		Level:         src.Level,
		Staircase:     src.Staircase,
		Entrance:      src.Entrance,
		POBox:         src.PoBox,
		Postcode:      src.Postcode,
		Suburb:        src.Suburb,
		CityDistrict:  src.CityDistrict,
		City:          src.City,
		Island:        src.Island,
		StateDistrict: src.StateDistrict,
		State:         src.State,
		CountryRegion: src.CountryRegion,
		Country:       src.Country,
		WorldRegion:   src.WorldRegion,
	}
}
