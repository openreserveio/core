package core_ledger_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openreserveio/core/integration-tests/generated/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var _ = Describe("Ledgers", func() {

	var client model.CoreLedgerServiceClient
	var ledgerId string

	BeforeEach(func() {
		conn, err := grpc.NewClient("localhost:4080", grpc.WithTransportCredentials(insecure.NewCredentials()))
		Expect(err).To(BeNil())
		Expect(conn).NotTo(BeNil())

		client = model.NewCoreLedgerServiceClient(conn)
	})

	Describe("Ledger Management", func() {

		It("Creates a new ledger", func() {

			response, err := client.CreateLedger(context.Background(), &model.CreateLedgerRequest{
				Name:        "test_ledger",
				IsSubledger: false,
			})
			Expect(err).To(BeNil())
			Expect(response.Name).To(Equal("test_ledger"))
			Expect(response.LedgerId).To(Not(BeNil()))
			ledgerId = response.LedgerId

		})

		It("Gets a ledger just created by ID", func() {

			response, err := client.GetLedger(context.Background(), &model.GetLedgerRequest{
				LedgerId: ledgerId,
			})
			Expect(err).To(BeNil())
			Expect(response.LedgerId).To(Equal(ledgerId))
			Expect(response.Name).To(Equal("test_ledger"))

		})

	})

})
