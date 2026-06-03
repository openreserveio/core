package activities

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/moov-io/watchman/pkg/search"
	"github.com/openreserveio/core/core-payments/generated/glmodel"
)

type SanctionScreenResult struct {
	Entity      glmodel.LedgerEntity
	Score       float64
	ScoreResult string
}

func (act *SanctionsScreenActivity) WatchmanScreen(ctx context.Context, entity glmodel.LedgerEntity) ([]SanctionScreenResult, error) {

	// Perform Watchman screen activity
	results := []SanctionScreenResult{}
	var query search.Entity[search.Value]
	if entity.EntityType == glmodel.LedgerEntity_INDIVIDUAL {
		query = search.Entity[search.Value]{
			Type: search.EntityPerson,
			Name: fmt.Sprintf("%s %s", entity.EntityName.IndividualGivenName, entity.EntityName.IndividualSurName),
			//Addresses: []search.Address{
			//	search.Address{
			//		Line1:      fmt.Sprintf("%s %s", entity.MailingAddress.HouseNumber, entity.MailingAddress.Road),
			//		Line2:      "",
			//		City:       entity.MailingAddress.City,
			//		PostalCode: entity.MailingAddress.Postcode,
			//		State:      entity.MailingAddress.State,
			//		Country:    entity.MailingAddress.Country,
			//	},
			//},
			//Person: &search.Person{
			//	Name:          fmt.Sprintf("%s %s", entity.EntityName.IndividualGivenName, entity.EntityName.IndividualSurName),
			//	AltNames:      nil,
			//	Gender:        "",
			//	BirthDate:     nil,
			//	PlaceOfBirth:  "",
			//	DeathDate:     nil,
			//	Titles:        nil,
			//	GovernmentIDs: nil,
			//},
		}
	}

	//   opts := SearchOpts{Limit: 10}
	resp, err := act.SanctionsClient.SearchByEntity(ctx, query, search.SearchOpts{MinMatch: 0.6})
	if err != nil {
		return results, err
	}

	for _, res := range resp.Entities {
		scoreResult, _ := json.Marshal(res)
		results = append(results, SanctionScreenResult{
			Entity:      entity,
			Score:       res.Match,
			ScoreResult: string(scoreResult),
		})
	}

	return results, nil

}
