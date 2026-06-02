package connector

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/openreserveio/core/core-payments/generated/ftpmodel"
	"github.com/openreserveio/core/core-payments/service/workflows"
	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
	"go.temporal.io/sdk/client"
)

type FednowMessage struct {
	Error           error  `json:"error" xml:"error"`
	Direction       string `json:"direction" xml:"direction"`
	MessageContents []byte `json:"message_contents" xml:"message_contents"`
}

type FedNowConnectorConfig struct {
	InboundStreamName  string
	InboundSubject     string
	OutboundStreamName string
	OutboundSubject    string
	NatsUrl            string
	TemporalUrl        string
}

type FedNowConnector struct {
	Queue                jetstream.JetStream
	PaymentsDB           *bun.DB
	Config               *FedNowConnectorConfig
	Signal               chan FednowMessage
	FintechProgramClient ftpmodel.FintechProgramServiceClient
	TemporalClient       client.Client
}

func NewFedNowConnector(ctx context.Context, config *FedNowConnectorConfig) (*FedNowConnector, error) {

	connector := FedNowConnector{
		Config: config,
		Signal: make(chan FednowMessage, 1),
	}

	// Set up Temporal
	temporalClient, err := client.Dial(client.Options{
		HostPort: config.TemporalUrl,
	})
	if err != nil {
		return nil, err
	}
	connector.TemporalClient = temporalClient

	// Setup Jetstream
	nc, err := nats.Connect(config.NatsUrl)
	if err != nil {
		return nil, err
	}
	js, err := jetstream.New(nc)
	if err != nil {
		return nil, err
	}
	connector.Queue = js

	return &connector, nil

}

func (fc *FedNowConnector) Start() error {

	// Context
	ctx := context.Background()

	// Spin up Jetstream listeners
	go fc.startJetstreamListeners(ctx, fc.Signal)

	for {

		log.Infof("Waiting for bidirectional messages")
		msg := <-fc.Signal
		switch msg.Direction {
		case "INBOUND":
			log.Infof("Received inbound message, kicking off workflow")
			wfrun, err := fc.TemporalClient.ExecuteWorkflow(ctx, client.StartWorkflowOptions{TaskQueue: "fednow-inbound-payment-queue"}, (&workflows.FednowInboundPaymentWorkflow{}).ProcessFednowInboundPayment, msg.MessageContents)
			if err != nil {
				log.Errorf("Error starting workflow: %v", err)
				continue
			}
			log.Infof("Workflow started: %s", wfrun.GetID())

		case "OUTBOUND":
			log.Infof("Received outbound message.  Routing to Jetstream Stream!")
			subj := fmt.Sprintf("%s.%s", fc.Config.OutboundStreamName, fc.Config.OutboundSubject)
			_, err := fc.Queue.Publish(ctx, subj, msg.MessageContents)
			if err != nil {
				log.Errorf("Error publishing message to Jetstream: %v", err)
				continue
			}

		}

	}

}

func (fc *FedNowConnector) HandleOutboundPost(ctx *gin.Context) {

	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		log.Errorf("Error reading body from HTTP Post: %v", err)
		fc.Signal <- FednowMessage{Error: err}
	}

	fc.Signal <- FednowMessage{Direction: "OUTBOUND", MessageContents: body}

}

func (fc *FedNowConnector) startJetstreamListeners(ctx context.Context, signal chan FednowMessage) {

	// fullSubjectInbound := fmt.Sprintf("%s.%s.*", fc.Config.StreamName, fc.Config.InboundSubject)
	// fullSubjectOutbound := fmt.Sprintf("%s.%s.*", fc.Config.StreamName, fc.Config.OutboundSubject)
	queueStreamIn, err := fc.Queue.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     fc.Config.InboundStreamName,
		Subjects: []string{fmt.Sprintf("%s.*", fc.Config.InboundStreamName)},
	})
	if err != nil {
		log.Fatalf("Failed to create or update inbound stream: %v", err)
		os.Exit(1)
	}

	_, err = fc.Queue.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     fc.Config.OutboundStreamName,
		Subjects: []string{fmt.Sprintf("%s.*", fc.Config.OutboundStreamName)},
	})
	if err != nil {
		log.Fatalf("Failed to create or update outbound stream: %v", err)
		os.Exit(1)
	}

	queueConsumer, err := queueStreamIn.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:   "fednow-connector",
		AckPolicy: jetstream.AckExplicitPolicy,
	})
	if err != nil {
		log.Fatalf("Failed to create or update consumer: %v", err)
		os.Exit(1)
	}

	for {

		log.Info("Waiting for message...")
		msg, err := queueConsumer.Next(jetstream.FetchMaxWait(1 * time.Minute))
		if err != nil {
			log.Errorf("Encountered error while reading message: %v", err)
			fednowMsg := FednowMessage{
				Error: err,
			}
			signal <- fednowMsg
			continue
		}

		messageData := msg.Data()
		if messageData == nil {
			msg.Ack()
			continue
		}

		log.Info("Received message: %v", msg.Headers())
		fednowMsg := FednowMessage{
			Direction:       "INBOUND",
			MessageContents: messageData,
		}
		signal <- fednowMsg
		msg.Ack()

	}

}
