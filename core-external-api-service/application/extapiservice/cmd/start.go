package cmd

import (
	"os"

	"github.com/openreserveio/core/core-external-api-service/service"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var busConnUrl string
var coreLedgerUrl string
var glUrl string
var httpListenHost string
var httpListenPort string

func init() {

	startCmd.PersistentFlags().StringVar(&busConnUrl, "busconnurl", "", "nats://localhost:4322")
	startCmd.PersistentFlags().StringVar(&coreLedgerUrl, "coreledgerurl", "", "localhost:4080")
	startCmd.PersistentFlags().StringVar(&glUrl, "glurl", "", "localhost:4081")
	startCmd.PersistentFlags().StringVar(&httpListenHost, "httpListenHost", "", "0.0.0.0")
	startCmd.PersistentFlags().StringVar(&httpListenPort, "httpListenPort", "", "8080")

	startCmd.MarkPersistentFlagRequired("busconnurl")
	startCmd.MarkPersistentFlagRequired("coreledgerurl")
	startCmd.MarkPersistentFlagRequired("glurl")
	startCmd.MarkPersistentFlagRequired("httpListenHost")
	startCmd.MarkPersistentFlagRequired("httpListenPort")

	rootCmd.AddCommand(startCmd)

}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Core External API Service",
	Long: `
Start the Core External API Service
`,
	Run: func(cmd *cobra.Command, args []string) {

		log.Info("Starting External API Service")
		externalApiService, err := service.NewCoreExternalApiService(httpListenHost, httpListenPort, coreLedgerUrl, glUrl)
		if err != nil {
			log.Fatalf("Unable to create External API Service: %v", err)
			os.Exit(1)
		}
		externalApiService.Start()

	},
}
