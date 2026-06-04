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

func init() {

	suspendPaymentWorkflowWorkerCmd.PersistentFlags().StringVar(&coreGlUrl, "coreglurl", "", "localhost:4081")
	suspendPaymentWorkflowWorkerCmd.PersistentFlags().StringVar(&temporalUrl, "temporalurl", "", "localhost:7233")
	suspendPaymentWorkflowWorkerCmd.PersistentFlags().StringVar(&paymentsDbUrl, "paymentsdburl", "", "Example for Postgres: postgresql://finsorbuser:finsorbpass@localhost:5432/paymentsdb?sslmode=disable")
	suspendPaymentWorkflowWorkerCmd.PersistentFlags().StringVar(&entityDbUrl, "entitydburl", "", "Example for Postgres: postgresql://finsorbuser:finsorbpass@localhost:5432/entitydb?sslmode=disable")
	suspendPaymentWorkflowWorkerCmd.PersistentFlags().StringVar(&accountingConfigFileLocation, "accountingconfig", "", "File path to accounting config file ./accounting_config.json")

	suspendPaymentWorkflowWorkerCmd.MarkPersistentFlagRequired("coreglurl")
	suspendPaymentWorkflowWorkerCmd.MarkPersistentFlagRequired("temporalurl")
	suspendPaymentWorkflowWorkerCmd.MarkPersistentFlagRequired("paymentsdburl")
	suspendPaymentWorkflowWorkerCmd.MarkPersistentFlagRequired("entitydburl")
	suspendPaymentWorkflowWorkerCmd.MarkPersistentFlagRequired("accountingconfig")

	rootCmd.AddCommand(suspendPaymentWorkflowWorkerCmd)

}

var suspendPaymentWorkflowWorkerCmd = &cobra.Command{
	Use:   "suspend-payment-wfworker",
	Short: "Start the Suspend Payment Workflow Worker Service",
	Long: `
Workflow to manage the Suspend Payment process
`,
	Run: func(cmd *cobra.Command, args []string) {

		// Load Accounting Config
		accountingConfig, err := loadAccountingConfig(accountingConfigFileLocation)
		if err != nil {
			log.Fatalf("Failed to load accounting config: %v", err)
			os.Exit(1)
		}

		log.Info("Starting Suspend Payment Workflow Service")
		suspendPaymentWorkflow := workflows.NewSuspendPaymentWorkflow(context.Background(), coreGlUrl, paymentsDbUrl, accountingConfig)

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

		// Create the Temporal worker
		sanctionsActivity := activities.NewSuspendPaymentActivity(suspendPaymentWorkflow.CoreGLClient, accountingConfig)
		paymentActivity := activities.NewPaymentActivity(suspendPaymentWorkflow.PaymentsDB, entityDbBun, suspendPaymentWorkflow.CoreGLClient)
		w := worker.New(c, workflows.TASK_QUEUE_SUSPEND_PAYMENT, worker.Options{})
		w.RegisterWorkflow(suspendPaymentWorkflow.SuspendPaymentForReview)
		w.RegisterActivity(sanctionsActivity.FednowSuspendPaymentLedgerEntries)
		w.RegisterActivity(paymentActivity.UpdatePaymentStatus)

		// Start the Worker
		err = w.Run(worker.InterruptCh())
		if err != nil {
			log.Fatalln("Unable to start Temporal worker", err)
		}

	},
}
