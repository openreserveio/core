package activities

import (
	"context"
	"fmt"

	"github.com/openreserveio/core/core-payments/generated/glmodel"
)

func (act *SanctionsScreenActivity) RetrieveEntity(ctx context.Context, entityId string) (glmodel.LedgerEntity, error) {

	entity := glmodel.LedgerEntity{}
	resp, err := act.CoreGLClient.GetEntity(ctx, &glmodel.GetEntityRequest{EntityId: entityId})
	if err != nil {
		return entity, err
	}
	if resp.Status.Code != 200 {
		return entity, fmt.Errorf("%s - %s", resp.Status.Code, resp.Status.StatusMessage)
	}

	entity = *resp.Entity
	return entity, nil

}
