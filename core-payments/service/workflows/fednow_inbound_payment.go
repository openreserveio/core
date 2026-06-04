package workflows

import (
	"context"
	"database/sql"
	"fmt"
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
	workflowId := workflow.GetInfo(ctx).WorkflowExecution.ID
	workflowRunId := workflow.GetInfo(ctx).WorkflowExecution.RunID

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

	// Validate Payment Instruction and parse the payment instruction
	err = workflow.ExecuteActivity(ctx, (&activities.PaymentActivity{}).ValidatePaymentInstruction, payment).Get(ctx, &payment)
	if err != nil {
		return "", err
	}

	// Update Payment Status
	err = workflow.ExecuteActivity(ctx, (&activities.PaymentActivity{}).UpdatePaymentStatus, payment, pmtmodel.PAYMENT_STATUS_PROCESSING).Get(ctx, &payment)
	if err != nil {
		return "", err
	}

	// Initial Ledger Entries
	err = workflow.ExecuteActivity(ctx, (&activities.PaymentActivity{}).FednowInitialLedgerEntries, payment, *wf.AccountingConfig).Get(ctx, &payment)
	if err != nil {
		return "", err
	}

	// Process all the entities in the payment (ie originator, beneficiary, etc)
	err = workflow.ExecuteActivity(ctx, (&activities.PaymentActivity{}).ProcessEntities, payment).Get(ctx, &payment)
	if err != nil {
		return "", err
	}

	// Kick off sanctions screening child workflow and wait for results
	var sanctionsScreenOriginator SanctionScreen
	var sanctionsScreenBeneficiary SanctionScreen
	childWorkflowOptions := workflow.ChildWorkflowOptions{
		TaskQueue: TASK_QUEUE_SANCTION_SCREEN,
	}
	ctx = workflow.WithChildOptions(ctx, childWorkflowOptions)

	// ultimate originator
	err = workflow.ExecuteChildWorkflow(ctx, (&EntitySanctionsScreenWorkflow{}).SanctionScreenEntity, payment.UltimateOriginatorEntityID).Get(ctx, &sanctionsScreenOriginator)
	if err != nil {
		return "", err
	}

	// ultimate beneficiary
	err = workflow.ExecuteChildWorkflow(ctx, (&EntitySanctionsScreenWorkflow{}).SanctionScreenEntity, payment.UltimateBeneficiaryEntityID).Get(ctx, &sanctionsScreenBeneficiary)
	if err != nil {
		return "", err
	}

	sanctionsScreen := SanctionScreen{
		Results: append(sanctionsScreenOriginator.Results, sanctionsScreenBeneficiary.Results...),
	}
	sanctionsScreen.Results = append(sanctionsScreen.Results, sanctionsScreenBeneficiary.Results...)

	if len(sanctionsScreen.Results) > 0 {
		// suspend payment, send to suspend payment workflow - async OK
		workflow.ExecuteChildWorkflow(ctx, (&SuspendPayment{}).SuspendPaymentForReview, payment, sanctionsScreen, WorkflowDetails{
			WorkflowID:    workflowId,
			WorkflowRunID: workflowRunId,
		})

		// Wait for signal?
		var resultOfReview ResultOfReview
		reviewCompleteChannel := workflow.GetSignalChannel(ctx, "ReviewCompleteWithResults")
		reviewCompleteChannel.Receive(ctx, &resultOfReview)

		switch resultOfReview.Result {

		case "PAYMENT_APPROVED":
			return "OK", nil

		case "PAYMENT_REJECTED":
			return "REJECTED", nil

		default:
			return "", fmt.Errorf("invalid review result: %v", resultOfReview)

		}

	}

	return "OK", nil
}
