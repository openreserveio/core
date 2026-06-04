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
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type SuspendPayment struct {
	CoreGLClient     glmodel.GeneralLedgerServiceClient
	PaymentsDB       *bun.DB
	AccountingConfig *pmtmodel.AccountingConfig
}

func NewSuspendPaymentWorkflow(ctx context.Context, coreGlUrl string, paymentsdbUrl string, accountConfig *pmtmodel.AccountingConfig) *SuspendPayment {

	// Create connection to PaymentDB
	paymentDbConn := sql.OpenDB(pgdriver.NewConnector(
		pgdriver.WithDSN(paymentsdbUrl),
	))
	paymentsDbBun := bun.NewDB(paymentDbConn, pgdialect.New())

	// Core GL Client
	conn, err := grpc.NewClient(coreGlUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		os.Exit(1)
	}
	gLServiceClient := glmodel.NewGeneralLedgerServiceClient(conn)

	return &SuspendPayment{
		PaymentsDB:       paymentsDbBun,
		CoreGLClient:     gLServiceClient,
		AccountingConfig: accountConfig,
	}
}

func (wf *SuspendPayment) SuspendPaymentForReview(ctx workflow.Context, payment pmtmodel.Payment, screenResults SanctionScreen, workflowDetails WorkflowDetails) error {

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second, //amount of time that must elapse before the first retry occurs
			MaximumInterval:    time.Minute, //maximum interval between retries
			BackoffCoefficient: 2,           //how much the retry interval increases
			MaximumAttempts:    5,           // Uncomment this if you want to limit attempts
		},
		ActivityID: "SuspendPaymentForReview",
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// 1: Update Payment Status
	err := workflow.ExecuteActivity(ctx, (&activities.PaymentActivity{}).UpdatePaymentStatus, payment, pmtmodel.PAYMENT_STATUS_IN_REVIEW).Get(ctx, &payment)
	if err != nil {
		return err
	}

	// 2: Do suspense ledger entry
	err = workflow.ExecuteActivity(ctx, (&activities.SuspendPaymentActivity{}).FednowSuspendPaymentLedgerEntries, payment, wf.AccountingConfig).Get(ctx, &payment)
	if err != nil {
		return err
	}

	// 3: Notify of review
	workflow.GetLogger(ctx).Info(fmt.Sprintf("Workflow ID: %s, Run ID: %s", workflowDetails.WorkflowID, workflowDetails.WorkflowRunID))

	return nil
}
