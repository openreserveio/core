package pmtmodel

type AccountingConfig struct {
	LedgerID                              string `json:"ledger-id"`
	FednowSettlementAccountCode           string `json:"us-fednow-settlement-account-code"`
	FednowClearingAccountCode             string `json:"us-fednow-clearing-account-code"`
	FednowSettlementInProgressAccountCode string `json:"us-fednow-settlement-in-progress-account-code"`
	FednowSuspenseAccountCode             string `json:"us-fednow-suspense-account-code"`
}
