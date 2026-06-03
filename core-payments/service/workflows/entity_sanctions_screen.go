package workflows

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/moov-io/watchman/pkg/search"
	"github.com/openreserveio/core/core-payments/generated/glmodel"
	"github.com/openreserveio/core/core-payments/service/activities"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type EntitySanctionsScreenWorkflow struct {
	GLServiceClient       glmodel.GeneralLedgerServiceClient
	SanctionsSearchClient search.Client
	PostalURL             *url.URL
}

type SanctionScreen struct {
	Results []activities.SanctionScreenResult
}

func NewEntitySanctionsScreenWorkflow(ctx context.Context, coreGlUrl string, watchmanUrl string, postalUrl string) *EntitySanctionsScreenWorkflow {

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

	// Postal URL
	postalUrlParsed, err := url.Parse(postalUrl)
	if err != nil {
		panic(err)
	}
	sanctionsWF.PostalURL = postalUrlParsed

	return &sanctionsWF

}

func (wf *EntitySanctionsScreenWorkflow) SanctionScreenEntity(ctx workflow.Context, entityId string) (SanctionScreen, error) {

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
	ss := SanctionScreen{}

	var entity glmodel.LedgerEntity
	err := workflow.ExecuteActivity(ctx, (&activities.SanctionsScreenActivity{}).RetrieveEntity, entityId).Get(ctx, &entity)
	if err != nil {
		return ss, err
	}

	ledgerEntryMailingAddress := glmodel.LedgerEntityAddress{}
	err = workflow.ExecuteActivity(ctx, (&activities.SanctionsScreenActivity{}).AddressParse, *entity.MailingAddress).Get(ctx, &ledgerEntryMailingAddress)
	if err != nil {
		return ss, err
	}
	entity.MailingAddress = &ledgerEntryMailingAddress

	ledgerEntryBusinessAddress := glmodel.LedgerEntityAddress{}
	err = workflow.ExecuteActivity(ctx, (&activities.SanctionsScreenActivity{}).AddressParse, *entity.BusinessAddress).Get(ctx, &ledgerEntryBusinessAddress)
	if err != nil {
		return ss, err
	}
	entity.BusinessAddress = &ledgerEntryBusinessAddress

	var updatedEntity glmodel.LedgerEntity
	err = workflow.ExecuteActivity(ctx, (&activities.SanctionsScreenActivity{}).UpdateEntity, entity).Get(ctx, &updatedEntity)
	if err != nil {
		return ss, err
	}

	var sanctionScreenResults []activities.SanctionScreenResult
	err = workflow.ExecuteActivity(ctx, (&activities.SanctionsScreenActivity{}).WatchmanScreen, updatedEntity).Get(ctx, &sanctionScreenResults)
	if err != nil {
		return ss, err
	}

	ss.Results = sanctionScreenResults

	return ss, nil

}
