package service

import (
	"context"
	"fmt"
	"os"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/nats-io/nats.go/micro"
	"github.com/openreserveio/core/core-ledger-poster/application"
	"github.com/openreserveio/core/core-ledger-poster/generated/model"
	"github.com/openreserveio/core/core-util/bus"
	"github.com/openreserveio/core/core-util/otel"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type CoreLedgerPosterService struct {
	LedgerClient    model.CoreLedgerServiceClient
	BusConn         *bus.BusConnection
	BusConnUrl      string
	JetstreamClient jetstream.JetStream
}

func NewCoreLedgerPosterService(ctx context.Context, coreLedgerUrl string, busConnUrl string) (*CoreLedgerPosterService, error) {

	ctx = otel.StartSpan(ctx, "CoreLedgerPosterService.NewCoreLedgerPosterService")
	defer otel.EndSpan(ctx)

	clpService := CoreLedgerPosterService{}
	clpService.BusConnUrl = busConnUrl

	// Setup Core Ledger Service
	conn, err := grpc.NewClient(coreLedgerUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Unable to connect to Core Ledger: %v", err)
		os.Exit(1)
	}
	clpService.LedgerClient = model.NewCoreLedgerServiceClient(conn)

	// Setup Nats Service
	// Bus Connection
	busConn, err := bus.NewBusConnection(ctx, busConnUrl)
	if err != nil {
		log.Error("Error creating bus connection: %v", err)
		return nil, err
	}
	clpService.BusConn = busConn

	// Define the service
	srv, err := micro.AddService(busConn.Conn, micro.Config{
		Name:        application.SERVICE_NAME_CORE_LEDGER_POSTER,
		Version:     "1.0.0",
		Description: "Core Ledger Poster Service",
		Metadata:    map[string]string{"label": "Core Ledger Poster Service", "environment": "v1.0.0-DEV"},
	})
	if err != nil {
		log.Fatal("Unable to define NATS service:  %v", err)
		return nil, err
	}
	log.Info(fmt.Sprintf("Created service: %s (%s)\n", srv.Info().Name, srv.Info().ID))

	rootGroup := srv.AddGroup(application.SERVICE_NAME_CORE_LEDGER_POSTER)
	rootGroup.AddEndpoint(application.SERVICE_ENDPOINT_POST_TRANSACTION, micro.HandlerFunc(clpService.ProcessMessage))

	return &clpService, nil
}

func (clps *CoreLedgerPosterService) Start(ctx context.Context) error {

	bus.DoMonitoringLoop(clps.BusConn.Conn, clps.BusConnUrl)
	return nil
}

func (clps *CoreLedgerPosterService) ProcessMessage(request micro.Request) {

	ctx := otel.ExtractNatsContext(request)
	ctx = otel.StartSpan(ctx, "CoreLedgerPosterService.ProcessMessage")
	defer otel.EndSpan(ctx)

	log.Infof("Processing message")
	otel.AddEvent("Processing message")
	var postLedgerTxRequest model.PostLedgerTransactionRequest
	err := bus.Receive(request, &postLedgerTxRequest)
	if err != nil {
		log.Errorf("Error processing post ledger transaction request: %v", err)
		otel.AddError("Error processing post ledger transaction request", err)
		bus.ReplyWithBadRequestError(ctx, request, err)
		return
	}

	log.Infof("Posting ledger transaction")
	otel.AddEvent("Posting ledger transaction")
	response, err := clps.LedgerClient.PostLedgerTransaction(ctx, &postLedgerTxRequest)
	if err != nil {
		otel.AddError("Call to Core Ledger resulted in an error", err)
		log.Fatalf("Call to Core Ledger resulted in an error:  %v", err)
		bus.ReplyWithSystemError(ctx, request, err)
		return
	}

	otel.AddEvent("Ledger transaction posted, reply OK")
	bus.ReplyOK(ctx, request, response)

}
