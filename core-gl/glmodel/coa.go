package glmodel

var (
	ACCOUNT_CLASS_ASSET     = "ASSET"
	ACCOUNT_CLASS_LIABILITY = "LIABILITY"
	ACCOUNT_CLASS_EQUITY    = "EQUITY"
	ACCOUNT_CLASS_INCOME    = "INCOME"
	ACCOUNT_CLASS_EXPENSE   = "EXPENSE"

	TX_ENTRY_TYPE_DEBIT  = "DEBIT"
	TX_ENTRY_TYPE_CREDIT = "CREDIT"
)

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
	Metadata  map[string]string  `json:"metadata"`
}
