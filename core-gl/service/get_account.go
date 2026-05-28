package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/openreserveio/core/core-gl/generated/model"
	log "github.com/sirupsen/logrus"
)

type AccountInfo struct {
	AccountId         string
	Code              string
	Name              string
	AccountClass      string
	OwningEntityId    string
	Currency          string
	Metadata          map[string]string
	ParentAccountCode string
	Balance           int64
	BalanceAsOfDate   string
}

func GetAccount(ctx context.Context, coreLedgerClient model.CoreLedgerServiceClient, ledgerId string, accountCode string, balanceAsOfDate string) (*AccountInfo, error) {

	// Get Account Info and Balances as of date
	log.Infof("Getting Account Info for %s", accountCode)
	getLedgerAccountResponse, err := coreLedgerClient.GetLedgerAccount(ctx, &model.GetLedgerAccountRequest{Code: accountCode, LedgerId: ledgerId})
	if err != nil {
		return nil, err
	}
	if getLedgerAccountResponse.Status.Code == http.StatusNotFound {
		return nil, nil
	}
	if getLedgerAccountResponse.Status.Code != 200 {
		return nil, fmt.Errorf("Error getting account info: %v", getLedgerAccountResponse.Status.StatusMessage)
	}

	var ledgerAccountBalanceResponse *model.GetLedgerAccountBalanceResponse
	if balanceAsOfDate != "" {

		log.Infof("...as of date %s", balanceAsOfDate)
		ledgerAccountBalanceResponse, err = coreLedgerClient.GetLedgerAccountBalance(ctx, &model.GetLedgerAccountBalanceRequest{
			LedgerId:  ledgerId,
			AccountId: getLedgerAccountResponse.AccountId,
			AsOfDate:  &balanceAsOfDate,
		})

	} else {

		log.Infof("...as of NOW")
		ledgerAccountBalanceResponse, err = coreLedgerClient.GetLedgerAccountBalance(ctx, &model.GetLedgerAccountBalanceRequest{
			LedgerId:  ledgerId,
			AccountId: getLedgerAccountResponse.AccountId,
		})

	}

	if err != nil {
		return nil, err
	}
	if ledgerAccountBalanceResponse.Status.Code != 200 {
		return nil, fmt.Errorf("Error getting account balance: %v", ledgerAccountBalanceResponse.Status.StatusMessage)
	}

	accountInfo := AccountInfo{
		AccountId:       getLedgerAccountResponse.AccountId,
		Code:            getLedgerAccountResponse.Code,
		Name:            getLedgerAccountResponse.Name,
		AccountClass:    getLedgerAccountResponse.Class,
		OwningEntityId:  "", // TODO
		Currency:        getLedgerAccountResponse.Currency,
		Metadata:        getLedgerAccountResponse.Metadata,
		Balance:         ledgerAccountBalanceResponse.Balance,
		BalanceAsOfDate: ledgerAccountBalanceResponse.BalanceAsOfDate,
	}

	// Checking for parent account
	if getLedgerAccountResponse.ParentAccountId != "" {
		parentLedgerAccount, err := coreLedgerClient.GetLedgerAccount(ctx, &model.GetLedgerAccountRequest{Code: getLedgerAccountResponse.ParentAccountId, LedgerId: ledgerId})
		if err != nil {
			return nil, err
		}
		accountInfo.ParentAccountCode = parentLedgerAccount.Code
	}

	return &accountInfo, nil

}
