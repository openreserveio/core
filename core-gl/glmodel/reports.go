package glmodel

type BalanceSheet struct {
	LedgerID string `json:"ledger_id"`

	Assets      []FinancialAccountBalance `json:"assets"`
	TotalAssets int64                     `json:"total_assets"`

	Liabilities      []FinancialAccountBalance `json:"liabilities"`
	TotalLiabilities int64                     `json:"total_liabilities"`

	Equity      []FinancialAccountBalance `json:"equity"`
	TotalEquity int64                     `json:"total_equity"`

	Income      []FinancialAccountBalance `json:"income"`
	TotalIncome int64                     `json:"total_income"`

	Expense      []FinancialAccountBalance `json:"expense"`
	TotalExpense int64                     `json:"total_expense"`
}

type FinancialAccountBalance struct {
	FinancialAccount
	Balance  int64  `json:"balance"`
	AsOfDate string `json:"as_of_date"`
}
