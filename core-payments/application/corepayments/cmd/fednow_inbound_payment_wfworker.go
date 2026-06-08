package cmd

import (
	"context"
	"encoding/json"
	"os"

	"github.com/openreserveio/core/core-payments/pmtmodel"
	"github.com/openreserveio/core/core-payments/service/activities"
	"github.com/openreserveio/core/core-payments/service/workflows"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

var paymentsDbUrl string
var entityDbUrl string

// var dbConnectionUrl string
// var busConnUrl string
// var listenHost string
// var listenPort string
var redisUrl string
var coreGlUrl string
var accountingConfigFileLocation string
var temporalUrl string

func init() {

	fednowInboundPaymentWorkflowWorkerCmd.PersistentFlags().StringVar(&accountingConfigFileLocation, "accountingconfig", "", "File path to accounting config file ./accounting_config.json")
	fednowInboundPaymentWorkflowWorkerCmd.PersistentFlags().StringVar(&coreGlUrl, "coreglurl", "", "localhost:4081")
	fednowInboundPaymentWorkflowWorkerCmd.PersistentFlags().StringVar(&entityDbUrl, "entitydburl", "", "Example for Postgres: postgresql://finsorbuser:finsorbpass@localhost:5432/entitydb?sslmode=disable")
	fednowInboundPaymentWorkflowWorkerCmd.PersistentFlags().StringVar(&paymentsDbUrl, "paymentsdburl", "", "Example for Postgres: postgresql://finsorbuser:finsorbpass@localhost:5432/paymentsdb?sslmode=disable")
	fednowInboundPaymentWorkflowWorkerCmd.PersistentFlags().StringVar(&busConnUrl, "busconnurl", "", "nats://localhost:4322")
	fednowInboundPaymentWorkflowWorkerCmd.PersistentFlags().StringVar(&redisUrl, "redisurl", "", "localhost:6379")
	fednowInboundPaymentWorkflowWorkerCmd.PersistentFlags().StringVar(&temporalUrl, "temporalurl", "", "localhost:7233")
	fednowInboundPaymentWorkflowWorkerCmd.PersistentFlags().StringVar(&listenHost, "listenHost", "", "0.0.0.0")
	fednowInboundPaymentWorkflowWorkerCmd.PersistentFlags().StringVar(&listenPort, "listenPort", "", "4083")

	fednowInboundPaymentWorkflowWorkerCmd.MarkPersistentFlagRequired("accountingconfig")
	fednowInboundPaymentWorkflowWorkerCmd.MarkPersistentFlagRequired("coreglurl")
	fednowInboundPaymentWorkflowWorkerCmd.MarkPersistentFlagRequired("entitydburl")
	fednowInboundPaymentWorkflowWorkerCmd.MarkPersistentFlagRequired("paymentsdburl")
	fednowInboundPaymentWorkflowWorkerCmd.MarkPersistentFlagRequired("busconnurl")
	fednowInboundPaymentWorkflowWorkerCmd.MarkPersistentFlagRequired("temporalurl")
	fednowInboundPaymentWorkflowWorkerCmd.MarkPersistentFlagRequired("redisurl")
	fednowInboundPaymentWorkflowWorkerCmd.MarkPersistentFlagRequired("listenHost")
	fednowInboundPaymentWorkflowWorkerCmd.MarkPersistentFlagRequired("listenPort")

	rootCmd.AddCommand(fednowInboundPaymentWorkflowWorkerCmd)

}

var fednowInboundPaymentWorkflowWorkerCmd = &cobra.Command{
	Use:   "fednow-inbound-payment-wfworker",
	Short: "Start the Fednow Inbound Payment Workflow Worker Service",
	Long: `
Workflow to manage the inbound processing of Fednow messages
`,
	Run: func(cmd *cobra.Command, args []string) {

		log.Info("Starting Fednow Inbound Payment Workflow Service")
		accountingConfig, err := loadAccountingConfig(accountingConfigFileLocation)
		if err != nil {
			log.Fatalf("Failed to load accounting config: %v", err)
			os.Exit(1)
		}
		fednowInboundWorkflow := workflows.NewFednowInboundPaymentWorkflow(context.Background(), paymentsDbUrl, entityDbUrl, coreGlUrl, accountingConfig)

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

		// Payment Processing Workflow
		ppWorkflow := workflows.NewPaymentProcessingWorkflow(paymentsDbUrl, entityDbUrl, coreGlUrl, accountingConfig)

		// Create the Temporal worker
		paymentActivity := activities.NewPaymentActivity(fednowInboundWorkflow.PaymentsDB, fednowInboundWorkflow.EntityDB, fednowInboundWorkflow.GLServiceClient)
		w := worker.New(c, workflows.TASK_QUEUE_FEDNOW_INBOUND_PAYMENT, worker.Options{})
		w.RegisterWorkflow(fednowInboundWorkflow.ProcessFednowInboundPayment)
		w.RegisterWorkflow(suspendPaymentWorkflow.SuspendPaymentForReview)
		w.RegisterWorkflow(ppWorkflow.ProcessPayment)
		w.RegisterActivity(paymentActivity.StoreRawPaymentInstruction)
		w.RegisterActivity(paymentActivity.ValidatePaymentInstruction)
		w.RegisterActivity(paymentActivity.UpdatePaymentStatus)
		w.RegisterActivity(paymentActivity.FednowInitialLedgerEntries)

		// Start the Worker
		err = w.Run(worker.InterruptCh())
		if err != nil {
			log.Fatalln("Unable to start Temporal worker", err)
		}

	},
}

func loadAccountingConfig(configPath string) (*pmtmodel.AccountingConfig, error) {

	contents, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config pmtmodel.AccountingConfig
	err = json.Unmarshal(contents, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil

}
