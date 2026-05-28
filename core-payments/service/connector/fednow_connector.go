package connector

import (
	"context"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/moov-io/fednow20022/gen/pacs_008_001_08"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/openreserveio/core/core-payments/generated/ftpmodel"
	"github.com/openreserveio/core/core-payments/pmtmodel"
	goflow "github.com/s8sg/goflow/v1"
	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type FednowMessage struct {
	Error           error  `json:"error" xml:"error"`
	Direction       string `json:"direction" xml:"direction"`
	MessageContents []byte `json:"message_contents" xml:"message_contents"`
}

type FedNowConnectorConfig struct {
	InboundStreamName        string
	InboundSubject           string
	OutboundStreamName       string
	OutboundSubject          string
	PaymentsDBUrl            string
	NatsUrl                  string
	ListenHost               string
	ListenPort               string
	RedisUrl                 string
	IsFintechProgramAware    bool
	FintechProgramServiceUrl string
	DefaultLedgerId          string
}

type FedNowConnector struct {
	Queue                jetstream.JetStream
	PaymentsDB           *bun.DB
	Gin                  *gin.Engine
	Config               *FedNowConnectorConfig
	Signal               chan FednowMessage
	FintechProgramClient ftpmodel.FintechProgramServiceClient
}

func NewFedNowConnector(ctx context.Context, config *FedNowConnectorConfig) (*FedNowConnector, error) {

	connector := FedNowConnector{
		Config: config,
		Signal: make(chan FednowMessage, 1),
	}

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

	// Payments EntityDB
	// EntityDB Connection Setup
	// Using pgdriver (recommended)
	dbConn := sql.OpenDB(pgdriver.NewConnector(
		pgdriver.WithDSN(config.PaymentsDBUrl),
	))
	dbBun := bun.NewDB(dbConn, pgdialect.New())
	connector.PaymentsDB = dbBun

	// Fintech Program Service Client
	if config.IsFintechProgramAware {
		log.Infof("Setting up Fintech Program Service Client: %v", config.FintechProgramServiceUrl)
		conn, err := grpc.NewClient(config.FintechProgramServiceUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("Unable to connect to Core Fintech Program Service: %v", err)
			os.Exit(1)
		}
		connector.FintechProgramClient = ftpmodel.NewFintechProgramServiceClient(conn)
	} else {
		log.Info("Fintech Program Service is not enabled.  Skipping setup.")
	}

	// Setup Gin
	connector.Gin = gin.Default()
	connector.Gin.POST("/fednow/outbound", connector.HandleOutboundPost)

	return &connector, nil

}

func (fc *FedNowConnector) Start() error {

	// Context
	ctx := context.Background()

	// Spin up Jetstream listeners
	go fc.startJetstreamListeners(ctx, fc.Signal)

	// Spin up Gin Server
	go fc.startGinService(ctx, fc.Signal)

	for {

		log.Infof("Waiting for bidirectional messages")
		msg := <-fc.Signal
		switch msg.Direction {
		case "INBOUND":
			log.Infof("Received inbound message.  Persisting Payment")
			paymentMessage, err := convertFednowMessageToPayment(msg.MessageContents)
			if err != nil {
				log.Errorf("Failed to convert inbound message to Payment: %v", err)
				continue
			}

			log.Infof("Determining Ledger ID for Payment")
			if fc.Config.IsFintechProgramAware {
				determineLedgerIdFromFintechProgram(ctx, fc.FintechProgramClient, paymentMessage)
			} else {
				log.Infof("Fintech Program Service is not enabled.  Using default ledger ID: %v", fc.Config.DefaultLedgerId)
				paymentMessage.LedgerID = fc.Config.DefaultLedgerId
			}

			_, err = fc.PaymentsDB.NewInsert().Model(paymentMessage).Exec(ctx)
			if err != nil {
				log.Errorf("Failed to insert payment: %v", err)
				continue
			}

			paymentStatusHistory := pmtmodel.PaymentStatusHistory{
				ID:            uuid.NewString(),
				PaymentID:     paymentMessage.ID,
				PaymentStatus: pmtmodel.PAYMENT_STATUS_INSTRUCTION_RECEIVED,
				StatusDetail:  "Payment Instruction Received via Fednow Queue",
				CreateDate:    time.Now(),
			}
			_, err = fc.PaymentsDB.NewInsert().Model(&paymentStatusHistory).Exec(ctx)
			if err != nil {
				log.Errorf("Failed to insert payment status history: %v", err)
				continue
			}

			log.Infof("Kicking off Fednow Inbound Payment Workflow")
			pmtJson, _ := json.Marshal(paymentMessage)
			fs := &goflow.FlowService{RedisURL: fc.Config.RedisUrl}
			fs.Execute("fednow-inbound-payment-workflow", &goflow.Request{
				Body: pmtJson,
				Header: map[string][]string{
					"header-key":  []string{"header-value"},
					"header2_key": []string{"header2-value"}},
				Query: map[string][]string{
					"query-key":     []string{"query-value"},
					"query_key-two": []string{"query-value-two"},
				},
			})

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

func (fc *FedNowConnector) startGinService(ctx context.Context, signal chan FednowMessage) {
	err := fc.Gin.Run(fmt.Sprintf("%s:%s", fc.Config.ListenHost, fc.Config.ListenPort))
	if err != nil {
		os.Exit(1)
	}
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

func convertFednowMessageToPayment(msg []byte) (*pmtmodel.Payment, error) {

	var fedNowMessage pacs_008_001_08.Document
	err := xml.Unmarshal(msg, &fedNowMessage)
	if err != nil {
		log.Errorf("Error unmarshalling FedNow Message: %v", err)
		return nil, err
	}

	sourceAmountFloat, err := strconv.ParseFloat(string(fedNowMessage.FIToFICstmrCdtTrf.CdtTrfTxInf[0].IntrBkSttlmAmt.Text), 64)
	if err != nil {
		log.Errorf("Error converting FedNow Message: %v", err)
		return nil, err
	}
	sourceAmount := int64(sourceAmountFloat * 100)

	settlementDate := time.Time(*fedNowMessage.FIToFICstmrCdtTrf.CdtTrfTxInf[0].IntrBkSttlmDt)

	pmt := pmtmodel.Payment{
		ID:                          uuid.NewString(),
		PaymentNetworkID:            pmtmodel.PAYMENT_NETWORK_US_FEDNOW,
		ServiceSpecificID:           uuid.NewString(),
		NetworkIdentifier:           string(fedNowMessage.FIToFICstmrCdtTrf.CdtTrfTxInf[0].PmtId.EndToEndId),
		CurrentPaymentStatus:        pmtmodel.PAYMENT_STATUS_INSTRUCTION_RECEIVED,
		SourceCurrency:              string(fedNowMessage.FIToFICstmrCdtTrf.CdtTrfTxInf[0].IntrBkSttlmAmt.Ccy),
		TargetCurrency:              string(fedNowMessage.FIToFICstmrCdtTrf.CdtTrfTxInf[0].IntrBkSttlmAmt.Ccy),
		SourceAmount:                sourceAmount,
		TargetAmount:                sourceAmount,
		PaymentMessage:              msg,
		UltimateOriginatorEntityID:  "",
		UltimateBeneficiaryEntityID: "",
		CreateDate:                  time.Now(),
		EffectiveDate:               settlementDate,
		ModifyDate:                  time.Now(),
		IsBatch:                     false,
		ParentPaymentID:             "",
	}
	return &pmt, nil

}

func determineLedgerIdFromFintechProgram(ctx context.Context, client ftpmodel.FintechProgramServiceClient, payment *pmtmodel.Payment) {

}
