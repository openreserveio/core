package fednow

import (
	"context"
	"database/sql"
	"net/url"
	"os"
	"strconv"

	"github.com/openreserveio/core/core-payments/generated/glmodel"
	"github.com/openreserveio/core/core-payments/pmtmodel"
	flow "github.com/s8sg/goflow/flow/v1"
	goflow "github.com/s8sg/goflow/v1"
	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

/*
builder.Function("update-status-instruction-received", wf.CreateUpdatePaymentStatusWorkflowStep(pmtmodel.PAYMENT_STATUS_INSTRUCTION_RECEIVED))
	builder.Function("validate-payment-instruction", wf.ValidatePaymentInstruction)
	builder.Task(tasks.NewIfTask(&wf.WorkflowState.PaymentInstructionValidated, nil))
	builder.Function("update-status-processing", wf.CreateUpdatePaymentStatusWorkflowStep(pmtmodel.PAYMENT_STATUS_PROCESSING))
	builder.Function("initial-ledger-postings", func(ctx context.Context) error { return nil })
	builder.Function("destination-account-lookup", func(ctx context.Context) error { return nil })
	builder.Function("destination-account-posting", func(ctx context.Context) error { return nil })
	builder.Function("finalize-clearing", func(ctx context.Context) error { return nil })
	builder.Function("update-status-processed", wf.CreateUpdatePaymentStatusWorkflowStep(pmtmodel.PAYMENT_STATUS_PROCESSED))

*/

var (
	STEP_NAME_VALIDATE_PAYMENT_INSTRUCTION = "validate-payment-instruction"
	STEP_NAME_UPDATE_STATUS_PROCESSING     = "update-status-processing"
	STEP_NAME_INITIAL_LEDGER_POSTINGS      = "initial-ledger-postings"
	STEP_NAME_PROCESS_ENTITIES             = "process-entities"
)

type FednowInboundPaymentWorkflow struct {
	FlowService      *goflow.FlowService
	PaymentsDB       *bun.DB
	EntityDB         *bun.DB
	CoreGL           glmodel.GeneralLedgerServiceClient
	AccountingConfig *pmtmodel.AccountingConfig
}

func NewFednowInboundPaymentWorkflow(listenHost string, listenPort string, redisUrl string, paymentsdbUrl string, entitydbUrl string, coreGlUrl string, accountingConfig *pmtmodel.AccountingConfig) *FednowInboundPaymentWorkflow {

	wf := FednowInboundPaymentWorkflow{
		AccountingConfig: accountingConfig,
	}

	portInt, _ := strconv.ParseInt(listenPort, 10, 64)
	flowService := goflow.FlowService{
		Port:              int(portInt),
		RedisURL:          redisUrl,
		WorkerConcurrency: 2,
		EnableMonitoring:  false,
	}
	flowService.Register("fednow-inbound-payment-workflows", wf.DefineWorkflow)

	wf.FlowService = &flowService

	// Payments EntityDB
	// EntityDB Connection Setup
	// Using pgdriver (recommended)
	dbConn := sql.OpenDB(pgdriver.NewConnector(
		pgdriver.WithDSN(paymentsdbUrl),
	))
	dbBun := bun.NewDB(dbConn, pgdialect.New())
	wf.PaymentsDB = dbBun

	// Entity EntityDB
	// EntityDB Connection Setup
	// Using pgdriver (recommended)
	entityDbConn := sql.OpenDB(pgdriver.NewConnector(
		pgdriver.WithDSN(entitydbUrl),
	))
	entityDbBun := bun.NewDB(entityDbConn, pgdialect.New())
	wf.EntityDB = entityDbBun

	// Core GL Client
	conn, err := grpc.NewClient(coreGlUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Unable to connect to Core Ledger: %v", err)
		os.Exit(1)
	}
	wf.CoreGL = glmodel.NewGeneralLedgerServiceClient(conn)

	// Required Accounts
	log.Infof("Checking for Fednow Settlement Account")
	exists := CheckForFednowSettlementAccount(context.Background(), wf.CoreGL, wf.AccountingConfig)
	if !exists {
		log.Errorf("Fednow Settlement Account does not exist. Cannot continue.")
		return nil
	}

	log.Infof("Checking for Fednow Settlement In Progress Account")
	exists = CheckForFednowSettlementInProgressAccount(context.Background(), wf.CoreGL, wf.AccountingConfig)
	if !exists {
		log.Errorf("Fednow Settlement In Progress Account does not exist. Cannot continue.")
		return nil
	}

	log.Infof("Checking for Fednow Clearing Account")
	exists = CheckForFednowClearingAccount(context.Background(), wf.CoreGL, wf.AccountingConfig)
	if !exists {
		log.Errorf("Fednow Clearing Account does not exist. Cannot continue.")
		return nil
	}

	log.Infof("Checking for Fednow Suspense Account")
	exists = CheckForFednowSuspenseAccount(context.Background(), wf.CoreGL, wf.AccountingConfig)
	if !exists {
		log.Errorf("Fednow Suspense Account does not exist. Cannot continue.")
		return nil
	}

	return &wf
}

func (f *FednowInboundPaymentWorkflow) Start() error {

	return f.FlowService.Start()

}

func (f *FednowInboundPaymentWorkflow) DefineWorkflow(workflow *flow.Workflow, flowContext *flow.Context) error {

	dag := workflow.Dag()
	dag.Node(STEP_NAME_VALIDATE_PAYMENT_INSTRUCTION, f.ValidatePaymentInstruction)
	dag.Node(STEP_NAME_UPDATE_STATUS_PROCESSING, f.CreateStatusUpdateFlowStep(pmtmodel.PAYMENT_STATUS_PROCESSING))
	dag.Node(STEP_NAME_INITIAL_LEDGER_POSTINGS, f.InitialLedgerPostings)
	dag.Node(STEP_NAME_PROCESS_ENTITIES, f.ProcessEntities)

	dag.Edge(STEP_NAME_VALIDATE_PAYMENT_INSTRUCTION, STEP_NAME_UPDATE_STATUS_PROCESSING)
	dag.Edge(STEP_NAME_UPDATE_STATUS_PROCESSING, STEP_NAME_INITIAL_LEDGER_POSTINGS)
	dag.Edge(STEP_NAME_INITIAL_LEDGER_POSTINGS, STEP_NAME_PROCESS_ENTITIES)

	flowContext.Query = make(url.Values)
	flowContext.Query.Set("flow-context-query", "flow-context-value")

	return nil

}
