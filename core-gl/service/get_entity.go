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

func GetEntity(ctx context.Context, db *bun.DB, entityId string) (*glmodel.LedgerEntity, error) {

	var entity glmodelint.Entity
	err := db.NewSelect().Model(&entity).Where("id = ?", entityId).Scan(ctx, &entity)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		log.Errorf("GetEntity db scan error: %s", err)
		return nil, err
	}

	var entityName glmodelint.EntityName
	err = db.NewSelect().Model(&entityName).Where("id = ?", entity.EntityNameID).Scan(ctx, &entityName)
	if err != nil && err != sql.ErrNoRows {
		log.Errorf("GetEntity entity_name scan error: %s", err)
		return nil, err
	}

	var mailingAddress glmodelint.EntityAddress
	err = db.NewSelect().Model(&mailingAddress).Where("id = ?", entity.MailingAddressID).Scan(ctx, &mailingAddress)
	if err != nil && err != sql.ErrNoRows {
		log.Errorf("GetEntity mailing_address scan error: %s", err)
		return nil, err
	}

	var businessAddress glmodelint.EntityAddress
	err = db.NewSelect().Model(&businessAddress).Where("id = ?", entity.BusinessAddressID).Scan(ctx, &businessAddress)
	if err != nil && err != sql.ErrNoRows {
		log.Errorf("GetEntity business_address scan error: %s", err)
		return nil, err
	}

	verificationStatus := glmodelint.VERIFICATION_STATUS_PENDING
	if entity.LatestVerificationID != "" {
		var verification glmodelint.EntityVerification
		err = db.NewSelect().Model(&verification).Where("id = ?", entity.LatestVerificationID).Scan(ctx, &verification)
		if err != nil && err != sql.ErrNoRows {
			log.Errorf("GetEntity entity_verification scan error: %s", err)
			return nil, err
		}
		if err == nil {
			verificationStatus = verification.VerificationStatus
		}
	}

	entityTypeVal := glmodel.LedgerEntity_EntityType_value[entity.EntityType]

	ledgerEntity := &glmodel.LedgerEntity{
		EntityId:   entity.ID,
		EntityType: glmodel.LedgerEntity_EntityType(entityTypeVal),
		EntityName: &glmodel.LedgerEntityName{
			EntityNameId:             entityName.ID,
			IndividualGivenName:      entityName.IndividualGivenName,
			IndividualSurName:        entityName.IndividualSurName,
			IndividualMiddleName:     entityName.IndividualMiddleName,
			UsSoleProprietorshipName: entityName.USSoleProprietorshipName,
			UsLLCName:                entityName.USLLCName,
			UsCorporateName:          entityName.USCorporationName,
			UsGovernmentAgencyName:   entityName.USGovernmentAgencyName,
			ForeignCorporationName:   entityName.ForeignCorporationName,
			CreateDate:               entityName.CreateDate.Format(time.RFC3339),
		},
		MailingAddress: &glmodel.LedgerEntityAddress{
			EntityAddressId: mailingAddress.ID,
			RawAddress:      mailingAddress.RawAddress,
			House:           mailingAddress.House,
			Category:        mailingAddress.Category,
			Near:            mailingAddress.Near,
			HouseNumber:     mailingAddress.HouseNumber,
			Road:            mailingAddress.Road,
			Unit:            mailingAddress.Unit,
			Level:           mailingAddress.Level,
			Staircase:       mailingAddress.Staircase,
			Entrance:        mailingAddress.Entrance,
			PoBox:           mailingAddress.POBox,
			Postcode:        mailingAddress.Postcode,
			Suburb:          mailingAddress.Suburb,
			CityDistrict:    mailingAddress.CityDistrict,
			City:            mailingAddress.City,
			Island:          mailingAddress.Island,
			StateDistrict:   mailingAddress.StateDistrict,
			State:           mailingAddress.State,
			CountryRegion:   mailingAddress.CountryRegion,
			Country:         mailingAddress.Country,
			WorldRegion:     mailingAddress.WorldRegion,
			CreateDate:      mailingAddress.CreateDate.String(),
			ParsedDate:      mailingAddress.ParsedDate.String(),
		},
		BusinessAddress: &glmodel.LedgerEntityAddress{
			EntityAddressId: businessAddress.ID,
			RawAddress:      businessAddress.RawAddress,
			House:           businessAddress.House,
			Category:        businessAddress.Category,
			Near:            businessAddress.Near,
			HouseNumber:     businessAddress.HouseNumber,
			Road:            businessAddress.Road,
			Unit:            businessAddress.Unit,
			Level:           businessAddress.Level,
			Staircase:       businessAddress.Staircase,
			Entrance:        businessAddress.Entrance,
			PoBox:           businessAddress.POBox,
			Postcode:        businessAddress.Postcode,
			Suburb:          businessAddress.Suburb,
			CityDistrict:    businessAddress.CityDistrict,
			City:            businessAddress.City,
			Island:          businessAddress.Island,
			StateDistrict:   businessAddress.StateDistrict,
			State:           businessAddress.State,
			CountryRegion:   businessAddress.CountryRegion,
			Country:         businessAddress.Country,
			WorldRegion:     businessAddress.WorldRegion,
			CreateDate:      businessAddress.CreateDate.String(),
			ParsedDate:      businessAddress.ParsedDate.String(),
		},
		VerificationStatus: verificationStatus,
		CreateDate:         entity.CreateDate.String(),
		ModifiedDate:       entity.UpdateDate.String(),
	}

	return ledgerEntity, nil
}
