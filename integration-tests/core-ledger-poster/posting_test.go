package core_ledger_poster_test

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openreserveio/core/integration-tests/generated/model"
	"google.golang.org/protobuf/proto"
)

var _ = Describe("Posting", func() {

	var jsClient jetstream.JetStream

	BeforeEach(func() {

		nc, _ := nats.Connect("localhost:4222")
		jsClient, _ = jetstream.New(nc)

	})

	Describe("Posting a transaction to the ledger", func() {

		debit := model.PostLedgerTransactionRequest_Entry{
			AccountId: "a7911733-c9dd-4ce7-a2b2-dd047f257637",
			Amount:    50,
			Currency:  "USD",
		}

		credit := model.PostLedgerTransactionRequest_Entry{
			AccountId: "6ffffa5e-e676-4558-8aaa-a847c78a2ccc",
			Amount:    50,
			Currency:  "USD",
		}

		request := model.PostLedgerTransactionRequest{
			Debits:  []*model.PostLedgerTransactionRequest_Entry{&debit},
			Credits: []*model.PostLedgerTransactionRequest_Entry{&credit},
		}

		It("Posts a transaction to queue and receives a reply", func() {

			// marshall to protobuf and post to queue
			replySubject := uuid.NewString()
			requestBytes, _ := proto.Marshal(&request)
			msg := nats.Msg{
				Subject: "CORELEDGER.posts",
				Reply:   "CORELEDGER.replies." + replySubject,
				Data:    requestBytes,
			}
			jsClient.PublishMsg(context.Background(), &msg)

			// wait for reply
			stream, err := jsClient.CreateOrUpdateStream(context.Background(), jetstream.StreamConfig{
				Name:     "CORELEDGER",
				Subjects: []string{"CORELEDGER.replies." + replySubject},
			})
			Expect(err).To(BeNil())

			consumer, _ := stream.CreateConsumer(context.Background(), jetstream.ConsumerConfig{AckPolicy: jetstream.AckNonePolicy})
			replyMsg, err := consumer.Next(jetstream.FetchMaxWait(30 * time.Second))

			Expect(err).To(BeNil())
			Expect(replyMsg.Data()).To(Not(BeNil()))

			var response model.PostLedgerTransactionResponse
			err = proto.Unmarshal(replyMsg.Data(), &response)
			Expect(err).To(BeNil())
			Expect(response.Status.Code).To(Equal(int64(http.StatusOK)))

		})

	})

})
