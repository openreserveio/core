package core_gl_test

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	core_gl "github.com/openreserveio/core/integration-tests/core-gl"
	"github.com/openreserveio/core/integration-tests/generated/glmodel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var _ = Describe("BalanceSheetReport", func() {

	var glClient glmodel.GeneralLedgerServiceClient
	var ledgerId string

	BeforeEach(func() {

		glGrpcConn, err := grpc.Dial("localhost:4081", grpc.WithTransportCredentials(insecure.NewCredentials()))
		Expect(err).To(BeNil())
		glClient = glmodel.NewGeneralLedgerServiceClient(glGrpcConn)

	})

	Describe("Complex Balance Sheet Report", func() {

		It("Creates a COA with a few layers", func() {

			layeredCoa := core_gl.ChartOfAccounts{
				Assets: []core_gl.FinancialAccount{
					{
						Name:     "Layered Asset Account One",
						Class:    "ASSET",
						Code:     "1001",
						Currency: "USD",
						Children: []core_gl.FinancialAccount{
							{
								Name:     "First Layered Asset Account Under One",
								Class:    "ASSET",
								Code:     "11001",
								Currency: "USD",
								Children: nil,
							},
							{
								Name:     "Second Layered Asset Account Under One",
								Class:    "ASSET",
								Code:     "11002",
								Currency: "USD",
								Children: nil,
							},
						},
					},
					{
						Name:     "Layered Asset Account Two",
						Class:    "ASSET",
						Code:     "1002",
						Currency: "USD",
						Children: []core_gl.FinancialAccount{
							{
								Name:     "First Layered Asset Account Under Two",
								Class:    "ASSET",
								Code:     "11003",
								Currency: "USD",
								Children: []core_gl.FinancialAccount{
									{
										Name:     "MultiLayered Asset Account One UNDER 11003",
										Class:    "ASSET",
										Code:     "11005",
										Currency: "USD",
										Children: nil,
									},
								},
							},
							{
								Name:     "Second Layered Asset Account Under Two",
								Class:    "ASSET",
								Code:     "11004",
								Currency: "USD",
								Children: nil,
							},
						},
					},
				},
				Liabilities: []core_gl.FinancialAccount{
					{
						Name:     "Layered Liability Account One",
						Class:    "LIABILITY",
						Code:     "2001",
						Currency: "USD",
						Children: nil,
					},
					{
						Name:     "Layered Liability Account Two",
						Class:    "LIABILITY",
						Code:     "2002",
						Currency: "USD",
						Children: nil,
					},
				},
				Equity: []core_gl.FinancialAccount{
					{
						Name:     "Layered Equity Account One",
						Class:    "EQUITY",
						Code:     "3001",
						Currency: "USD",
						Children: nil,
					},
					{
						Name:     "Layered Equity Account Two",
						Class:    "EQUITY",
						Code:     "3002",
						Currency: "USD",
						Children: nil,
					},
				},
				Income: []core_gl.FinancialAccount{
					{
						Name:     "Layered Income Account One",
						Class:    "INCOME",
						Code:     "4001",
						Currency: "USD",
						Children: nil,
					},
					{
						Name:     "Layered Income Account Two",
						Class:    "INCOME",
						Code:     "4002",
						Currency: "USD",
						Children: nil,
					},
				},
				Expense: []core_gl.FinancialAccount{
					{
						Name:     "Layered Expense Account One",
						Class:    "EXPENSE",
						Code:     "5001",
						Currency: "USD",
						Children: nil,
					},
					{
						Name:     "Layered Expense Account Two",
						Class:    "EXPENSE",
						Code:     "5002",
						Currency: "USD",
						Children: nil,
					},
				},
			}
			flatCoaJSON, _ := json.Marshal(&layeredCoa)
			request := glmodel.CreateChartOfAccountsRequest{
				ProposedChartJSON: flatCoaJSON,
				Title:             "Balance Sheet Report COA",
			}
			response, err := glClient.CreateChartOfAccounts(context.Background(), &request)
			Expect(err).To(BeNil())
			Expect(response.Status.Code).To(Equal(int64(http.StatusOK)))

			ledgerId = response.CreatedLedgerId

		})

		It("Posts a journal entry", func() {

			postRequest := glmodel.PostTransactionRequest{
				LedgerId:        ledgerId,
				TransactionType: glmodel.PostTransactionRequest_JOURNAL_ENTRY,
				JournalEntry: &glmodel.JournalEntry{
					Debits: []*glmodel.JournalEntryItem{
						&glmodel.JournalEntryItem{
							AccountCode:    "1001",
							Amount:         500,
							Note:           "This is a test",
							TaggedEntityId: nil,
						},
					},
					Credits: []*glmodel.JournalEntryItem{
						&glmodel.JournalEntryItem{
							AccountCode:    "2001",
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

		It("Runs a Balance Sheet Report", func() {

			res, err := glClient.GenerateReport(context.Background(), &glmodel.GenerateReportRequest{
				LedgerId:       ledgerId,
				BaseReportType: glmodel.GenerateReportRequest_BALANCE_SHEET,
				AsOfDate:       time.Now().String(),
			})

			Expect(err).To(BeNil())
			Expect(res.Status.Code).To(Equal(int64(200)))

			var balanceSheet core_gl.BalanceSheet
			err = json.Unmarshal(res.EncodedReport, &balanceSheet)
			Expect(err).To(BeNil())
			Expect(balanceSheet).To(Not(BeNil()))
			Expect(balanceSheet.LedgerID).To(Equal(ledgerId))
			Expect(balanceSheet.TotalAssets).To(Equal(int64(500)))
			Expect(balanceSheet.TotalLiabilities).To(Equal(int64(500)))

		})

	})

})
