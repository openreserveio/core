package activities

import (
	"net/url"

	"github.com/moov-io/watchman/pkg/search"
	"github.com/openreserveio/core/core-payments/generated/glmodel"
)

type SanctionsScreenActivity struct {
	CoreGLClient    glmodel.GeneralLedgerServiceClient
	SanctionsClient search.Client
	PostalURL       *url.URL
}

func NewSanctionsScreenActivity(coreGlClient glmodel.GeneralLedgerServiceClient, sanctionsClient search.Client, postalURL *url.URL) *SanctionsScreenActivity {
	return &SanctionsScreenActivity{
		CoreGLClient:    coreGlClient,
		SanctionsClient: sanctionsClient,
		PostalURL:       postalURL,
	}
}
