package reports

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/openreserveio/core/core-gl/generated/model"
	"github.com/openreserveio/core/core-gl/glmodel"
	"github.com/openreserveio/core/core-gl/service"
)

func GenerateBalanceSheetReport(ctx context.Context, ledgerClient model.CoreLedgerServiceClient, ledgerId string, asOfDate time.Time) (*glmodel.BalanceSheet, error) {

	// Get Assets
	assets, err := GetBalanceSheetAccountsByClass(ctx, ledgerClient, ledgerId, glmodel.ACCOUNT_CLASS_ASSET)
	if err != nil {
		return nil, err
	}
	assetBalances, err := GetFinancialAccountBalances(ctx, ledgerClient, ledgerId, asOfDate.String(), assets)
	if err != nil {
		return nil, err
	}

	var totalAssets int64 = 0
	for _, balance := range assetBalances {
		totalAssets += balance.Balance
	}

	// Get Liabilities
	liabilities, err := GetBalanceSheetAccountsByClass(ctx, ledgerClient, ledgerId, glmodel.ACCOUNT_CLASS_LIABILITY)
	if err != nil {
		return nil, err
	}
	liabilitiesBalances, err := GetFinancialAccountBalances(ctx, ledgerClient, ledgerId, asOfDate.String(), liabilities)
	if err != nil {
		return nil, err
	}

	var totalLiabilities int64 = 0
	for _, balance := range liabilitiesBalances {
		totalLiabilities += balance.Balance
	}

	// Get Equity
	equity, err := GetBalanceSheetAccountsByClass(ctx, ledgerClient, ledgerId, glmodel.ACCOUNT_CLASS_EQUITY)
	if err != nil {
		return nil, err
	}
	equityBalances, err := GetFinancialAccountBalances(ctx, ledgerClient, ledgerId, asOfDate.String(), equity)
	if err != nil {
		return nil, err
	}

	var totalEquity int64 = 0
	for _, balance := range equityBalances {
		totalEquity += balance.Balance
	}

	// Get Income
	income, err := GetBalanceSheetAccountsByClass(ctx, ledgerClient, ledgerId, glmodel.ACCOUNT_CLASS_INCOME)
	if err != nil {
		return nil, err
	}
	incomeBalances, err := GetFinancialAccountBalances(ctx, ledgerClient, ledgerId, asOfDate.String(), income)
	if err != nil {
		return nil, err
	}

	var totalIncome int64 = 0
	for _, balance := range incomeBalances {
		totalIncome += balance.Balance
	}

	// Get Expenses
	expenses, err := GetBalanceSheetAccountsByClass(ctx, ledgerClient, ledgerId, glmodel.ACCOUNT_CLASS_EXPENSE)
	if err != nil {
		return nil, err
	}
	expensesBalances, err := GetFinancialAccountBalances(ctx, ledgerClient, ledgerId, asOfDate.String(), expenses)
	if err != nil {
		return nil, err
	}

	var totalExpense int64 = 0
	for _, balance := range expensesBalances {
		totalExpense += balance.Balance
	}

	report := glmodel.BalanceSheet{
		LedgerID:         ledgerId,
		Assets:           assetBalances,
		TotalAssets:      totalAssets,
		Liabilities:      liabilitiesBalances,
		TotalLiabilities: totalLiabilities,
		Equity:           equityBalances,
		TotalEquity:      totalEquity,
		Income:           incomeBalances,
		TotalIncome:      totalIncome,
		Expense:          expensesBalances,
		TotalExpense:     totalExpense,
	}

	return &report, nil
}

func GetBalanceSheetAccountsByClass(ctx context.Context, ledgerClient model.CoreLedgerServiceClient, ledgerId string, accountClass string) ([]*glmodel.FinancialAccount, error) {

	var financialAccounts []*glmodel.FinancialAccount
	response, err := ledgerClient.FindLedgerAccounts(ctx, &model.FindLedgerAccountsRequest{
		CriteriaType: model.FindLedgerAccountsRequest_BY_ACCOUNT_CLASS,
		AccountClass: accountClass,
		LedgerId:     ledgerId,
	})
	if err != nil {
		return nil, err
	}
	if response.Status.Code != http.StatusOK {
		return nil, fmt.Errorf("failed to get financialAccounts: %s", response.Status.StatusMessage)
	}

	for _, account := range response.Accounts {

		// Ignore all children accounts
		if account.ParentAccountId != "" {
			continue
		}

		financialAccounts = append(financialAccounts, &glmodel.FinancialAccount{
			AccountID: account.AccountId,
			Name:      account.Name,
			Class:     account.Class,
			Code:      account.Code,
			Currency:  account.Currency,
			Children:  nil,
			Metadata:  account.Metadata,
		})
	}

	return financialAccounts, nil

}

func GetFinancialAccountBalances(ctx context.Context, ledgerClient model.CoreLedgerServiceClient, ledgerId string, asOfDate string, financialAccounts []*glmodel.FinancialAccount) ([]glmodel.FinancialAccountBalance, error) {

	response := []glmodel.FinancialAccountBalance{}
	for _, account := range financialAccounts {
		acctInfo, err := service.GetAccount(ctx, ledgerClient, ledgerId, account.Code, asOfDate)
		if err != nil {
			return nil, err
		}
		response = append(response, glmodel.FinancialAccountBalance{
			FinancialAccount: glmodel.FinancialAccount{
				AccountID: acctInfo.AccountId,
				Name:      acctInfo.Name,
				Class:     acctInfo.AccountClass,
				Code:      acctInfo.Code,
				Currency:  acctInfo.Currency,
				Children:  nil,
				Metadata:  acctInfo.Metadata,
			},
			Balance:  acctInfo.Balance,
			AsOfDate: acctInfo.BalanceAsOfDate,
		})
	}

	return response, nil

}
