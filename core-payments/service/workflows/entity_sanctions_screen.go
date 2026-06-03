package workflows

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/moov-io/watchman/pkg/search"
	"github.com/openreserveio/core/core-payments/generated/glmodel"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type EntitySanctionsScreenWorkflow struct {
	GLServiceClient       glmodel.GeneralLedgerServiceClient
	SanctionsSearchClient search.Client
}

func NewEntitySanctionsScreenWorkflow(ctx context.Context, coreGlUrl string, watchmanUrl string) *EntitySanctionsScreenWorkflow {

	sanctionsWF := EntitySanctionsScreenWorkflow{}

	// Core GL Client
	conn, err := grpc.NewClient(coreGlUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		os.Exit(1)
	}
	sanctionsWF.GLServiceClient = glmodel.NewGeneralLedgerServiceClient(conn)

	// Watchman Client
	wc := search.NewClient(http.DefaultClient, watchmanUrl)
	sanctionsWF.SanctionsSearchClient = wc

	return &sanctionsWF

}

func (wf *EntitySanctionsScreenWorkflow) SanctionScreenEntity(ctx workflow.Context, entityId string) error {

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second, //amount of time that must elapse before the first retry occurs
			MaximumInterval:    time.Minute, //maximum interval between retries
			BackoffCoefficient: 2,           //how much the retry interval increases
			MaximumAttempts:    5,           // Uncomment this if you want to limit attempts
		},
		ActivityID: "SanctionScreenEntity",
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	/** SANCTIONS SCREEN WORKFLOW
	 *
	 */

	return nil

}
