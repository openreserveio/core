package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/openreserveio/core/core-ledger/otel"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
}

var rootCmd = &cobra.Command{
	Use:   "coreledger",
	Short: "coreledger root command",
	Long:  `coreledger Collection of Services`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Specify a Subcommand")
		os.Exit(1)
	},
}

func Execute(ctx context.Context) {

	otelExporter := otel.NewExporter(ctx, otel.EXPORTER_TYPE_OTLP)
	otel.NewTracerProvider("core-ledger", otelExporter)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
