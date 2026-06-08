package workflows

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/openreserveio/core/core-payments/generated/glmodel"
	"github.com/openreserveio/core/core-payments/pmtmodel"
	"github.com/openreserveio/core/core-payments/service/activities"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type PaymentProcessingWorkflow struct {
	PaymentsDB       *bun.DB
	EntityDB         *bun.DB
	GLServiceClient  glmodel.GeneralLedgerServiceClient
	AccountingConfig *pmtmodel.AccountingConfig
}

func NewPaymentProcessingWorkflow(paymentsdbUrl string, entitydbUrl string, coreGlUrl string, accountingConfig *pmtmodel.AccountingConfig) *PaymentProcessingWorkflow {

	ppw := PaymentProcessingWorkflow{
		AccountingConfig: accountingConfig,
	}

	dbConn := sql.OpenDB(pgdriver.NewConnector(
		pgdriver.WithDSN(paymentsdbUrl),
	))
	dbBunPayments := bun.NewDB(dbConn, pgdialect.New())
	ppw.PaymentsDB = dbBunPayments

	entityDbConn := sql.OpenDB(pgdriver.NewConnector(
		pgdriver.WithDSN(entitydbUrl),
	))
	entityDbBun := bun.NewDB(entityDbConn, pgdialect.New())
	ppw.EntityDB = entityDbBun

	// Core GL Client
	conn, err := grpc.NewClient(coreGlUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		os.Exit(1)
	}
	ppw.GLServiceClient = glmodel.NewGeneralLedgerServiceClient(conn)

	return &ppw
}

func (ppw *PaymentProcessingWorkflow) ProcessPayment(ctx workflow.Context, payment pmtmodel.Payment) (string, error) {

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second, //amount of time that must elapse before the first retry occurs
			MaximumInterval:    time.Minute, //maximum interval between retries
			BackoffCoefficient: 2,           //how much the retry interval increases
			MaximumAttempts:    5,           // Uncomment this if you want to limit attempts
		},
		ActivityID: "InboundPaymentProcessing",
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	/****************************************************************************************************************************************************************************************************
	 ****************************************************************************************************************************************************************************************************
										 INBOUND PAYMENT PROCESSING WORKFLOW
	 ****************************************************************************************************************************************************************************************************
	 ****************************************************************************************************************************************************************************************************
	*/

	workflowId := workflow.GetInfo(ctx).WorkflowExecution.ID
	workflowRunId := workflow.GetInfo(ctx).WorkflowExecution.RunID

	// Update Payment Status
	err := workflow.ExecuteActivity(ctx, (&activities.PaymentActivity{}).UpdatePaymentStatus, payment, pmtmodel.PAYMENT_STATUS_PROCESSING).Get(ctx, &payment)
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

	// Sanctions Review Required
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
			// Update Payment Status
			err = workflow.ExecuteActivity(ctx, (&activities.PaymentActivity{}).UpdatePaymentStatus, payment, pmtmodel.PAYMENT_STATUS_PROCESSING).Get(ctx, &payment)
			if err != nil {
				return "", err
			}

		case "PAYMENT_REJECTED":
			// Send to Child Workflow "Return"

		default:
			return "", fmt.Errorf("invalid review result: %v", resultOfReview)

		}

	}

	// Now do Transaction Monitoring
	err = workflow.ExecuteActivity(ctx, (&activities.PaymentActivity{}).GetTransactionMonitoringRisk, ctx, payment).Get(ctx, &payment)
	if err != nil {
		return "", err
	}

	return "OK", nil

}
