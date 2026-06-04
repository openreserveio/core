package activities

import (
	"github.com/openreserveio/core/core-payments/generated/glmodel"
	"github.com/openreserveio/core/core-payments/pmtmodel"
)

type SuspendPaymentActivity struct {
	GLServiceClient  glmodel.GeneralLedgerServiceClient
	AccountingConfig *pmtmodel.AccountingConfig
}

func NewSuspendPaymentActivity(coreGlClient glmodel.GeneralLedgerServiceClient, accountingConfig *pmtmodel.AccountingConfig) *SuspendPaymentActivity {
	return &SuspendPaymentActivity{
		GLServiceClient:  coreGlClient,
		AccountingConfig: accountingConfig,
	}
}
