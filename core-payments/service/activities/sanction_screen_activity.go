package activities

import (
	"github.com/moov-io/watchman/pkg/search"
	"github.com/openreserveio/core/core-payments/generated/glmodel"
)

type SanctionsScreenActivity struct {
	CoreGLClient    glmodel.GeneralLedgerServiceClient
	SanctionsClient search.Client
}

func NewSanctionsScreenActivity(coreGlClient glmodel.GeneralLedgerServiceClient, sanctionsClient search.Client) *SanctionsScreenActivity {
	return &SanctionsScreenActivity{
		CoreGLClient:    coreGlClient,
		SanctionsClient: sanctionsClient,
	}
}
