package core_gl_test

import (
	"context"
	"encoding/json"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	core_gl "github.com/openreserveio/core/integration-tests/core-gl"
	"github.com/openreserveio/core/integration-tests/generated/glmodel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var _ = Describe("ChartOfAccounts", func() {

	var glClient glmodel.GeneralLedgerServiceClient

	BeforeEach(func() {

		glGrpcConn, err := grpc.Dial("localhost:4081", grpc.WithTransportCredentials(insecure.NewCredentials()))
		Expect(err).To(BeNil())
		glClient = glmodel.NewGeneralLedgerServiceClient(glGrpcConn)

	})

	Describe("Creating a flat chart of accounts", func() {

		It("Creates Chart of Accounts with a single layer of accounting structure, including new ledger", func() {

			flatCoa := core_gl.ChartOfAccounts{
				Assets: []core_gl.FinancialAccount{
					{
						Name:     "Asset Account One",
						Class:    "ASSET",
						Code:     "101",
						Currency: "USD",
						Children: nil,
					},
					{
						Name:     "Asset Account Two",
						Class:    "ASSET",
						Code:     "102",
						Currency: "USD",
						Children: nil,
					},
				},
				Liabilities: []core_gl.FinancialAccount{
					{
						Name:     "Liability Account One",
						Class:    "LIABILITY",
						Code:     "201",
						Currency: "USD",
						Children: nil,
					},
					{
						Name:     "Liability Account Two",
						Class:    "LIABILITY",
						Code:     "202",
						Currency: "USD",
						Children: nil,
					},
				},
				Equity: []core_gl.FinancialAccount{
					{
						Name:     "Equity Account One",
						Class:    "EQUITY",
						Code:     "301",
						Currency: "USD",
						Children: nil,
					},
					{
						Name:     "Equity Account Two",
						Class:    "EQUITY",
						Code:     "302",
						Currency: "USD",
						Children: nil,
					},
				},
				Income: []core_gl.FinancialAccount{
					{
						Name:     "Income Account One",
						Class:    "INCOME",
						Code:     "401",
						Currency: "USD",
						Children: nil,
					},
					{
						Name:     "Income Account Two",
						Class:    "INCOME",
						Code:     "402",
						Currency: "USD",
						Children: nil,
					},
				},
				Expense: []core_gl.FinancialAccount{
					{
						Name:     "Expense Account One",
						Class:    "EXPENSE",
						Code:     "501",
						Currency: "USD",
						Children: nil,
					},
					{
						Name:     "Expense Account Two",
						Class:    "EXPENSE",
						Code:     "502",
						Currency: "USD",
						Children: nil,
					},
				},
			}
			flatCoaJSON, _ := json.Marshal(&flatCoa)
			request := glmodel.CreateChartOfAccountsRequest{
				ProposedChartJSON: flatCoaJSON,
				Title:             "Flat Chart of Accounts",
			}
			response, err := glClient.CreateChartOfAccounts(context.Background(), &request)
			Expect(err).To(BeNil())
			Expect(response.Status.Code).To(Equal(int64(http.StatusOK)))

		})

	})

	Describe("Creating a hierarchial chart of accounts", func() {

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
				Title:             "Layered Chart of Accounts",
			}
			response, err := glClient.CreateChartOfAccounts(context.Background(), &request)
			Expect(err).To(BeNil())
			Expect(response.Status.Code).To(Equal(int64(http.StatusOK)))

		})

	})

})
