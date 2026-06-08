package cmd

import (
	"context"
	"os"

	"github.com/openreserveio/core/core-payments/service/activities"
	"github.com/openreserveio/core/core-payments/service/workflows"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func init() {

	inboundPaymentProcessingWorkflowWorkerCmd.PersistentFlags().StringVar(&accountingConfigFileLocation, "accountingconfig", "", "File path to accounting config file ./accounting_config.json")
	inboundPaymentProcessingWorkflowWorkerCmd.PersistentFlags().StringVar(&coreGlUrl, "coreglurl", "", "localhost:4081")
	inboundPaymentProcessingWorkflowWorkerCmd.PersistentFlags().StringVar(&entityDbUrl, "entitydburl", "", "Example for Postgres: postgresql://finsorbuser:finsorbpass@localhost:5432/entitydb?sslmode=disable")
	inboundPaymentProcessingWorkflowWorkerCmd.PersistentFlags().StringVar(&paymentsDbUrl, "paymentsdburl", "", "Example for Postgres: postgresql://finsorbuser:finsorbpass@localhost:5432/paymentsdb?sslmode=disable")
	inboundPaymentProcessingWorkflowWorkerCmd.PersistentFlags().StringVar(&busConnUrl, "busconnurl", "", "nats://localhost:4322")
	inboundPaymentProcessingWorkflowWorkerCmd.PersistentFlags().StringVar(&redisUrl, "redisurl", "", "localhost:6379")
	inboundPaymentProcessingWorkflowWorkerCmd.PersistentFlags().StringVar(&temporalUrl, "temporalurl", "", "localhost:7233")
	inboundPaymentProcessingWorkflowWorkerCmd.PersistentFlags().StringVar(&listenHost, "listenHost", "", "0.0.0.0")
	inboundPaymentProcessingWorkflowWorkerCmd.PersistentFlags().StringVar(&listenPort, "listenPort", "", "4083")

	inboundPaymentProcessingWorkflowWorkerCmd.MarkPersistentFlagRequired("accountingconfig")
	inboundPaymentProcessingWorkflowWorkerCmd.MarkPersistentFlagRequired("coreglurl")
	inboundPaymentProcessingWorkflowWorkerCmd.MarkPersistentFlagRequired("entitydburl")
	inboundPaymentProcessingWorkflowWorkerCmd.MarkPersistentFlagRequired("paymentsdburl")
	inboundPaymentProcessingWorkflowWorkerCmd.MarkPersistentFlagRequired("busconnurl")
	inboundPaymentProcessingWorkflowWorkerCmd.MarkPersistentFlagRequired("temporalurl")
	inboundPaymentProcessingWorkflowWorkerCmd.MarkPersistentFlagRequired("redisurl")

	rootCmd.AddCommand(inboundPaymentProcessingWorkflowWorkerCmd)

}

var inboundPaymentProcessingWorkflowWorkerCmd = &cobra.Command{
	Use:   "inbound-payment-processing-wfworker",
	Short: "Start the Inbound Payment Processing Workflow Worker Service",
	Long: `
Workflow to manage the inbound processing of payments
`,
	Run: func(cmd *cobra.Command, args []string) {

		log.Info("Starting Inbound Payment Processing Workflow Service")
		accountingConfig, err := loadAccountingConfig(accountingConfigFileLocation)
		if err != nil {
			log.Fatalf("Failed to load accounting config: %v", err)
			os.Exit(1)
		}
		ppWorkflow := workflows.NewPaymentProcessingWorkflow(paymentsDbUrl, entityDbUrl, coreGlUrl, accountingConfig)

		// Create the Temporal client
		c, err := client.Dial(client.Options{
			HostPort: temporalUrl,
		})
		if err != nil {
			log.Fatalln("Unable to create Temporal client", err)
		}
		defer c.Close()

		// Suspend Workflow
		suspendPaymentWorkflow := workflows.NewSuspendPaymentWorkflow(context.Background(), coreGlUrl, paymentsDbUrl, accountingConfig)

		// Payments Activities
		paymentActivity := activities.NewPaymentActivity(ppWorkflow.PaymentsDB, ppWorkflow.EntityDB, ppWorkflow.GLServiceClient)

		// Create the Temporal worker
		w := worker.New(c, workflows.TASK_QUEUE_INBOUND_PAYMENT_PROCESSING, worker.Options{})
		w.RegisterWorkflow(ppWorkflow.ProcessPayment)
		w.RegisterWorkflow(suspendPaymentWorkflow.SuspendPaymentForReview)
		w.RegisterActivity(paymentActivity.UpdatePaymentStatus)
		w.RegisterActivity(paymentActivity.ProcessEntities)
		w.RegisterActivity(paymentActivity.GetTransactionMonitoringRisk)

		// Start the Worker
		err = w.Run(worker.InterruptCh())
		if err != nil {
			log.Fatalln("Unable to start Temporal worker", err)
		}

	},
}
