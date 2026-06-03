package cmd

import (
	"context"

	"github.com/openreserveio/core/core-payments/service/activities"
	"github.com/openreserveio/core/core-payments/service/workflows"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

var watchmanUrl string
var postalUrl string

func init() {

	sanctionScreenWorkflowWorkerCmd.PersistentFlags().StringVar(&coreGlUrl, "coreglurl", "", "localhost:4081")
	sanctionScreenWorkflowWorkerCmd.PersistentFlags().StringVar(&temporalUrl, "temporalurl", "", "localhost:7233")
	sanctionScreenWorkflowWorkerCmd.PersistentFlags().StringVar(&watchmanUrl, "watchmanurl", "", "localhost:7233")
	sanctionScreenWorkflowWorkerCmd.PersistentFlags().StringVar(&postalUrl, "postalurl", "", "http://localhost:7233/parser")

	sanctionScreenWorkflowWorkerCmd.MarkPersistentFlagRequired("coreglurl")
	sanctionScreenWorkflowWorkerCmd.MarkPersistentFlagRequired("temporalurl")
	sanctionScreenWorkflowWorkerCmd.MarkPersistentFlagRequired("watchmanurl")
	sanctionScreenWorkflowWorkerCmd.MarkPersistentFlagRequired("postalurl")

	rootCmd.AddCommand(sanctionScreenWorkflowWorkerCmd)

}

var sanctionScreenWorkflowWorkerCmd = &cobra.Command{
	Use:   "sanction-screen-wfworker",
	Short: "Start the Sanction Screen Workflow Worker Service",
	Long: `
Workflow to manage the sanctions screen process
`,
	Run: func(cmd *cobra.Command, args []string) {

		log.Info("Starting Sanction Screen Workflow Service")
		sanctionScreenWorkflow := workflows.NewEntitySanctionsScreenWorkflow(context.Background(), coreGlUrl, watchmanUrl, postalUrl)

		// Create the Temporal client
		c, err := client.Dial(client.Options{
			HostPort: temporalUrl,
		})
		if err != nil {
			log.Fatalln("Unable to create Temporal client", err)
		}
		defer c.Close()

		// Create the Temporal worker
		sanctionsActivity := activities.NewSanctionsScreenActivity(sanctionScreenWorkflow.GLServiceClient, sanctionScreenWorkflow.SanctionsSearchClient, sanctionScreenWorkflow.PostalURL)
		w := worker.New(c, "sanction-screen-queue", worker.Options{})
		w.RegisterWorkflow(sanctionScreenWorkflow.SanctionScreenEntity)
		w.RegisterActivity(sanctionsActivity.RetrieveEntity)
		w.RegisterActivity(sanctionsActivity.AddressParse)
		w.RegisterActivity(sanctionsActivity.UpdateEntity)

		// Start the Worker
		err = w.Run(worker.InterruptCh())
		if err != nil {
			log.Fatalln("Unable to start Temporal worker", err)
		}

	},
}
