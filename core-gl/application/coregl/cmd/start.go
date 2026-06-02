package cmd

import (
	"fmt"
	"net"
	"os"

	"github.com/openreserveio/core/core-gl/generated/glmodel"
	"github.com/openreserveio/core/core-gl/service"
	"github.com/openreserveio/core/core-util/otel"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var entityDbUrl string
var busConnUrl string
var coreLedgerUrl string
var listenHost string
var listenPort string

func init() {

	startCmd.PersistentFlags().StringVar(&entityDbUrl, "entitydburl", "", "Example for Postgres: postgresql://finsorbuser:finsorbpass@localhost:5432/entitydb?sslmode=disable")
	startCmd.PersistentFlags().StringVar(&busConnUrl, "busconnurl", "", "nats://localhost:4322")
	startCmd.PersistentFlags().StringVar(&coreLedgerUrl, "coreledgerurl", "", "localhost:4080")
	startCmd.PersistentFlags().StringVar(&listenHost, "listenHost", "", "0.0.0.0")
	startCmd.PersistentFlags().StringVar(&listenPort, "listenPort", "", "4080")

	startCmd.MarkPersistentFlagRequired("entitydburl")
	startCmd.MarkPersistentFlagRequired("busconnurl")
	startCmd.MarkPersistentFlagRequired("coreledgerurl")
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

		log.Info("Starting Core GL Service")
		coreGlService, err := service.NewCoreGLService(entityDbUrl, busConnUrl, coreLedgerUrl, listenHost, listenPort)
		if err != nil {
			log.Fatalf("Error starting Core GL Service: %v", err)
			os.Exit(1)
		}

		listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", listenHost, listenPort))
		if err != nil {
			log.Fatalf("Error listening on %s:%s: %v", listenHost, listenPort, err)
			os.Exit(1)
		}

		var grpcOpts []grpc.ServerOption
		grpcOpts = append(grpcOpts, otel.InjectServerGRPCHeaders())
		grpcServer := grpc.NewServer(grpcOpts...)
		glmodel.RegisterGeneralLedgerServiceServer(grpcServer, coreGlService)

		err = grpcServer.Serve(listener)
		if err != nil {
			log.Fatalf("GL Service Not Serving Clients: %v", err)
			os.Exit(1)
		}
	},
}
