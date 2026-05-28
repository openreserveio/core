package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/openreserveio/core/core-gl/generated/model"
	glmodelint "github.com/openreserveio/core/core-gl/glmodel"
)

func GetChartOfAccounts(ctx context.Context, coreLedgerClient model.CoreLedgerServiceClient, ledgerId string) (*glmodelint.ChartOfAccounts, error) {

	findResp, err := coreLedgerClient.FindLedgerAccounts(ctx, &model.FindLedgerAccountsRequest{
		LedgerId:     ledgerId,
		CriteriaType: model.FindLedgerAccountsRequest_ALL_IN_LEDGER,
	})
	if err != nil {
		return nil, err
	}
	if findResp.Status.Code == http.StatusNotFound {
		return nil, nil
	}
	if findResp.Status.Code != http.StatusOK {
		return nil, fmt.Errorf("error finding ledger accounts: %s", findResp.Status.StatusMessage)
	}

	// Index accounts by id and group child ids by parent id
	byId := make(map[string]*model.FindLedgerAccountsResponse_MatchedAccount, len(findResp.Accounts))
	childrenOf := make(map[string][]string, len(findResp.Accounts))
	var rootIds []string
	for _, a := range findResp.Accounts {
		byId[a.AccountId] = a
		if a.ParentAccountId != "" {
			childrenOf[a.ParentAccountId] = append(childrenOf[a.ParentAccountId], a.AccountId)
		} else {
			rootIds = append(rootIds, a.AccountId)
		}
	}

	var build func(accountId string) glmodelint.FinancialAccount
	build = func(accountId string) glmodelint.FinancialAccount {
		a := byId[accountId]
		node := glmodelint.FinancialAccount{
			AccountID: a.AccountId,
			Name:      a.Name,
			Class:     a.Class,
			Code:      a.Code,
			Currency:  a.Currency,
			Metadata:  a.Metadata,
		}
		for _, childId := range childrenOf[accountId] {
			node.Children = append(node.Children, build(childId))
		}
		return node
	}

	coa := &glmodelint.ChartOfAccounts{LedgerID: ledgerId}
	for _, id := range rootIds {
		root := build(id)
		switch root.Class {
		case glmodelint.ACCOUNT_CLASS_ASSET:
			coa.Assets = append(coa.Assets, root)
		case glmodelint.ACCOUNT_CLASS_LIABILITY:
			coa.Liabilities = append(coa.Liabilities, root)
		case glmodelint.ACCOUNT_CLASS_EQUITY:
			coa.Equity = append(coa.Equity, root)
		case glmodelint.ACCOUNT_CLASS_INCOME:
			coa.Income = append(coa.Income, root)
		case glmodelint.ACCOUNT_CLASS_EXPENSE:
			coa.Expense = append(coa.Expense, root)
		}
	}

	return coa, nil

}
