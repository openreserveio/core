package workflows

import (
	"context"
	"database/sql"
	"os"
	"time"

	"github.com/openreserveio/core/core-payments/generated/glmodel"
	"github.com/openreserveio/core/core-payments/pmtmodel"
	"github.com/openreserveio/core/core-payments/service/activities"
	"github.com/openreserveio/core/core-util/otel"
	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type FednowInboundPaymentWorkflow struct {
	PaymentsDB       *bun.DB
	EntityDB         *bun.DB
	GLServiceClient  glmodel.GeneralLedgerServiceClient
	AccountingConfig *pmtmodel.AccountingConfig
}

func NewFednowInboundPaymentWorkflow(ctx context.Context, paymentsdbUrl string, entitydbUrl string, coreGlUrl string, accountingConfig *pmtmodel.AccountingConfig) *FednowInboundPaymentWorkflow {

	ctx, st := otel.StartSpan(ctx, "workflows.NewFednowInboundPaymentWorkflow")
	defer otel.EndSpan(ctx, st)

	fednowInboundWF := FednowInboundPaymentWorkflow{
		AccountingConfig: accountingConfig,
	}

	otel.AddEvent(st, "Connecting to Payments DB")
	dbConn := sql.OpenDB(pgdriver.NewConnector(
		pgdriver.WithDSN(paymentsdbUrl),
	))
	dbBun := bun.NewDB(dbConn, pgdialect.New())
	fednowInboundWF.PaymentsDB = dbBun

	otel.AddEvent(st, "Connecting to Entity DB")
	entityDbConn := sql.OpenDB(pgdriver.NewConnector(
		pgdriver.WithDSN(entitydbUrl),
	))
	entityDbBun := bun.NewDB(entityDbConn, pgdialect.New())
	fednowInboundWF.EntityDB = entityDbBun

	// Core GL Client
	otel.AddEvent(st, "Connecting to Core GL")
	conn, err := grpc.NewClient(coreGlUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		os.Exit(1)
	}
	fednowInboundWF.GLServiceClient = glmodel.NewGeneralLedgerServiceClient(conn)

	// Required Accounts
	log.Infof("Checking for Fednow Settlement Account")
	exists := CheckForFednowSettlementAccount(context.Background(), fednowInboundWF.GLServiceClient, fednowInboundWF.AccountingConfig)
	if !exists {
		log.Errorf("Fednow Settlement Account does not exist. Cannot continue.")
		return nil
	}

	log.Infof("Checking for Fednow Settlement In Progress Account")
	exists = CheckForFednowSettlementInProgressAccount(context.Background(), fednowInboundWF.GLServiceClient, fednowInboundWF.AccountingConfig)
	if !exists {
		log.Errorf("Fednow Settlement In Progress Account does not exist. Cannot continue.")
		return nil
	}

	log.Infof("Checking for Fednow Clearing Account")
	exists = CheckForFednowClearingAccount(context.Background(), fednowInboundWF.GLServiceClient, fednowInboundWF.AccountingConfig)
	if !exists {
		log.Errorf("Fednow Clearing Account does not exist. Cannot continue.")
		return nil
	}

	log.Infof("Checking for Fednow Suspense Account")
	exists = CheckForFednowSuspenseAccount(context.Background(), fednowInboundWF.GLServiceClient, fednowInboundWF.AccountingConfig)
	if !exists {
		log.Errorf("Fednow Suspense Account does not exist. Cannot continue.")
		return nil
	}

	return &fednowInboundWF
}

func (wf *FednowInboundPaymentWorkflow) ProcessFednowInboundPayment(ctx workflow.Context, rawPaymentInstruction []byte) (string, error) {

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second, //amount of time that must elapse before the first retry occurs
			MaximumInterval:    time.Minute, //maximum interval between retries
			BackoffCoefficient: 2,           //how much the retry interval increases
			MaximumAttempts:    5,           // Uncomment this if you want to limit attempts
		},
		ActivityID: "ProcessFednowInboundPayment",
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	/****************************************************************************************************************************************************************************************************
	 ****************************************************************************************************************************************************************************************************
										PROCESS FEDNOW PAYMENT WORKFLOW
	 ****************************************************************************************************************************************************************************************************
	 ****************************************************************************************************************************************************************************************************
	*/

	var payment pmtmodel.Payment

	// Store Raw Payment Instruction for further processing
	err := workflow.ExecuteActivity(ctx, (&activities.PaymentActivity{}).StoreRawPaymentInstruction, rawPaymentInstruction).Get(ctx, &payment)
	if err != nil {
		return "", err
	}

	// Update Payment Status
	err = workflow.ExecuteActivity(ctx, (&activities.PaymentActivity{}).UpdatePaymentStatus, payment, pmtmodel.PAYMENT_STATUS_INSTRUCTION_RECEIVED).Get(ctx, &payment)
	if err != nil {
		return "", err
	}

	// Validate Payment Instruction and parse the payment instruction
	err = workflow.ExecuteActivity(ctx, (&activities.PaymentActivity{}).ValidatePaymentInstruction, payment).Get(ctx, &payment)
	if err != nil {
		return "", err
	}

	// Initial Ledger Entries
	err = workflow.ExecuteActivity(ctx, (&activities.PaymentActivity{}).FednowInitialLedgerEntries, payment, *wf.AccountingConfig).Get(ctx, &payment)
	if err != nil {
		return "", err
	}

	// Pass off to generic payment processing workflow
	childWorkflowOptions := workflow.ChildWorkflowOptions{
		ParentClosePolicy: enums.PARENT_CLOSE_POLICY_ABANDON,
		TaskQueue:         TASK_QUEUE_INBOUND_PAYMENT_PROCESSING,
	}
	childOptionsCtx := workflow.WithChildOptions(ctx, childWorkflowOptions)
	wfFut := workflow.ExecuteChildWorkflow(childOptionsCtx, (&PaymentProcessingWorkflow{}).ProcessPayment, payment)

	// Wait for the Child Workflow Execution to spawn
	var childWE workflow.Execution
	if err := wfFut.GetChildWorkflowExecution().Get(ctx, &childWE); err != nil {
		return "", err
	}

	return "OK", nil

}
