package workflows

import (
	"context"

	"github.com/openreserveio/core/core-payments/generated/glmodel"
	"github.com/openreserveio/core/core-payments/pmtmodel"
	log "github.com/sirupsen/logrus"
)

func CheckForFednowSettlementInProgressAccount(ctx context.Context, glClient glmodel.GeneralLedgerServiceClient, config *pmtmodel.AccountingConfig) bool {

	// Ensure accounts exist from required config
	fednowSettlementInProgressAccount, err := glClient.GetAccount(ctx, &glmodel.GetAccountRequest{LedgerId: config.LedgerID, AccountCode: config.FednowSettlementInProgressAccountCode})
	if err != nil {
		log.Errorf("Error getting FedNow settlement in progress account for setup: %v", err)
		return false
	}
	if fednowSettlementInProgressAccount.Status.Code == 404 {

		// Create the settlement account
		createdSettlementAccount, err := glClient.CreateAccount(ctx, &glmodel.CreateAccountRequest{
			AccountType:       glmodel.CreateAccountRequest_PAYMENT_US_FEDNOW,
			LedgerId:          config.LedgerID,
			AccountCode:       config.FednowSettlementInProgressAccountCode,
			Currency:          "USD",
			Tags:              nil,
			ParentAccountCode: "",
			AccountClass:      "LIABILITY",
			OwningEntityId:    "",
			Name:              "Fednow Settlement In Progress",
		})
		if err != nil {
			log.Errorf("Error creating FedNow settlement in progress account: %v", err)
			return false
		}
		if createdSettlementAccount.Status.Code != 200 {
			log.Errorf("Error creating FedNow settlement in progress account: %v", createdSettlementAccount.Status.StatusMessage)
			return false
		}

	}

	if fednowSettlementInProgressAccount.Status.Code == 200 {
		return true
	}

	log.Errorf("Error getting FedNow settlement in progress account: %v", fednowSettlementInProgressAccount.Status.StatusMessage)
	return false

}

func CheckForFednowClearingAccount(ctx context.Context, glClient glmodel.GeneralLedgerServiceClient, config *pmtmodel.AccountingConfig) bool {

	// Ensure accounts exist from required config
	fednowClearingAccount, err := glClient.GetAccount(ctx, &glmodel.GetAccountRequest{LedgerId: config.LedgerID, AccountCode: config.FednowClearingAccountCode})
	if err != nil {
		log.Errorf("Error getting FedNow clearing account for setup: %v", err)
		return false
	}
	if fednowClearingAccount.Status.Code == 404 {

		// Create the settlement account
		createdSettlementAccount, err := glClient.CreateAccount(ctx, &glmodel.CreateAccountRequest{
			AccountType:       glmodel.CreateAccountRequest_PAYMENT_US_FEDNOW,
			LedgerId:          config.LedgerID,
			AccountCode:       config.FednowClearingAccountCode,
			Currency:          "USD",
			Tags:              nil,
			ParentAccountCode: "",
			AccountClass:      "LIABILITY",
			OwningEntityId:    "",
			Name:              "Fednow Clearing",
		})
		if err != nil {
			log.Errorf("Error creating FedNow clearing account: %v", err)
			return false
		}
		if createdSettlementAccount.Status.Code != 200 {
			log.Errorf("Error creating FedNow clearing account: %v", createdSettlementAccount.Status.StatusMessage)
			return false
		}

	}

	if fednowClearingAccount.Status.Code == 200 {
		return true
	}

	log.Errorf("Error getting FedNow clearing account: %v", fednowClearingAccount.Status.StatusMessage)
	return false

}

func CheckForFednowSuspenseAccount(ctx context.Context, glClient glmodel.GeneralLedgerServiceClient, config *pmtmodel.AccountingConfig) bool {

	// Ensure accounts exist from required config
	fednowSuspenseAccount, err := glClient.GetAccount(ctx, &glmodel.GetAccountRequest{LedgerId: config.LedgerID, AccountCode: config.FednowSuspenseAccountCode})
	if err != nil {
		log.Errorf("Error getting FedNow suspense account for setup: %v", err)
		return false
	}
	if fednowSuspenseAccount.Status.Code == 404 {

		// Create the settlement account
		createdSettlementAccount, err := glClient.CreateAccount(ctx, &glmodel.CreateAccountRequest{
			AccountType:       glmodel.CreateAccountRequest_PAYMENT_US_FEDNOW,
			LedgerId:          config.LedgerID,
			AccountCode:       config.FednowSuspenseAccountCode,
			Currency:          "USD",
			Tags:              nil,
			ParentAccountCode: "",
			AccountClass:      "LIABILITY",
			OwningEntityId:    "",
			Name:              "Fednow Suspense",
		})
		if err != nil {
			log.Errorf("Error creating FedNow suspense account: %v", err)
			return false
		}
		if createdSettlementAccount.Status.Code != 200 {
			log.Errorf("Error creating FedNow suspense account: %v", createdSettlementAccount.Status.StatusMessage)
			return false
		}

	}

	if fednowSuspenseAccount.Status.Code == 200 {
		return true
	}

	log.Errorf("Error getting FedNow clearing account: %v", fednowSuspenseAccount.Status.StatusMessage)
	return false

}

func CheckForFednowSettlementAccount(ctx context.Context, glClient glmodel.GeneralLedgerServiceClient, config *pmtmodel.AccountingConfig) bool {

	// Ensure accounts exist from required config
	fednowSettlementAccount, err := glClient.GetAccount(ctx, &glmodel.GetAccountRequest{LedgerId: config.LedgerID, AccountCode: config.FednowSettlementAccountCode})
	if err != nil {
		log.Errorf("Error getting FedNow settlement account for setup: %v", err)
		return false
	}
	if fednowSettlementAccount.Status.Code == 404 {

		// Create the settlement account
		createdSettlementAccount, err := glClient.CreateAccount(ctx, &glmodel.CreateAccountRequest{
			AccountType:       glmodel.CreateAccountRequest_PAYMENT_US_FEDNOW,
			LedgerId:          config.LedgerID,
			AccountCode:       config.FednowSettlementAccountCode,
			Currency:          "USD",
			Tags:              nil,
			ParentAccountCode: "",
			AccountClass:      "ASSET",
			OwningEntityId:    "",
			Name:              "Fednow Settlement",
		})
		if err != nil {
			log.Errorf("Error creating FedNow settlement account: %v", err)
			return false
		}
		if createdSettlementAccount.Status.Code != 200 {
			log.Errorf("Error creating FedNow settlement account: %v", createdSettlementAccount.Status.StatusMessage)
			return false
		}

	}

	if fednowSettlementAccount.Status.Code == 200 {
		return true
	}

	log.Errorf("Error getting FedNow settlement account: %v", fednowSettlementAccount.Status.StatusMessage)
	return false

}
