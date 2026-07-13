package service

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/araddon/dateparse"
	"github.com/google/uuid"
	"github.com/openreserveio/core/core-ledger/generated/model"
	ledgermodel "github.com/openreserveio/core/core-ledger/model"
	"github.com/openreserveio/core/core-util/otel"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

type CoreLedgerService struct {
	model.UnimplementedCoreLedgerServiceServer
	DB *bun.DB
}

func NewCoreLedgerService(ctx context.Context, dbConnUrl string) (CoreLedgerService, error) {

	ctx, st := otel.StartSpan(ctx, "CoreLedgerService.NewCoreLedgerService")
	defer otel.EndSpan(ctx, st)

	cls := CoreLedgerService{}

	// EntityDB Connection Setup
	// Using pgdriver (recommended)
	otel.AddEvent(st, "Setting up DB Connection")
	dbConn := sql.OpenDB(pgdriver.NewConnector(
		pgdriver.WithDSN(dbConnUrl),
	))
	dbBun := bun.NewDB(dbConn, pgdialect.New())
	cls.DB = dbBun

	return cls, nil
}

func (cls CoreLedgerService) CreateLedger(ctx context.Context, request *model.CreateLedgerRequest) (*model.CreateLedgerResponse, error) {

	ctx, st := otel.StartSpan(ctx, "CoreLedgerService.CreateLedger")
	defer otel.EndSpan(ctx, st)

	var parentLedger string
	var parentLedgerValid bool
	if request.ParentLedgerId == nil {
		parentLedger = ""
		parentLedgerValid = false
	} else if *request.ParentLedgerId == "" {
		parentLedger = ""
		parentLedgerValid = false
	}

	ledger := ledgermodel.Ledger{
		ID:             uuid.NewString(),
		Name:           request.Name,
		IsSubledger:    request.IsSubledger,
		ParentLedgerID: sql.NullString{String: parentLedger, Valid: parentLedgerValid},
	}

	otel.AddEvent(st, "Creating Ledger")
	createdLedger, err := CreateLedger(ctx, cls.DB, &ledger)
	if err != nil {
		response := model.CreateLedgerResponse{
			Status: &model.Status{Code: http.StatusInternalServerError, StatusMessage: err.Error()},
		}
		otel.AddError(st, "Error creating ledger", err)
		return &response, err
	}

	response := model.CreateLedgerResponse{
		Status:   &model.Status{Code: http.StatusOK},
		Name:     createdLedger.Name,
		LedgerId: createdLedger.ID,
	}

	otel.AddEvent(st, "Ledger created!")
	return &response, nil

}

func (cls CoreLedgerService) GetLedger(ctx context.Context, request *model.GetLedgerRequest) (*model.GetLedgerResponse, error) {

	ctx, st := otel.StartSpan(ctx, "CoreLedgerService.GetLedger")
	defer otel.EndSpan(ctx, st)

	otel.AddEvent(st, "Fetching Ledger: %s", request.LedgerId)
	ledger, err := GetLedger(ctx, cls.DB, request.LedgerId)
	if err != nil {
		response := model.GetLedgerResponse{
			Status: &model.Status{Code: http.StatusInternalServerError, StatusMessage: err.Error()},
		}

		otel.AddError(st, "Ledger fetch failed", err)
		return &response, err
	}

	var parentLedgerId string
	if ledger.ParentLedgerID.Valid {
		parentLedgerId = ledger.ParentLedgerID.String
	} else {
		parentLedgerId = ""
	}

	response := model.GetLedgerResponse{
		Status:         &model.Status{Code: http.StatusOK},
		Name:           ledger.Name,
		LedgerId:       ledger.ID,
		ParentLedgerId: &parentLedgerId,
		IsSubledger:    ledger.IsSubledger,
	}

	return &response, nil

}

func (cls CoreLedgerService) PostLedgerTransaction(ctx context.Context, request *model.PostLedgerTransactionRequest) (*model.PostLedgerTransactionResponse, error) {

	ctx, st := otel.StartSpan(ctx, "CoreLedgerService.PostLedgerTransaction")
	defer otel.EndSpan(ctx, st)

	// setup domain objects
	var response model.PostLedgerTransactionResponse
	var debits []ledgermodel.LedgerTransactionEntry
	var credits []ledgermodel.LedgerTransactionEntry
	for _, entry := range request.Debits {
		en := ledgermodel.LedgerTransactionEntry{
			AccountID: entry.AccountId,
			Amount:    entry.Amount,
			Currency:  entry.Currency,
			Metadata:  ConvertMapStringToMapInterface(entry.Metadata),
		}
		debits = append(debits, en)
		otel.AddEvent(st, "Debit: %s - %s %d", entry.AccountId, entry.Currency, entry.Amount)
		otel.AddEvent(st, "Debit Metadata: %v", entry.Metadata)
	}

	for _, entry := range request.Credits {
		en := ledgermodel.LedgerTransactionEntry{
			AccountID: entry.AccountId,
			Amount:    entry.Amount,
			Currency:  entry.Currency,
			Metadata:  ConvertMapStringToMapInterface(entry.Metadata),
		}
		credits = append(credits, en)
		otel.AddEvent(st, "Credit: %s - %s %d", entry.AccountId, entry.Currency, entry.Amount)
		otel.AddEvent(st, "Credit Metadata: %v", entry.Metadata)
	}

	otel.AddEvent(st, "Posting Ledger Transaction")
	tx, balances, err := PostLedgerTransaction(ctx, cls.DB, request.LedgerId, debits, credits, ConvertMapStringToMapInterface(request.Metadata))
	if err != nil {
		otel.AddError(st, "Error posting ledger transaction", err)
		response.Status = &model.Status{Code: http.StatusBadRequest, StatusMessage: err.Error()}
		return &response, nil
	}

	otel.AddEvent(st, "Ledger Transaction Posted")
	response.LedgerTransactionId = tx.ID
	response.LedgerId = tx.LedgerID
	response.PostingDate = tx.TransactionDate.String()
	for _, bal := range balances {

		newBal := model.PostLedgerTransactionResponse_Balance{
			AccountId: bal.AccountID,
			Code:      "",
			Class:     "",
			Name:      "",
			Balance:   bal.Balance,
		}
		response.Balances = append(response.Balances, &newBal)
		otel.AddEvent(st, "Balance: %s - %s %d", newBal.AccountId, newBal.Name, newBal.Balance)
	}

	response.Status = &model.Status{Code: http.StatusOK}
	return &response, nil

}

func (cls CoreLedgerService) GetLedgerTransaction(ctx context.Context, request *model.GetLedgerTransactionRequest) (*model.GetLedgerTransactionResponse, error) {

	ctx, st := otel.StartSpan(ctx, "CoreLedgerService.GetLedgerTransaction")
	defer otel.EndSpan(ctx, st)

	response := model.GetLedgerTransactionResponse{}

	otel.AddEvent(st, "Retrieving Transaction in ledger (%s) - TXID: %s", request.LedgerId, request.LedgerTransactionId)
	ledgerTx, err := GetLedgerTransaction(ctx, cls.DB, request.LedgerId, request.LedgerTransactionId)
	if err != nil {
		otel.AddError(st, "Error retrieving ledger transaction", err)
		response.Status = &model.Status{Code: http.StatusInternalServerError, StatusMessage: err.Error()}
		return &response, nil
	}
	if ledgerTx == nil {
		otel.AddEvent(st, "Ledger Transaction not found")
		response.Status = &model.Status{Code: http.StatusNotFound, StatusMessage: fmt.Sprintf("ledger transaction not found: %s", request.LedgerTransactionId)}
		return &response, nil
	}

	otel.AddEvent(st, "Transaction retrieved successfully")
	response.LedgerId = ledgerTx.LedgerID
	response.Metadata = ConvertMapInterfaceToMapString(ledgerTx.Metadata)

	for _, entry := range ledgerTx.Debits {
		response.Debits = append(response.Debits, &model.GetLedgerTransactionResponse_Entry{
			AccountId: entry.AccountID,
			Amount:    entry.Amount,
		})
		otel.AddEvent(st, "Debit: %s - %d", entry.AccountID, entry.Amount)
	}

	for _, entry := range ledgerTx.Credits {
		response.Credits = append(response.Credits, &model.GetLedgerTransactionResponse_Entry{
			AccountId: entry.AccountID,
			Amount:    entry.Amount,
		})
		otel.AddEvent(st, "Credit: %s - %d", entry.AccountID, entry.Amount)
	}

	response.Status = &model.Status{Code: http.StatusOK}
	return &response, nil

}

func (cls CoreLedgerService) GetLedgerAccountBalance(ctx context.Context, request *model.GetLedgerAccountBalanceRequest) (*model.GetLedgerAccountBalanceResponse, error) {

	var response model.GetLedgerAccountBalanceResponse
	var balance *ledgermodel.AccountBalance
	var err error

	opts := BalanceOpts{
		ForAccountID: request.AccountId,
	}
	if request.AsOfDate != nil {
		asOfDate, _ := dateparse.ParseStrict(*request.AsOfDate)
		opts.AsOfDatetime = &asOfDate
	}

	balance, err = GetBalance(ctx, cls.DB, opts)
	if err != nil {
		response.Status = &model.Status{Code: http.StatusBadRequest, StatusMessage: err.Error()}
		return &response, nil
	}
	if balance == nil {
		response.Status = &model.Status{Code: http.StatusNotFound}
		return &response, nil
	}

	response.Status = &model.Status{Code: http.StatusOK}
	response.AccountId = balance.AccountID
	response.Balance = balance.Balance
	response.BalanceAsOfDate = balance.BalanceDate.String()

	return &response, nil

}

func (cls CoreLedgerService) CreateLedgerAccount(ctx context.Context, request *model.CreateLedgerAccountRequest) (*model.CreateLedgerAccountResponse, error) {

	ctx, st := otel.StartSpan(ctx, "CoreLedgerService.CreateLedgerAccount")
	defer otel.EndSpan(ctx, st)

	var response model.CreateLedgerAccountResponse

	otel.AddEvent(st, "Creating Ledger Account: %s", request.Name)
	acct, err := CreateAccount(ctx, cls.DB, request.LedgerId, request.Name, request.Code, request.Class, request.Metadata, request.ParentAccountId, request.Currency)
	if err != nil {
		otel.AddError(st, "Error creating ledger account", err)
		response.Status = &model.Status{Code: http.StatusBadRequest, StatusMessage: err.Error()}
		return &response, nil
	}

	response.AccountId = acct.ID
	response.Name = acct.Name
	response.Code = acct.Code
	response.LedgerId = acct.LedgerID
	response.Status = &model.Status{Code: http.StatusOK}

	otel.AddEvent(st, "Ledger Account Created: %s", acct.ID)

	return &response, nil
}

func (cls CoreLedgerService) GetLedgerAccount(ctx context.Context, request *model.GetLedgerAccountRequest) (*model.GetLedgerAccountResponse, error) {

	var response model.GetLedgerAccountResponse
	acct, err := GetAccount(ctx, cls.DB, request.LedgerId, request.AccountId, request.Code)
	if err != nil {
		response.Status = &model.Status{Code: http.StatusBadRequest, StatusMessage: err.Error()}
		return &response, nil
	}

	if acct == nil || acct.ID == "" {
		response.Status = &model.Status{Code: http.StatusNotFound}
		return &response, nil
	}

	response.Status = &model.Status{Code: http.StatusOK}
	response.AccountId = acct.ID
	response.Name = acct.Name
	response.Code = acct.Code
	response.Class = acct.Class
	response.LedgerId = acct.LedgerID
	response.Metadata = ConvertMapInterfaceToMapString(acct.Metadata)

	if acct.ParentAccountID.Valid {
		response.ParentAccountId = acct.ParentAccountID.String
	} else {
		response.ParentAccountId = ""
	}

	return &response, nil

}

func (cls CoreLedgerService) FindLedgerAccounts(ctx context.Context, request *model.FindLedgerAccountsRequest) (*model.FindLedgerAccountsResponse, error) {

	response := model.FindLedgerAccountsResponse{}

	switch request.CriteriaType {
	case model.FindLedgerAccountsRequest_BY_METADATA:
		accounts, err := FindAccountsByMetadata(ctx, cls.DB, request.LedgerId, request.MetadataCriteria)
		if err != nil {
			response.Status = &model.Status{Code: http.StatusInternalServerError, StatusMessage: err.Error()}
			return &response, nil
		}
		if accounts == nil {
			response.Status = &model.Status{Code: http.StatusNotFound}
			return &response, nil
		}

		for _, account := range accounts {

			acct := model.FindLedgerAccountsResponse_MatchedAccount{
				MatchScore:      "100",
				AccountId:       account.ID,
				LedgerId:        account.LedgerID,
				Code:            account.Code,
				Name:            account.Name,
				Class:           account.Class,
				Metadata:        ConvertMapInterfaceToMapString(account.Metadata),
				ParentAccountId: account.ParentAccountID.String,
				Currency:        account.Currency,
			}
			response.Accounts = append(response.Accounts, &acct)

		}

	case model.FindLedgerAccountsRequest_ALL_IN_LEDGER:
		accounts, err := FindAllAccountsInLedger(ctx, cls.DB, request.LedgerId)
		if err != nil {
			response.Status = &model.Status{Code: http.StatusInternalServerError, StatusMessage: err.Error()}
			return &response, nil
		}

		for _, account := range accounts {

			acct := model.FindLedgerAccountsResponse_MatchedAccount{
				MatchScore:      "100",
				AccountId:       account.ID,
				LedgerId:        account.LedgerID,
				Code:            account.Code,
				Name:            account.Name,
				Class:           account.Class,
				Metadata:        ConvertMapInterfaceToMapString(account.Metadata),
				ParentAccountId: account.ParentAccountID.String,
				Currency:        account.Currency,
			}
			response.Accounts = append(response.Accounts, &acct)

		}

	case model.FindLedgerAccountsRequest_BY_ACCOUNT_CLASS:
		accounts, err := FindAccountsByClass(ctx, cls.DB, request.LedgerId, request.AccountClass)
		if err != nil {
			response.Status = &model.Status{Code: http.StatusInternalServerError, StatusMessage: err.Error()}
			return &response, nil
		}

		for _, account := range accounts {

			matchedAccount := model.FindLedgerAccountsResponse_MatchedAccount{
				MatchScore:      "100",
				AccountId:       account.ID,
				LedgerId:        account.LedgerID,
				Code:            account.Code,
				Name:            account.Name,
				Class:           account.Class,
				Metadata:        ConvertMapInterfaceToMapString(account.Metadata),
				ParentAccountId: account.ParentAccountID.String,
				Currency:        account.Currency,
			}
			response.Accounts = append(response.Accounts, &matchedAccount)

		}

	}

	response.Status = &model.Status{Code: http.StatusOK}
	return &response, nil

}

func (cls CoreLedgerService) mustEmbedUnimplementedCoreLedgerServiceServer() {
	//TODO implement me
	panic("implement me")
}

//func (cls CoreLedgerService) mustEmbedUnimplementedCoreLedgerServiceServer() {
//	//TODO implement me
//	panic("implement me")
//}
