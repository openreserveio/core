package core_ledger_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openreserveio/core/integration-tests/generated/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var _ = Describe("Transactions", func() {

	var client model.CoreLedgerServiceClient
	var ledgerId string
	var assetAccountId string
	var secondAssetAccountId string
	var liabilityAccountId string
	var transactionId string

	BeforeEach(func() {

		conn, err := grpc.NewClient("localhost:4080", grpc.WithTransportCredentials(insecure.NewCredentials()))
		Expect(err).To(BeNil())
		Expect(conn).NotTo(BeNil())
		client = model.NewCoreLedgerServiceClient(conn)

	})

	Describe("Posting transactions to ledger", func() {

		It("Creates ledger, accounts", func() {

			// Ledger
			response, err := client.CreateLedger(context.Background(), &model.CreateLedgerRequest{
				Name:        "test_ledger_for_posting",
				IsSubledger: false,
			})
			Expect(err).To(BeNil())
			Expect(response.Name).To(Equal("test_ledger_for_posting"))
			Expect(response.LedgerId).To(Not(BeNil()))
			ledgerId = response.LedgerId

			// Asset Account
			responseAsset, err := client.CreateLedgerAccount(context.Background(), &model.CreateLedgerAccountRequest{
				Name:     "cash",
				Code:     "101",
				LedgerId: ledgerId,
				Class:    "ASSET",
				Currency: "USD",
			})
			Expect(err).To(BeNil())
			Expect(responseAsset.Status.Code).To(Equal(int64(200)))
			Expect(responseAsset.Name).To(Equal("cash"))
			Expect(responseAsset.AccountId).To(Not(BeNil()))

			assetAccountId = responseAsset.AccountId
			_ = assetAccountId

			// Liability Account
			responseLiability, err := client.CreateLedgerAccount(context.Background(), &model.CreateLedgerAccountRequest{
				Name:     "deposits",
				Code:     "201",
				LedgerId: ledgerId,
				Class:    "LIABILITY",
				Currency: "USD",
			})
			Expect(err).To(BeNil())
			Expect(responseLiability.Status.Code).To(Equal(int64(200)))
			Expect(responseLiability.Name).To(Equal("deposits"))
			Expect(responseLiability.AccountId).To(Not(BeNil()))

			liabilityAccountId = responseLiability.AccountId
			_ = liabilityAccountId

		})

		It("Posts a balanced transaction between the asset (cash) and liability (deposits)", func() {

			debitEntry := model.PostLedgerTransactionRequest_Entry{
				AccountId: assetAccountId,
				Amount:    1000,
				Currency:  "USD",
			}
			creditEntry := model.PostLedgerTransactionRequest_Entry{
				AccountId: liabilityAccountId,
				Amount:    1000,
				Currency:  "USD",
			}

			tx := model.PostLedgerTransactionRequest{
				LedgerId: ledgerId,
				Debits:   []*model.PostLedgerTransactionRequest_Entry{&debitEntry},
				Credits:  []*model.PostLedgerTransactionRequest_Entry{&creditEntry},
			}

			responseTxPost, err := client.PostLedgerTransaction(context.Background(), &tx)
			Expect(err).To(BeNil())
			Expect(responseTxPost.Status.Code).To(Equal(int64(200)))

		})

		It("Is able to get the latest balances for the accounts", func() {

			assetResponse, err := client.GetLedgerAccountBalance(context.Background(), &model.GetLedgerAccountBalanceRequest{
				AccountId: assetAccountId,
				LedgerId:  ledgerId,
			})
			Expect(err).To(BeNil())
			Expect(assetResponse.Status.Code).To(Equal(int64(200)))
			Expect(assetResponse.Balance).To(Equal(int64(1000)))

			liabilityResponse, err := client.GetLedgerAccountBalance(context.Background(), &model.GetLedgerAccountBalanceRequest{
				AccountId: liabilityAccountId,
				LedgerId:  ledgerId,
			})
			Expect(err).To(BeNil())
			Expect(liabilityResponse.Status.Code).To(Equal(int64(200)))
			Expect(liabilityResponse.Balance).To(Equal(int64(1000)))

		})

		It("Creates another asset account", func() {

			// Asset Account
			responseAsset, err := client.CreateLedgerAccount(context.Background(), &model.CreateLedgerAccountRequest{
				Name:     "savings",
				Code:     "102",
				LedgerId: ledgerId,
				Class:    "ASSET",
				Currency: "USD",
			})
			Expect(err).To(BeNil())
			Expect(responseAsset.Status.Code).To(Equal(int64(200)))
			Expect(responseAsset.Name).To(Equal("savings"))
			Expect(responseAsset.AccountId).To(Not(BeNil()))

			secondAssetAccountId = responseAsset.AccountId

		})

		It("Creates another Transaction to update the balances, essentially a transfer between assets", func() {

			debitEntry := model.PostLedgerTransactionRequest_Entry{
				AccountId: secondAssetAccountId,
				Amount:    750,
				Currency:  "USD",
			}
			creditEntry := model.PostLedgerTransactionRequest_Entry{
				AccountId: assetAccountId,
				Amount:    750,
				Currency:  "USD",
			}

			tx := model.PostLedgerTransactionRequest{
				LedgerId: ledgerId,
				Debits:   []*model.PostLedgerTransactionRequest_Entry{&debitEntry},
				Credits:  []*model.PostLedgerTransactionRequest_Entry{&creditEntry},
			}

			responseTxPost, err := client.PostLedgerTransaction(context.Background(), &tx)
			Expect(err).To(BeNil())
			Expect(responseTxPost.Status.Code).To(Equal(int64(200)))

			transactionId = responseTxPost.LedgerTransactionId

		})

		It("Is able to get the latest balances for the accounts reflecting the updates", func() {

			assetResponse, err := client.GetLedgerAccountBalance(context.Background(), &model.GetLedgerAccountBalanceRequest{
				AccountId: assetAccountId,
			})
			Expect(err).To(BeNil())
			Expect(assetResponse.Status.Code).To(Equal(int64(200)))
			Expect(assetResponse.Balance).To(Equal(int64(250)))

			secondAssetResponse, err := client.GetLedgerAccountBalance(context.Background(), &model.GetLedgerAccountBalanceRequest{
				AccountId: secondAssetAccountId,
				LedgerId:  ledgerId,
			})
			Expect(err).To(BeNil())
			Expect(secondAssetResponse.Status.Code).To(Equal(int64(200)))
			Expect(secondAssetResponse.Balance).To(Equal(int64(750)))

		})

		It("Is able to get the transaction it posted", func() {

			ledgerTxResponse, err := client.GetLedgerTransaction(context.Background(), &model.GetLedgerTransactionRequest{
				LedgerId:            ledgerId,
				LedgerTransactionId: transactionId,
			})
			Expect(err).To(BeNil())
			Expect(ledgerTxResponse.Status.Code).To(Equal(int64(200)))
			Expect(ledgerTxResponse.LedgerId).To(Equal(ledgerId))

		})

	})

})
