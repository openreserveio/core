package core_ledger_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openreserveio/core/integration-tests/generated/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var _ = Describe("Accounts", func() {

	var client model.CoreLedgerServiceClient
	var ledgerId string
	var accountId string

	BeforeEach(func() {

		conn, err := grpc.NewClient("localhost:4080", grpc.WithTransportCredentials(insecure.NewCredentials()))
		Expect(err).To(BeNil())
		Expect(conn).NotTo(BeNil())
		client = model.NewCoreLedgerServiceClient(conn)

	})

	Describe("Accounts Management", func() {

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

		It("Creates a new account", func() {

			response, err := client.CreateLedgerAccount(context.Background(), &model.CreateLedgerAccountRequest{
				Name:     "test_account",
				Code:     "101",
				LedgerId: ledgerId,
				Class:    "ASSET",
				Currency: "USD",
			})
			Expect(err).To(BeNil())
			Expect(response.Status.Code).To(Equal(int64(200)))
			Expect(response.Name).To(Equal("test_account"))
			Expect(response.AccountId).To(Not(BeNil()))

			accountId = response.AccountId

		})

		It("Gets the account just created by ID", func() {

			response, err := client.GetLedgerAccount(context.Background(), &model.GetLedgerAccountRequest{
				AccountId: accountId,
				LedgerId:  ledgerId,
			})
			Expect(err).To(BeNil())
			Expect(response.Status.Code).To(Equal(int64(200)))
			Expect(response.AccountId).To(Not(BeNil()))
			Expect(response.Name).To(Equal("test_account"))
			Expect(response.Code).To(Equal("101"))
			Expect(response.Class).To(Equal("ASSET"))

		})

		It("Gets the account just created by Code", func() {

			response, err := client.GetLedgerAccount(context.Background(), &model.GetLedgerAccountRequest{
				Code:     "101",
				LedgerId: ledgerId,
			})
			Expect(err).To(BeNil())
			Expect(response.Status.Code).To(Equal(int64(200)))
			Expect(response.AccountId).To(Not(BeNil()))
			Expect(response.Name).To(Equal("test_account"))
			Expect(response.Code).To(Equal("101"))
			Expect(response.Class).To(Equal("ASSET"))

		})

		It("Creates a new account with metadata", func() {

			response, err := client.CreateLedgerAccount(context.Background(), &model.CreateLedgerAccountRequest{
				Name:     "test_account_metadata",
				Code:     "103",
				LedgerId: ledgerId,
				Class:    "ASSET",
				Currency: "USD",
				Metadata: map[string]string{
					"name": "test_account_metadata",
					"code": "102",
				},
			})
			Expect(err).To(BeNil())
			Expect(response.Status.Code).To(Equal(int64(200)))
			Expect(response.Name).To(Equal("test_account_metadata"))
			Expect(response.AccountId).To(Not(BeNil()))

			// get the account and check the metadata
			getAccount, err := client.GetLedgerAccount(context.Background(), &model.GetLedgerAccountRequest{
				AccountId: response.AccountId,
				LedgerId:  ledgerId,
			})
			Expect(err).To(BeNil())

			metadata := getAccount.Metadata

			Expect(metadata).To(HaveKeyWithValue("name", "test_account_metadata"))
			Expect(metadata).To(HaveKeyWithValue("code", "102"))

		})

		It("Finds the account by metadata", func() {

			resp, err := client.FindLedgerAccounts(context.Background(), &model.FindLedgerAccountsRequest{
				CriteriaType: model.FindLedgerAccountsRequest_BY_METADATA,
				MetadataCriteria: map[string]string{
					"name": "test_account_metadata",
				},
				LedgerId: ledgerId,
			})
			Expect(err).To(BeNil())
			Expect(resp.Status.Code).To(Equal(int64(200)))
			Expect(resp.Accounts).To(HaveLen(1))
			Expect(resp.Accounts[0].Name).To(Equal("test_account_metadata"))
			Expect(resp.Accounts[0].AccountId).To(Not(BeNil()))

			respCode, err := client.FindLedgerAccounts(context.Background(), &model.FindLedgerAccountsRequest{
				CriteriaType: model.FindLedgerAccountsRequest_BY_METADATA,
				MetadataCriteria: map[string]string{
					"code": "102",
				},
				LedgerId: ledgerId,
			})
			Expect(err).To(BeNil())
			Expect(respCode.Status.Code).To(Equal(int64(200)))
			Expect(respCode.Accounts).To(HaveLen(1))
			Expect(respCode.Accounts[0].Name).To(Equal("test_account_metadata"))
			Expect(respCode.Accounts[0].AccountId).To(Not(BeNil()))

		})

		It("Finds the accounts by class", func() {

			respClass, err := client.FindLedgerAccounts(context.Background(), &model.FindLedgerAccountsRequest{
				CriteriaType: model.FindLedgerAccountsRequest_BY_ACCOUNT_CLASS,
				AccountClass: "ASSET",
				LedgerId:     ledgerId,
			})
			Expect(err).To(BeNil())
			Expect(respClass.Status.Code).To(Equal(int64(200)))
			Expect(respClass.Accounts).To(HaveLen(2))

		})

	})

})
