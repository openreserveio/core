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

type RegularGLAccountConfig struct {
	AccountID         string
	AccountCode       string
	AccountName       string
	Currency          string
	Tags              []string
	ParentAccountCode string
	AccountClass      string
}

func CreateRegularGLAccount(ctx context.Context, coreLedgerClient model.CoreLedgerServiceClient, ledgerId string, glAccountConfig *RegularGLAccountConfig) error {

	log.Info("Creating GL Account")

	// Check for existing account with code
	existingAccount, err := coreLedgerClient.GetLedgerAccount(ctx, &model.GetLedgerAccountRequest{Code: glAccountConfig.AccountCode, LedgerId: ledgerId})
	if err != nil {
		log.Errorf("Unable to check for existing account due to error:  %v", err)
		return err
	}
	if existingAccount.Status.Code != http.StatusNotFound {
		return fmt.Errorf("Account with code %s already exists", glAccountConfig.AccountCode)
	}

	// If there is a parent account code, lookup the parent account ID
	var parentAccountCode string
	if glAccountConfig.ParentAccountCode != "" {
		parentResponse, err := coreLedgerClient.GetLedgerAccount(ctx, &model.GetLedgerAccountRequest{Code: glAccountConfig.ParentAccountCode, LedgerId: ledgerId})
		if err != nil {
			return err
		}
		parentAccountCode = parentResponse.AccountId
	}

	// Create account
	createLedgerAccountReq := model.CreateLedgerAccountRequest{
		LedgerId: ledgerId,
		Name:     glAccountConfig.AccountName,
		Code:     glAccountConfig.AccountCode,
		Class:    glAccountConfig.AccountClass,
		Metadata: map[string]string{
			glmodel.MD_KEY_TAGS:         strings.Join(glAccountConfig.Tags, ","),
			glmodel.MD_KEY_ACCOUNT_TYPE: glmodel.MD_ACCOUNT_TYPE_REGULAR_GL,
		},
		ParentAccountId: parentAccountCode,
		Currency:        glAccountConfig.Currency,
	}
	createLedgerAccountResp, err := coreLedgerClient.CreateLedgerAccount(ctx, &createLedgerAccountReq)
	if err != nil {
		return err
	}
	glAccountConfig.AccountID = createLedgerAccountResp.AccountId

	return nil

}
