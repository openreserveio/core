package activities

import (
	"context"
	"encoding/xml"
	"strings"

	"github.com/moov-io/fednow20022/gen/pacs_008_001_08"
	"github.com/openreserveio/core/core-payments/generated/glmodel"
	"github.com/openreserveio/core/core-payments/pmtmodel"
	log "github.com/sirupsen/logrus"
)

func (act *PaymentActivity) ProcessEntities(ctx context.Context, payment pmtmodel.Payment) (pmtmodel.Payment, error) {

	// We need to pull the entities from the raw fednow message
	fednowMessage := ConvertToPacsMessage(payment.PaymentMessage)
	for _, txfr := range fednowMessage.FIToFICstmrCdtTrf.CdtTrfTxInf {

		var udName string
		var udBuildingNumber string
		var udStreetName string
		var udCity string
		var udState string
		var udPostCode string
		var udCountryCode string

		var ucName string
		var ucBuildingNumber string
		var ucStreetName string
		var ucCity string
		var ucState string
		var ucPostCode string
		var ucCountryCode string

		// Ultimate Debtor (Originator)
		if txfr.UltmtDbtr != nil && txfr.UltmtDbtr.Nm != nil {

			udName = string(*txfr.UltmtDbtr.Nm)
			if txfr.UltmtDbtr.PstlAdr != nil {

				if txfr.UltmtDbtr.PstlAdr.BldgNb != nil {
					udBuildingNumber = string(*txfr.UltmtDbtr.PstlAdr.BldgNb)
				}
				if txfr.UltmtDbtr.PstlAdr.StrtNm != nil {
					udStreetName = string(*txfr.UltmtDbtr.PstlAdr.StrtNm)
				}
				if txfr.UltmtDbtr.PstlAdr.TwnNm != nil {
					udCity = string(*txfr.UltmtDbtr.PstlAdr.TwnNm)
				}
				if txfr.UltmtDbtr.PstlAdr.PstCd != nil {
					udPostCode = string(*txfr.UltmtDbtr.PstlAdr.PstCd)
				}
				if txfr.UltmtDbtr.PstlAdr.CtrySubDvsn != nil {
					udState = string(*txfr.UltmtDbtr.PstlAdr.CtrySubDvsn)
				}
				if txfr.UltmtDbtr.PstlAdr.Ctry != nil {
					udCountryCode = string(*txfr.UltmtDbtr.PstlAdr.Ctry)
				}

			}

		}

		// Ultimate Creditor (Beneficiary)
		if txfr.UltmtCdtr != nil && txfr.UltmtCdtr.Nm != nil {

			ucName = string(*txfr.UltmtCdtr.Nm)
			if txfr.UltmtCdtr.PstlAdr != nil {

				if txfr.UltmtCdtr.PstlAdr.BldgNb != nil {
					ucBuildingNumber = string(*txfr.UltmtCdtr.PstlAdr.BldgNb)
				}
				if txfr.UltmtCdtr.PstlAdr.StrtNm != nil {
					ucStreetName = string(*txfr.UltmtCdtr.PstlAdr.StrtNm)
				}
				if txfr.UltmtCdtr.PstlAdr.TwnNm != nil {
					ucCity = string(*txfr.UltmtCdtr.PstlAdr.TwnNm)
				}
				if txfr.UltmtCdtr.PstlAdr.PstCd != nil {
					ucPostCode = string(*txfr.UltmtCdtr.PstlAdr.PstCd)
				}
				if txfr.UltmtCdtr.PstlAdr.CtrySubDvsn != nil {
					ucState = string(*txfr.UltmtCdtr.PstlAdr.CtrySubDvsn)
				}
				if txfr.UltmtCdtr.PstlAdr.Ctry != nil {
					ucCountryCode = string(*txfr.UltmtCdtr.PstlAdr.Ctry)
				}

			}

		}

		// Store these entities in the EntityDB
		// originator
		ucNameParts := strings.Split(ucName, " ")
		var ucFirstName string
		var ucLastName string
		if len(ucNameParts) > 1 {
			ucFirstName = ucNameParts[0]
			ucLastName = ucNameParts[1]
		}
		ucAddrRaw := strings.Join([]string{ucBuildingNumber, ucStreetName, ucCity, ucState, ucPostCode, ucCountryCode}, " ")

		respUC, err := act.CoreGLClient.CreateEntity(ctx, &glmodel.CreateEntityRequest{
			Entity: &glmodel.LedgerEntity{
				EntityType: glmodel.LedgerEntity_INDIVIDUAL,
				EntityName: &glmodel.LedgerEntityName{
					IndividualGivenName: ucFirstName,
					IndividualSurName:   ucLastName,
				},
				MailingAddress: &glmodel.LedgerEntityAddress{
					RawAddress: ucAddrRaw,
				},
				BusinessAddress: &glmodel.LedgerEntityAddress{
					RawAddress: "",
				},
			},
		})
		if err != nil {
			log.Errorf("Error creating entity: %v", err)
			return payment, err
		}
		if respUC.Status.Code != 200 {
			log.Errorf("Error creating entity: %v", respUC.Status.StatusMessage)
			return payment, err
		}

		ucEntityId := respUC.EntityId

		// beneficiary
		udNameParts := strings.Split(udName, " ")
		var udFirstName string
		var udLastName string
		if len(udNameParts) > 1 {
			udFirstName = udNameParts[0]
			udLastName = udNameParts[1]
		}
		udAddrRaw := strings.Join([]string{udBuildingNumber, udStreetName, udCity, udState, udPostCode, udCountryCode}, " ")

		respUD, err := act.CoreGLClient.CreateEntity(ctx, &glmodel.CreateEntityRequest{
			Entity: &glmodel.LedgerEntity{
				EntityType: glmodel.LedgerEntity_INDIVIDUAL,
				EntityName: &glmodel.LedgerEntityName{
					IndividualGivenName: udFirstName,
					IndividualSurName:   udLastName,
				},
				MailingAddress: &glmodel.LedgerEntityAddress{
					RawAddress: udAddrRaw,
				},
				BusinessAddress: &glmodel.LedgerEntityAddress{
					RawAddress: "",
				},
			},
		})
		if err != nil {
			log.Errorf("Error creating entity: %v", err)
			return payment, err
		}
		if respUD.Status.Code != 200 {
			log.Errorf("Error creating entity: %v", respUD.Status.StatusMessage)
			return payment, err
		}

		udEntityId := respUD.EntityId

		payment.UltimateOriginatorEntityID = udEntityId
		payment.UltimateBeneficiaryEntityID = ucEntityId

		_, err = act.PaymentsDB.NewUpdate().Model(&payment).Where("id = ?", payment.ID).Exec(ctx)
		if err != nil {
			log.Errorf("Error updating payment: %v", err)
			return payment, err
		}

	}

	return payment, nil

}

func ConvertToPacsMessage(rawMessage []byte) pacs_008_001_08.Document {

	var fedNowMessage pacs_008_001_08.Document
	err := xml.Unmarshal(rawMessage, &fedNowMessage)
	if err != nil {
		log.Errorf("Error unmarshalling FedNow Message: %v", err)
		return pacs_008_001_08.Document{}
	}

	return fedNowMessage

}
