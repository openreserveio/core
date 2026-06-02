package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/openreserveio/core/core-gl/generated/model"
	"github.com/openreserveio/core/core-gl/glmodel"
	"github.com/openreserveio/core/core-util/otel"
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

	ctx, st := otel.StartSpan(ctx, "service.CreateRegularGLAccount")
	defer otel.EndSpan(ctx, st)

	log.Info("Creating GL Account")
	otel.AddEvent(st, "Creating GL Account: %v", glAccountConfig)

	// Check for existing account with code
	otel.AddEvent(st, "Checking for existing account with code: %s", glAccountConfig.AccountCode)
	existingAccount, err := coreLedgerClient.GetLedgerAccount(ctx, &model.GetLedgerAccountRequest{Code: glAccountConfig.AccountCode, LedgerId: ledgerId})
	if err != nil {
		otel.AddError(st, "Error checking for existing account", err)
		log.Errorf("Unable to check for existing account due to error:  %v", err)
		return err
	}
	if existingAccount.Status.Code != http.StatusNotFound {
		otel.AddEvent(st, "Account with code %s already exists", glAccountConfig.AccountCode)
		return fmt.Errorf("Account with code %s already exists", glAccountConfig.AccountCode)
	}

	// If there is a parent account code, lookup the parent account ID
	var parentAccountId string
	if glAccountConfig.ParentAccountCode != "" {
		otel.AddEvent(st, "Looking up parent account ID for code: %s", glAccountConfig.ParentAccountCode)
		parentResponse, err := coreLedgerClient.GetLedgerAccount(ctx, &model.GetLedgerAccountRequest{Code: glAccountConfig.ParentAccountCode, LedgerId: ledgerId})
		if err != nil {
			otel.AddError(st, "Error looking up parent account ID", err)
			return err
		}
		parentAccountId = parentResponse.AccountId
	}

	// Create account
	otel.AddEvent(st, "Creating Account via LedgerService")
	createLedgerAccountReq := model.CreateLedgerAccountRequest{
		LedgerId: ledgerId,
		Name:     glAccountConfig.AccountName,
		Code:     glAccountConfig.AccountCode,
		Class:    glAccountConfig.AccountClass,
		Metadata: map[string]string{
			glmodel.MD_KEY_TAGS:         strings.Join(glAccountConfig.Tags, ","),
			glmodel.MD_KEY_ACCOUNT_TYPE: glmodel.MD_ACCOUNT_TYPE_REGULAR_GL,
		},
		ParentAccountId: parentAccountId,
		Currency:        glAccountConfig.Currency,
	}
	createLedgerAccountResp, err := coreLedgerClient.CreateLedgerAccount(ctx, &createLedgerAccountReq)
	if err != nil {
		otel.AddError(st, "Error creating account via LedgerService", err)
		return err
	}
	glAccountConfig.AccountID = createLedgerAccountResp.AccountId
	otel.AddEvent(st, "Created Account via LedgerService: %s", glAccountConfig.AccountID)

	return nil

}
