package core_gl_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openreserveio/core/integration-tests/generated/glmodel"
	"github.com/openreserveio/core/integration-tests/generated/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var _ = Describe("PostJournalEntry", func() {

	var ledgerId string
	var coreLedgerClient model.CoreLedgerServiceClient
	var glClient glmodel.GeneralLedgerServiceClient

	BeforeEach(func() {

		clGrpcConn, err := grpc.Dial("localhost:4080", grpc.WithTransportCredentials(insecure.NewCredentials()))
		Expect(err).To(BeNil())
		coreLedgerClient = model.NewCoreLedgerServiceClient(clGrpcConn)

		glGrpcConn, err := grpc.Dial("localhost:4081", grpc.WithTransportCredentials(insecure.NewCredentials()))
		Expect(err).To(BeNil())
		glClient = glmodel.NewGeneralLedgerServiceClient(glGrpcConn)

	})

	Describe("Create a ledger and some accounts", func() {

		It("Creates a ledger", func() {

			ledger, err := coreLedgerClient.CreateLedger(context.Background(), &model.CreateLedgerRequest{
				Name:        "test_ledger",
				IsSubledger: false,
			})
			Expect(err).To(BeNil())
			Expect(ledger.LedgerId).To(Not(BeNil()))
			ledgerId = ledger.LedgerId

		})

		It("Creates an asset account", func() {

			assetAccount, err := coreLedgerClient.CreateLedgerAccount(context.Background(), &model.CreateLedgerAccountRequest{
				Name:     "cash",
				Code:     "101",
				LedgerId: ledgerId,
				Class:    "ASSET",
				Currency: "USD",
			})
			Expect(err).To(BeNil())
			Expect(assetAccount.AccountId).To(Not(BeNil()))

		})

		It("Creates a liability account", func() {

			liabAccount, err := coreLedgerClient.CreateLedgerAccount(context.Background(), &model.CreateLedgerAccountRequest{
				Name:     "deposits",
				Code:     "201",
				LedgerId: ledgerId,
				Class:    "LIABILITY",
				Currency: "USD",
			})
			Expect(err).To(BeNil())
			Expect(liabAccount.AccountId).To(Not(BeNil()))

		})

	})

	Describe("Posting a journal entry to the GL", func() {

		It("Posts a journal entry", func() {

			postRequest := glmodel.PostTransactionRequest{
				LedgerId:        ledgerId,
				TransactionType: glmodel.PostTransactionRequest_JOURNAL_ENTRY,
				JournalEntry: &glmodel.JournalEntry{
					Debits: []*glmodel.JournalEntryItem{
						&glmodel.JournalEntryItem{
							AccountCode:    "101",
							Amount:         500,
							Note:           "This is a test",
							TaggedEntityId: nil,
						},
					},
					Credits: []*glmodel.JournalEntryItem{
						&glmodel.JournalEntryItem{
							AccountCode:    "201",
							Amount:         500,
							Note:           "This is a test",
							TaggedEntityId: nil,
						},
					},
					Purpose:        "Integration Test",
					PosterEntityId: "",
					TaggedEntityId: nil,
				},
			}

			res, err := glClient.PostTransaction(context.Background(), &postRequest)
			Expect(err).To(BeNil())
			Expect(res.Status.Code).To(Equal(int64(200)))
			Expect(res.TransactionId).To(Not(BeNil()))

		})

	})

})
