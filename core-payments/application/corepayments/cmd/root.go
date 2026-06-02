package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/openreserveio/core/core-util/otel"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
}

var rootCmd = &cobra.Command{
	Use:   "corepayments",
	Short: "corepayments root command",
	Long:  `corepayments Collection of Services`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Specify a Subcommand")
		os.Exit(1)
	},
}

func Execute(ctx context.Context) {

	otelExporter := otel.NewExporter(ctx, otel.EXPORTER_TYPE_OTLP)
	otel.NewTracerProvider("core-payments-fednow-inbound-payment-wfworker", otelExporter)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
