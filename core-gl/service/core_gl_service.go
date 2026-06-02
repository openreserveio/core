package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/openreserveio/core/core-gl/generated/glmodel"
	"github.com/openreserveio/core/core-gl/generated/model"
	glmodelint "github.com/openreserveio/core/core-gl/glmodel"
	"github.com/openreserveio/core/core-util/bus"
	"github.com/openreserveio/core/core-util/otel"
	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type CoreGLService struct {
	glmodel.UnimplementedGeneralLedgerServiceServer
	EntityDB         *bun.DB
	Bus              *bus.BusConnection
	CoreLedgerClient model.CoreLedgerServiceClient
}

func NewCoreGLService(entityDbUrl string, busconnurl string, coreLedgerUrl string, listenHost string, listenPort string) (*CoreGLService, error) {

	ctx := context.Background()
	ctx, st := otel.StartSpan(ctx, "CoreGLService.NewCoreGLService")
	defer otel.EndSpan(ctx, st)

	coreGL := CoreGLService{}

	// Bus
	otel.AddEvent(st, "Setting up Bus Connection")
	busConn, err := bus.NewBusConnection(ctx, busconnurl)
	if err != nil {
		otel.AddError(st, "Error creating bus connection", err)
		log.Error("Error creating bus connection: %v", err)
		return nil, err
	}
	coreGL.Bus = busConn

	// EntityDB
	// EntityDB Connection Setup
	// Using pgdriver (recommended)
	otel.AddEvent(st, "Setting up DB Connection")
	dbConn := sql.OpenDB(pgdriver.NewConnector(
		pgdriver.WithDSN(entityDbUrl),
	))
	dbBun := bun.NewDB(dbConn, pgdialect.New())
	coreGL.EntityDB = dbBun

	// Setup Core Ledger Service
	otel.AddEvent(st, "Setting up Core Ledger Service")
	conn, err := grpc.NewClient(coreLedgerUrl, grpc.WithTransportCredentials(insecure.NewCredentials()), otel.InjectClientGRPCHeaders())
	if err != nil {
		otel.AddError(st, "Error creating Core Ledger Client", err)
		log.Fatalf("Unable to connect to Core Ledger: %v", err)
		os.Exit(1)
	}
	coreGL.CoreLedgerClient = model.NewCoreLedgerServiceClient(conn)

	return &coreGL, nil
}

func (cgls *CoreGLService) GetChartOfAccounts(ctx context.Context, request *glmodel.GetChartOfAccountsRequest) (*glmodel.GetChartOfAccountsResponse, error) {

	response := glmodel.GetChartOfAccountsResponse{}

	coa, err := GetChartOfAccounts(ctx, cgls.CoreLedgerClient, request.LedgerId)
	if err != nil {
		response.Status = &glmodel.GetChartOfAccountsResponse_Status{Code: http.StatusInternalServerError, StatusMessage: err.Error()}
		return &response, nil
	}
	if coa == nil {
		response.Status = &glmodel.GetChartOfAccountsResponse_Status{Code: http.StatusNotFound, StatusMessage: fmt.Sprintf("No accounts found for ledger %s", request.LedgerId)}
		return &response, nil
	}

	encodedCoa, err := json.Marshal(coa)
	if err != nil {
		response.Status = &glmodel.GetChartOfAccountsResponse_Status{Code: http.StatusInternalServerError, StatusMessage: err.Error()}
		return &response, nil
	}

	response.Status = &glmodel.GetChartOfAccountsResponse_Status{Code: http.StatusOK}
	response.ChartJSON = encodedCoa

	return &response, nil

}

func (cgls *CoreGLService) CreateChartOfAccounts(ctx context.Context, request *glmodel.CreateChartOfAccountsRequest) (*glmodel.CreateChartOfAccountsResponse, error) {

	response := glmodel.CreateChartOfAccountsResponse{}

	// Creating a new ledger to hold the CoA
	ledgerRequest := model.CreateLedgerRequest{
		Name:           request.Title,
		IsSubledger:    false,
		ParentLedgerId: nil,
	}
	createLedgerResponse, err := cgls.CoreLedgerClient.CreateLedger(ctx, &ledgerRequest)
	if err != nil {
		response.Status = &glmodel.CreateChartOfAccountsResponse_Status{Code: http.StatusInternalServerError, StatusMessage: err.Error()}
		return &response, nil
	}

	// Create CoA in new ledger
	var coa glmodelint.ChartOfAccounts
	ledgerId := createLedgerResponse.LedgerId
	err = json.Unmarshal(request.ProposedChartJSON, &coa)
	if err != nil {
		response.Status = &glmodel.CreateChartOfAccountsResponse_Status{Code: http.StatusBadRequest, StatusMessage: err.Error()}
		return &response, nil
	}
	coa.LedgerID = ledgerId

	updatedCoa, err := CreateChartOfAccounts(ctx, cgls.CoreLedgerClient, &coa)
	if err != nil {
		response.Status = &glmodel.CreateChartOfAccountsResponse_Status{Code: http.StatusInternalServerError, StatusMessage: err.Error()}
		return &response, nil
	}

	encodedCoa, err := json.Marshal(updatedCoa)
	if err != nil {
		response.Status = &glmodel.CreateChartOfAccountsResponse_Status{Code: http.StatusInternalServerError, StatusMessage: err.Error()}
		return &response, nil
	}

	response.Status = &glmodel.CreateChartOfAccountsResponse_Status{Code: http.StatusOK}
	response.ChartJSON = encodedCoa

	return &response, nil

}

func (cgls *CoreGLService) PostTransaction(ctx context.Context, request *glmodel.PostTransactionRequest) (*glmodel.PostTransactionResponse, error) {

	response := glmodel.PostTransactionResponse{}

	switch request.TransactionType {

	case glmodel.PostTransactionRequest_JOURNAL_ENTRY:
		transactionId, err := PostJournalEntry(ctx, cgls.Bus, cgls.CoreLedgerClient, request.LedgerId, request.JournalEntry)
		if err != nil {
			response.Status = &glmodel.PostTransactionResponse_Status{Code: http.StatusInternalServerError, StatusMessage: fmt.Sprintf("Error posting journal entry: %v", err)}
			return &response, nil
		}
		response.TransactionId = transactionId
		response.TransactionStatus = "POSTED" // for now

	case glmodel.PostTransactionRequest_US_PAYMENT_FEDNOW:
		fednowResult, err := PostFedNowPayment(ctx, cgls.Bus, cgls.CoreLedgerClient, request.LedgerId, request.JournalEntry, request.CorePayment)
		if err != nil {
			response.Status = &glmodel.PostTransactionResponse_Status{Code: http.StatusInternalServerError, StatusMessage: fmt.Sprintf("Error posting FedNow payment: %v", err)}
			return &response, nil
		}
		response.TransactionId = fednowResult.TransactionID
		response.TransactionStatus = fednowResult.Status

	default:
		response.Status = &glmodel.PostTransactionResponse_Status{Code: http.StatusNotImplemented, StatusMessage: fmt.Sprintf("Unimplemented transaction type: %v", request.TransactionType)}
		return &response, nil

	}

	response.Status = &glmodel.PostTransactionResponse_Status{Code: http.StatusOK}
	return &response, nil

}

func (cgls *CoreGLService) CreateEntity(ctx context.Context, request *glmodel.CreateEntityRequest) (*glmodel.CreateEntityResponse, error) {

	response := glmodel.CreateEntityResponse{}

	// validate
	if request.Entity.EntityId != "" {
		response.Status = &glmodel.CreateEntityResponse_Status{Code: http.StatusBadRequest, StatusMessage: fmt.Sprintf("EntityId is not empty: %v", request.Entity.EntityId)}
		return &response, nil
	}

	// Create the entity
	err := CreateEntity(ctx, cgls.EntityDB, request.Entity)
	if err != nil {
		response.Status = &glmodel.CreateEntityResponse_Status{Code: http.StatusInternalServerError, StatusMessage: err.Error()}
		return &response, nil
	}

	response.Status = &glmodel.CreateEntityResponse_Status{Code: http.StatusOK}
	response.EntityId = request.Entity.EntityId

	return &response, nil
}

func (cgls *CoreGLService) GetEntity(ctx context.Context, request *glmodel.GetEntityRequest) (*glmodel.GetEntityResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (cgls *CoreGLService) GetAccount(ctx context.Context, request *glmodel.GetAccountRequest) (*glmodel.GetAccountResponse, error) {

	response := glmodel.GetAccountResponse{}

	log.Infof("Getting account: %v", request.AccountCode)
	accountInfo, err := GetAccount(ctx, cgls.CoreLedgerClient, request.LedgerId, request.AccountCode, request.AsOfDate)
	if err != nil {
		log.Errorf("Error getting account info: %v", err)
		response.Status = &glmodel.GetAccountResponse_Status{Code: http.StatusInternalServerError, StatusMessage: err.Error()}
		return &response, nil
	}

	if accountInfo == nil {
		response.Status = &glmodel.GetAccountResponse_Status{Code: http.StatusNotFound, StatusMessage: fmt.Sprintf("Account with code %s not found", request.AccountCode)}
		return &response, nil
	}

	response.Status = &glmodel.GetAccountResponse_Status{Code: http.StatusOK}
	response.AccountId = accountInfo.AccountId
	response.AccountClass = accountInfo.AccountClass
	response.AccountCode = accountInfo.Code
	response.ParentAccountCode = accountInfo.ParentAccountCode

	var tags []string
	tagsRaw := accountInfo.Metadata[glmodelint.MD_KEY_TAGS]
	if tagsRaw != "" {
		tags = strings.Split(tagsRaw, ",")
	}
	response.Tags = tags

	return &response, nil

}

func (cgls *CoreGLService) CreateAccount(ctx context.Context, request *glmodel.CreateAccountRequest) (*glmodel.CreateAccountResponse, error) {

	response := glmodel.CreateAccountResponse{}

	switch request.AccountType {

	case glmodel.CreateAccountRequest_REGULAR_GL:
		glAccountConfig := RegularGLAccountConfig{
			AccountCode:       request.AccountCode,
			AccountName:       request.Name,
			Currency:          request.Currency,
			Tags:              request.Tags,
			ParentAccountCode: request.ParentAccountCode,
			AccountClass:      request.AccountClass,
		}
		err := CreateRegularGLAccount(ctx, cgls.CoreLedgerClient, request.LedgerId, &glAccountConfig)
		if err != nil {
			log.Errorf("Error creating GL Account: %v", err)
			response.Status = &glmodel.CreateAccountResponse_Status{Code: http.StatusInternalServerError, StatusMessage: err.Error()}
			return &response, nil
		}

		response.Status = &glmodel.CreateAccountResponse_Status{Code: http.StatusOK}
		response.AccountId = glAccountConfig.AccountID
		response.AccountCode = glAccountConfig.AccountCode

	case glmodel.CreateAccountRequest_PAYMENT_US_FEDNOW:
		fednowAccountConfig := USFednowAccountConfig{
			AccountCode:       request.AccountCode,
			AccountName:       request.Name,
			Currency:          request.Currency,
			Tags:              request.Tags,
			ParentAccountCode: request.ParentAccountCode,
			AccountClass:      request.AccountClass,
		}
		err := CreateUSFedNowAccount(ctx, cgls.CoreLedgerClient, request.LedgerId, &fednowAccountConfig)
		if err != nil {
			log.Errorf("Error creating Fednow Account: %v", err)
			response.Status = &glmodel.CreateAccountResponse_Status{Code: http.StatusInternalServerError, StatusMessage: err.Error()}
			return &response, nil
		}

		response.Status = &glmodel.CreateAccountResponse_Status{Code: http.StatusOK}
		response.AccountId = fednowAccountConfig.AccountID
		response.AccountCode = fednowAccountConfig.AccountCode

	default:
		response.Status = &glmodel.CreateAccountResponse_Status{Code: http.StatusNotImplemented, StatusMessage: fmt.Sprintf("Unimplemented account type: %v", request.AccountType)}
		return &response, nil

	}

	return &response, nil

}

func (cgls *CoreGLService) UpdateAccountMetadata(ctx context.Context, request *glmodel.UpdateAccountMetadataRequest) (*glmodel.UpdateAccountMetadataResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (cgls *CoreGLService) GenerateReport(ctx context.Context, request *glmodel.GenerateReportRequest) (*glmodel.GenerateReportResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (cgls *CoreGLService) mustEmbedUnimplementedGeneralLedgerServiceServer() {
	//TODO implement me
	panic("implement me")
}
