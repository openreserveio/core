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

	fednowConnectorCmd.PersistentFlags().StringVar(&busConnUrl, "busconnurl", "", "nats://localhost:4322")
	fednowConnectorCmd.PersistentFlags().StringVar(&ftpServiceUrl, "ftpserviceurl", "", "localhost:6379")
	fednowConnectorCmd.PersistentFlags().StringVar(&temporalUrl, "temporalurl", "", "localhost:7233")

	fednowConnectorCmd.MarkPersistentFlagRequired("busconnurl")

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
			NatsUrl:            busConnUrl,
			InboundStreamName:  "FEDNOWIN",
			InboundSubject:     "inbound",
			OutboundStreamName: "FEDNOWOUT",
			OutboundSubject:    "outbound",
			TemporalUrl:        temporalUrl,
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
