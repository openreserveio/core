package cmd

import (
	"context"
	"os"

	"github.com/openreserveio/core/core-payments/service/connector"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var defaultLedgerId string
var busConnUrl string
var listenHost string
var listenPort string
var isFintechProgramsAware bool
var ftpServiceUrl string

func init() {

	fednowConnectorCmd.PersistentFlags().StringVar(&defaultLedgerId, "defaultledgerid", "", "")
	fednowConnectorCmd.PersistentFlags().StringVar(&paymentsDbUrl, "paymentsdburl", "", "Example for Postgres: postgresql://finsorbuser:finsorbpass@localhost:5432/paymentsdb?sslmode=disable")
	fednowConnectorCmd.PersistentFlags().StringVar(&busConnUrl, "busconnurl", "", "nats://localhost:4322")
	fednowConnectorCmd.PersistentFlags().StringVar(&listenHost, "listenHost", "", "0.0.0.0")
	fednowConnectorCmd.PersistentFlags().StringVar(&listenPort, "listenPort", "", "4083")
	fednowConnectorCmd.PersistentFlags().StringVar(&redisUrl, "redisurl", "", "localhost:6379")
	fednowConnectorCmd.PersistentFlags().BoolVar(&isFintechProgramsAware, "ftpaware", false, "true or false, indicates whether the connector should be aware of Fintech Programs")
	fednowConnectorCmd.PersistentFlags().StringVar(&ftpServiceUrl, "ftpserviceurl", "", "localhost:6379")

	fednowConnectorCmd.MarkPersistentFlagRequired("defaultledgerid")
	fednowConnectorCmd.MarkPersistentFlagRequired("paymentsdburl")
	fednowConnectorCmd.MarkPersistentFlagRequired("busconnurl")
	fednowConnectorCmd.MarkPersistentFlagRequired("redisurl")
	fednowConnectorCmd.MarkPersistentFlagRequired("listenPort")
	fednowConnectorCmd.MarkPersistentFlagRequired("listenPort")

	rootCmd.AddCommand(fednowConnectorCmd)

}

var fednowConnectorCmd = &cobra.Command{
	Use:   "fednow-connector",
	Short: "Start the Fednow Payments Connector Service",
	Long: `
Connects to the Fednow Gateway and listens for payments events, while also sending payment instructions outbound
`,
	Run: func(cmd *cobra.Command, args []string) {

		log.Info("Starting Fednow Connector Service")
		config := connector.FedNowConnectorConfig{
			NatsUrl:                  busConnUrl,
			ListenHost:               listenHost,
			ListenPort:               listenPort,
			PaymentsDBUrl:            paymentsDbUrl,
			InboundStreamName:        "FEDNOWIN",
			InboundSubject:           "inbound",
			OutboundStreamName:       "FEDNOWOUT",
			OutboundSubject:          "outbound",
			RedisUrl:                 redisUrl,
			IsFintechProgramAware:    isFintechProgramsAware,
			FintechProgramServiceUrl: ftpServiceUrl,
			DefaultLedgerId:          defaultLedgerId,
		}

		fednowConnector, err := connector.NewFedNowConnector(context.Background(), &config)
		if err != nil {
			log.Error("Error creating Fednow Connector: %v", err)
			os.Exit(1)
		}

		if err := fednowConnector.Start(); err != nil {
			log.Error("Fednow Connector Shutting Down: %v", err)
		}

	},
}
