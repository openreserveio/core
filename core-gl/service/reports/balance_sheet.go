package reports

import (
	"context"
	"time"

	"github.com/openreserveio/core/core-gl/generated/model"
)

func GenerateBalanceSheetReport(ctx context.Context, ledgerClient model.CoreLedgerServiceClient, ledgerId string, asOfDate time.Time) (any, error) {

	//ledgerClient.FindLedgerAccounts(ctx, &model.FindLedgerAccountsRequest{
	//	CriteriaType:     model.FindLedgerAccountsRequest_BY_METADATA,
	//	MetadataCriteria: map[string]string {
	//		glmodel.ACCOUNT_CLASS_ASSET
	//	},
	//	LedgerId:         "",
	//})

	return nil, nil
}
