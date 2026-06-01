package cmd

import (
	"context"
	"os"

	"github.com/openreserveio/core/core-ledger-poster/service"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var natsConnectionUrl string
var coreLedgerServiceConnectionUrl string

//var listenHost string
//var listenPort string

func init() {

	startCmd.PersistentFlags().StringVar(&natsConnectionUrl, "natsurl", "", "NATS Connection URL")
	startCmd.PersistentFlags().StringVar(&coreLedgerServiceConnectionUrl, "coreledgerurl", "", "0.0.0.0:4080")

	startCmd.MarkPersistentFlagRequired("natsurl")
	startCmd.MarkPersistentFlagRequired("coreledgerurl")

	rootCmd.AddCommand(startCmd)

}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Core Ledger Service",
	Long: `
Start the Core Ledger Service
`,
	Run: func(cmd *cobra.Command, args []string) {

		ctx := context.Background()

		log.Info("Starting Core Ledger Poster Service")
		clps, err := service.NewCoreLedgerPosterService(ctx, coreLedgerServiceConnectionUrl, natsConnectionUrl)
		if err != nil {
			log.Fatalf("Unable to start Core Ledger Poster Service: %v", err)
			os.Exit(1)
		}

		err = clps.Start(ctx)
		if err != nil {
			log.Fatalf("Core Ledger Poster Service shut down with error: %v", err)
			return
		}

		log.Info("Shutting down Core Ledger Poster Service")

	},
}
