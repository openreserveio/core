package activities

import (
	"context"
	"fmt"

	"github.com/openreserveio/core/core-payments/generated/glmodel"
)

func (act *SanctionsScreenActivity) UpdateEntity(ctx context.Context, entity glmodel.LedgerEntity) (glmodel.LedgerEntity, error) {

	resp, err := act.CoreGLClient.UpdateEntity(ctx, &glmodel.UpdateEntityRequest{Entity: &entity})
	if err != nil {
		return entity, err
	}
	if resp.Status.Code != 200 {
		return entity, fmt.Errorf("%d - %s", resp.Status.Code, resp.Status.StatusMessage)
	}

	return *resp.Entity, nil

}
