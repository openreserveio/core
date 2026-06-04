package cmd

import (
	"context"
	"database/sql"
	"os"

	"github.com/openreserveio/core/core-payments/service/activities"
	"github.com/openreserveio/core/core-payments/service/workflows"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
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
	sanctionScreenWorkflowWorkerCmd.PersistentFlags().StringVar(&entityDbUrl, "entitydburl", "", "Example for Postgres: postgresql://finsorbuser:finsorbpass@localhost:5432/entitydb?sslmode=disable")
	sanctionScreenWorkflowWorkerCmd.PersistentFlags().StringVar(&paymentsDbUrl, "paymentsdburl", "", "Example for Postgres: postgresql://finsorbuser:finsorbpass@localhost:5432/entitydb?sslmode=disable")
	sanctionScreenWorkflowWorkerCmd.PersistentFlags().StringVar(&accountingConfigFileLocation, "accountingconfig", "", "File path to accounting config file ./accounting_config.json")

	sanctionScreenWorkflowWorkerCmd.MarkPersistentFlagRequired("coreglurl")
	sanctionScreenWorkflowWorkerCmd.MarkPersistentFlagRequired("temporalurl")
	sanctionScreenWorkflowWorkerCmd.MarkPersistentFlagRequired("watchmanurl")
	sanctionScreenWorkflowWorkerCmd.MarkPersistentFlagRequired("postalurl")
	sanctionScreenWorkflowWorkerCmd.MarkPersistentFlagRequired("entitydburl")
	sanctionScreenWorkflowWorkerCmd.MarkPersistentFlagRequired("paymentsdburl")
	sanctionScreenWorkflowWorkerCmd.MarkPersistentFlagRequired("accountingconfig")

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

		// Create connection to EntityDB
		entityDbConn := sql.OpenDB(pgdriver.NewConnector(
			pgdriver.WithDSN(entityDbUrl),
		))
		entityDbBun := bun.NewDB(entityDbConn, pgdialect.New())

		accountingConfig, err := loadAccountingConfig(accountingConfigFileLocation)
		if err != nil {
			log.Fatalf("Failed to load accounting config: %v", err)
			os.Exit(1)
		}

		// Suspend Workflow
		suspendPaymentWorkflow := workflows.NewSuspendPaymentWorkflow(context.Background(), coreGlUrl, paymentsDbUrl, accountingConfig)
		suspendPaymentActivity := activities.NewSuspendPaymentActivity(suspendPaymentWorkflow.CoreGLClient, accountingConfig)

		// Payment wf
		paymentActivity := activities.NewPaymentActivity(suspendPaymentWorkflow.PaymentsDB, entityDbBun, suspendPaymentWorkflow.CoreGLClient)

		// Create the Temporal worker
		sanctionsActivity := activities.NewSanctionsScreenActivity(sanctionScreenWorkflow.GLServiceClient, sanctionScreenWorkflow.SanctionsSearchClient, sanctionScreenWorkflow.PostalURL)
		w := worker.New(c, workflows.TASK_QUEUE_SANCTION_SCREEN, worker.Options{})
		w.RegisterWorkflow(sanctionScreenWorkflow.SanctionScreenEntity)
		w.RegisterWorkflow(suspendPaymentWorkflow.SuspendPaymentForReview)
		w.RegisterActivity(sanctionsActivity.RetrieveEntity)
		w.RegisterActivity(sanctionsActivity.AddressParse)
		w.RegisterActivity(sanctionsActivity.UpdateEntity)
		w.RegisterActivity(sanctionsActivity.WatchmanScreen)
		w.RegisterActivity(suspendPaymentActivity.FednowSuspendPaymentLedgerEntries)
		w.RegisterActivity(paymentActivity.UpdatePaymentStatus)

		// Start the Worker
		err = w.Run(worker.InterruptCh())
		if err != nil {
			log.Fatalln("Unable to start Temporal worker", err)
		}

	},
}
