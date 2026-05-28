package cmd

import (
	"fmt"
	"net"
	"os"

	"github.com/openreserveio/core/core-ledger/generated/model"
	"github.com/openreserveio/core/core-ledger/service"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

var dbConnectionUrl string
var listenHost string
var listenPort string

func init() {

	startCmd.PersistentFlags().StringVar(&dbConnectionUrl, "dburl", "", "Example for Postgres: postgresql://postgres:finsorbpass@localhost:5432/coreledgerdb?sslmode=disable")
	startCmd.PersistentFlags().StringVar(&listenHost, "listenHost", "", "0.0.0.0")
	startCmd.PersistentFlags().StringVar(&listenPort, "listenPort", "", "4080")

	startCmd.MarkPersistentFlagRequired("dburl")
	startCmd.MarkPersistentFlagRequired("listenHost")
	startCmd.MarkPersistentFlagRequired("listenPort")

	rootCmd.AddCommand(startCmd)

}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Core Ledger Service",
	Long: `
Start the Core Ledger Service
`,
	Run: func(cmd *cobra.Command, args []string) {

		// Get Context
		ctx := cmd.Context()

		log.Info("Starting Core Ledger Service")
		var listenHost string = "0.0.0.0"
		var listenPort string = "4080"

		ledgerService, err := service.NewCoreLedgerService(ctx, dbConnectionUrl)
		if err != nil {
			log.Fatalf("Error starting Core Ledger Service: %v", err)
			os.Exit(1)
		}

		listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", listenHost, listenPort))
		if err != nil {
			log.Fatalf("Error listening on %s:%s: %v", listenHost, listenPort, err)
			os.Exit(1)
		}

		var grpcOpts []grpc.ServerOption
		grpcOpts = append(grpcOpts, grpc.StatsHandler(otelgrpc.NewServerHandler()))
		grpcServer := grpc.NewServer(grpcOpts...)
		model.RegisterCoreLedgerServiceServer(grpcServer, ledgerService)

		err = grpcServer.Serve(listener)
		if err != nil {
			log.Fatalf("Ledger Service Not Serving Clients: %v", err)
			os.Exit(1)
		}
	},
}
