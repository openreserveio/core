package service

import (
	"context"
	"errors"
	"net/http"

	"github.com/openreserveio/core/core-gl/generated/model"
	glmodelint "github.com/openreserveio/core/core-gl/glmodel"
	log "github.com/sirupsen/logrus"
)

func CreateChartOfAccounts(ctx context.Context, coreLedgerClient model.CoreLedgerServiceClient, coa *glmodelint.ChartOfAccounts) (*glmodelint.ChartOfAccounts, error) {

	for i, acctDef := range coa.Assets {

		err := createAccountTree(ctx, coreLedgerClient, coa.LedgerID, &acctDef, "")
		if err != nil {
			log.Errorf("CreateChartOfAccounts Asset error : %v", err)
			return nil, err
		}
		coa.Assets[i].AccountID = acctDef.AccountID

	}

	for i, acctDef := range coa.Liabilities {

		err := createAccountTree(ctx, coreLedgerClient, coa.LedgerID, &acctDef, "")
		if err != nil {
			log.Errorf("CreateChartOfAccounts Liability error : %v", err)
			return nil, err
		}
		coa.Liabilities[i].AccountID = acctDef.AccountID

	}

	for i, acctDef := range coa.Equity {

		err := createAccountTree(ctx, coreLedgerClient, coa.LedgerID, &acctDef, "")
		if err != nil {
			log.Errorf("CreateChartOfAccounts Equity error : %v", err)
			return nil, err
		}
		coa.Equity[i].AccountID = acctDef.AccountID

	}

	for i, acctDef := range coa.Income {

		err := createAccountTree(ctx, coreLedgerClient, coa.LedgerID, &acctDef, "")
		if err != nil {
			log.Errorf("CreateChartOfAccounts Income error : %v", err)
			return nil, err
		}
		coa.Income[i].AccountID = acctDef.AccountID

	}

	for i, acctDef := range coa.Expense {

		err := createAccountTree(ctx, coreLedgerClient, coa.LedgerID, &acctDef, "")
		if err != nil {
			log.Errorf("CreateChartOfAccounts Expense error : %v", err)
			return nil, err
		}
		coa.Expense[i].AccountID = acctDef.AccountID

	}

	return coa, nil

}

func createAccountTree(ctx context.Context, coreLedgerClient model.CoreLedgerServiceClient, ledgerId string, account *glmodelint.FinancialAccount, parentAccountId string) error {

	// Create this account
	response, err := coreLedgerClient.CreateLedgerAccount(ctx, &model.CreateLedgerAccountRequest{
		LedgerId:        ledgerId,
		Name:            account.Name,
		Code:            account.Code,
		Class:           account.Class,
		Metadata:        nil,
		ParentAccountId: parentAccountId,
		Currency:        account.Currency,
	})
	if err != nil {
		log.Printf("CreateAccount error: %s", err)
		return err
	}
	if response.Status.Code != http.StatusOK {
		return errors.New(response.Status.StatusMessage)
	}

	account.AccountID = response.AccountId
	log.Infof("Created Account %s", account.Name)

	// Process children, if any
	for i, child := range account.Children {
		err = createAccountTree(ctx, coreLedgerClient, ledgerId, &child, account.AccountID)
		if err != nil {
			return err
		}
		account.Children[i].AccountID = child.AccountID
	}

	return nil

}
