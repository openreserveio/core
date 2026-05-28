package model

import (
	"database/sql"
	"time"

	"github.com/uptrace/bun"
)

var (
	ACCOUNT_CLASS_ASSET     = "ASSET"
	ACCOUNT_CLASS_LIABILITY = "LIABILITY"
	ACCOUNT_CLASS_EQUITY    = "EQUITY"
	ACCOUNT_CLASS_INCOME    = "INCOME"
	ACCOUNT_CLASS_EXPENSE   = "EXPENSE"

	TX_ENTRY_TYPE_DEBIT  = "DEBIT"
	TX_ENTRY_TYPE_CREDIT = "CREDIT"
)

type Ledger struct {
	bun.BaseModel  `bun:"table:ledger"`
	ID             string         `json:"id" bun:"ledger_id,pk"`
	Name           string         `json:"name" bun:"name,notnull"`
	IsSubledger    bool           `json:"is_subledger" bun:"is_subledger,notnull"`
	ParentLedgerID sql.NullString `json:"parent_ledger_id" bun:"parent_ledger_id"`
}

type Account struct {
	bun.BaseModel   `bun:"table:account"`
	ID              string                 `json:"id" bun:"account_id,pk"`
	LedgerID        string                 `json:"ledger_id" bun:"ledger_id,notnull"`
	Class           string                 `json:"class" bun:"class,notnull"`
	Code            string                 `json:"code" bun:"code,notnull"`
	Name            string                 `json:"name" bun:"name,notnull"`
	Metadata        map[string]interface{} `json:"metadata" bun:"metadata"`
	ParentAccountID sql.NullString         `json:"parent_account_id" bun:"parent_account_id"`
	Currency        string                 `json:"currency" bun:"currency,notnull"`
}

type LedgerTransaction struct {
	bun.BaseModel   `bun:"table:ledger_transaction"`
	ID              string                   `json:"id" bun:"ledger_transaction_id,pk"`
	LedgerID        string                   `json:"ledger_id" bun:"ledger_id,notnull"`
	Debits          []LedgerTransactionEntry `json:"debits" bun:"-"`
	Credits         []LedgerTransactionEntry `json:"credits" bun:"-"`
	Metadata        map[string]interface{}   `json:"metadata" bun:"metadata,type:jsonb"`
	TransactionDate time.Time                `json:"transaction_date" bun:"transaction_date,notnull"`
}

type AccountBalance struct {
	bun.BaseModel            `bun:"table:account_balance"`
	AccountID                string    `json:"account_id" bun:"account_id"`
	BalanceAsOfTransactionID string    `json:"balance_as_of_transaction_id" bun:"balance_as_of_transaction_id"`
	Balance                  int64     `json:"balance" bun:"balance"`
	BalanceDate              time.Time `json:"balance_date" bun:"balance_date"`
}

type LedgerTransactionEntry struct {
	bun.BaseModel        `bun:"table:ledger_transaction_entry"`
	LedgerTransactionID  string                 `json:"ledger_transaction_id" bun:"ledger_transaction_id,notnull"`
	AccountID            string                 `json:"account_id" bun:"account_id,notnull"`
	TransactionEntryType string                 `json:"transaction_entry_type" bun:"transaction_entry_type,notnull"`
	Amount               int64                  `json:"amount" bun:"amount"`
	Currency             string                 `json:"currency" bun:"currency"`
	Metadata             map[string]interface{} `json:"metadata" bun:"metadata"`
}
