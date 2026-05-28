package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/openreserveio/core/core-gl/generated/model"
	"github.com/openreserveio/core/core-gl/glmodel"
	log "github.com/sirupsen/logrus"
)

type USFednowAccountConfig struct {
	AccountID         string
	AccountCode       string
	AccountName       string
	AccountClass      string
	Tags              []string
	ParentAccountCode string
	Currency          string
}

func CreateUSFedNowAccount(ctx context.Context, coreLedgerClient model.CoreLedgerServiceClient, ledgerId string, fednowAccountConfig *USFednowAccountConfig) error {

	log.Info("Creating US Fednow Account")

	// Check for existing account with code
	existingAccount, err := coreLedgerClient.GetLedgerAccount(ctx, &model.GetLedgerAccountRequest{Code: fednowAccountConfig.AccountCode, LedgerId: ledgerId})
	if err != nil {
		log.Errorf("Unable to check for existing account due to error:  %v", err)
		return err
	}
	if existingAccount.Status.Code != http.StatusNotFound {
		return fmt.Errorf("Account with code %s already exists", fednowAccountConfig.AccountCode)
	}

	// If there is a parent account code, lookup the parent account ID
	var parentAccountCode string
	if fednowAccountConfig.ParentAccountCode != "" {
		parentResponse, err := coreLedgerClient.GetLedgerAccount(ctx, &model.GetLedgerAccountRequest{Code: fednowAccountConfig.ParentAccountCode, LedgerId: ledgerId})
		if err != nil {
			return err
		}
		parentAccountCode = parentResponse.AccountId
	}

	// Create account
	createLedgerAccountReq := model.CreateLedgerAccountRequest{
		LedgerId: ledgerId,
		Name:     fednowAccountConfig.AccountName,
		Code:     fednowAccountConfig.AccountCode,
		Class:    fednowAccountConfig.AccountClass,
		Metadata: map[string]string{
			glmodel.MD_KEY_TAGS:            strings.Join(fednowAccountConfig.Tags, ","),
			glmodel.MD_KEY_ACCOUNT_TYPE:    glmodel.MD_ACCOUNT_TYPE_PAYMENT,
			glmodel.MD_KEY_PAYMENT_CHANNEL: glmodel.MD_PAYMENT_CHANNEL_US_FEDNOW,
		},
		ParentAccountId: parentAccountCode,
		Currency:        fednowAccountConfig.Currency,
	}
	createLedgerAccountResp, err := coreLedgerClient.CreateLedgerAccount(ctx, &createLedgerAccountReq)
	if err != nil {
		return err
	}
	fednowAccountConfig.AccountID = createLedgerAccountResp.AccountId

	return nil

}
