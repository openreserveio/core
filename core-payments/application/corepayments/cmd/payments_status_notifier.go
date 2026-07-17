package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var paymentStatusNotifierCmd = &cobra.Command{
	Use:   "payments-status-notifier",
	Short: "Start the Payments Status Notifier",
	Long: `
Notifies interested parties about their payment status changes'
`,
	Run: func(cmd *cobra.Command, args []string) {

		log.Info("Starting Payments Status Notifier")

	},
}
