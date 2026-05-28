package core_gl

type ChartOfAccounts struct {
	LedgerID    string             `json:"ledger_id"`
	Assets      []FinancialAccount `json:"assets"`
	Liabilities []FinancialAccount `json:"liabilities"`
	Equity      []FinancialAccount `json:"equity"`
	Income      []FinancialAccount `json:"income"`
	Expense     []FinancialAccount `json:"expense"`
}

type FinancialAccount struct {
	AccountID string             `json:"account_id"`
	Name      string             `json:"name"`
	Class     string             `json:"class"`
	Code      string             `json:"code"`
	Currency  string             `json:"currency"`
	Children  []FinancialAccount `json:"children"`
}
