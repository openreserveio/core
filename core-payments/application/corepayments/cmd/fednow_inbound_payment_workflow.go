package cmd

import (
	"encoding/json"
	"os"

	"github.com/openreserveio/core/core-payments/pmtmodel"
	"github.com/openreserveio/core/core-payments/service/workflows/fednow"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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

func init() {

	fednowInboundPaymentWorkflowCmd.PersistentFlags().StringVar(&accountingConfigFileLocation, "accountingconfig", "", "File path to accounting config file ./accounting_config.json")
	fednowInboundPaymentWorkflowCmd.PersistentFlags().StringVar(&coreGlUrl, "coreglurl", "", "localhost:4081")
	fednowInboundPaymentWorkflowCmd.PersistentFlags().StringVar(&entityDbUrl, "entitydburl", "", "Example for Postgres: postgresql://finsorbuser:finsorbpass@localhost:5432/entitydb?sslmode=disable")
	fednowInboundPaymentWorkflowCmd.PersistentFlags().StringVar(&paymentsDbUrl, "paymentsdburl", "", "Example for Postgres: postgresql://finsorbuser:finsorbpass@localhost:5432/paymentsdb?sslmode=disable")
	fednowInboundPaymentWorkflowCmd.PersistentFlags().StringVar(&busConnUrl, "busconnurl", "", "nats://localhost:4322")
	fednowInboundPaymentWorkflowCmd.PersistentFlags().StringVar(&redisUrl, "redisurl", "", "localhost:6379")
	fednowInboundPaymentWorkflowCmd.PersistentFlags().StringVar(&listenHost, "listenHost", "", "0.0.0.0")
	fednowInboundPaymentWorkflowCmd.PersistentFlags().StringVar(&listenPort, "listenPort", "", "4083")

	fednowInboundPaymentWorkflowCmd.MarkPersistentFlagRequired("accountingconfig")
	fednowInboundPaymentWorkflowCmd.MarkPersistentFlagRequired("coreglurl")
	fednowInboundPaymentWorkflowCmd.MarkPersistentFlagRequired("entitydburl")
	fednowInboundPaymentWorkflowCmd.MarkPersistentFlagRequired("paymentsdburl")
	fednowInboundPaymentWorkflowCmd.MarkPersistentFlagRequired("busconnurl")
	fednowInboundPaymentWorkflowCmd.MarkPersistentFlagRequired("redisurl")
	fednowInboundPaymentWorkflowCmd.MarkPersistentFlagRequired("listenHost")
	fednowInboundPaymentWorkflowCmd.MarkPersistentFlagRequired("listenPort")

	rootCmd.AddCommand(fednowInboundPaymentWorkflowCmd)

}

var fednowInboundPaymentWorkflowCmd = &cobra.Command{
	Use:   "fednow-inbound-payment-workflow",
	Short: "Start the Fednow Inbound Payment Workflow Service",
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
		fednowInboundWorkflow := fednow.NewFednowInboundPaymentWorkflow(listenHost, listenPort, redisUrl, paymentsDbUrl, entityDbUrl, coreGlUrl, accountingConfig)
		err = fednowInboundWorkflow.Start()
		if err != nil {
			log.Fatalf("Failed to start Fednow Inbound Payment Workflow Service: %v", err)
			log.Fatal(err)
			os.Exit(1)
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
