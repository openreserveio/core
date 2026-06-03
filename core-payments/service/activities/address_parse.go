package activities

import (
	"context"
	"errors"
	"time"

	"github.com/openreserveio/core/core-payments/generated/glmodel"
	"resty.dev/v3"
)

type AddressParseRequest struct {
	Query string `json:"query"`
}

func (act *SanctionsScreenActivity) AddressParse(ctx context.Context, address glmodel.LedgerEntityAddress) (glmodel.LedgerEntityAddress, error) {

	if address.RawAddress == "" {
		return address, nil
	}

	client := resty.New()
	defer client.Close()

	parseRequest := AddressParseRequest{Query: address.RawAddress}
	res, err := client.R().
		SetContentType("application/json").
		SetBody(parseRequest).
		SetResult(ParsedAddressComponents{}).
		Post(act.PostalURL.String())

	if err != nil {
		return address, err
	}
	if res.IsError() {
		return address, errors.New(res.String())
	}

	parsedAddress := res.Result().(*ParsedAddressComponents)
	parsedLedgerEntryAddress := MapToLedgerEntityAddress(*parsedAddress)
	parsedLedgerEntryAddress.EntityAddressId = address.EntityAddressId
	parsedLedgerEntryAddress.RawAddress = address.RawAddress
	parsedLedgerEntryAddress.CreateDate = address.CreateDate
	parsedLedgerEntryAddress.ParsedDate = time.Now().String()

	return *parsedLedgerEntryAddress, nil

}

// house: venue name e.g. "Brooklyn Academy of Music", and building names e.g. "Empire State Building"
// category: for category queries like "restaurants", etc.
// near: phrases like "in", "near", etc. used after a category phrase to help with parsing queries like "restaurants in Brooklyn"
// house_number: usually refers to the external (street-facing) building number. In some countries this may be a compount, hyphenated number which also includes an apartment number, or a block number (a la Japan), but libpostal will just call it the house_number for simplicity.
// road: street name(s)
// unit: an apartment, unit, office, lot, or other secondary unit designator
// level: expressions indicating a floor number e.g. "3rd Floor", "Ground Floor", etc.
// staircase: numbered/lettered staircase
// entrance: numbered/lettered entrance
// po_box: post office box: typically found in non-physical (mail-only) addresses
// postcode: postal codes used for mail sorting
// suburb: usually an unofficial neighborhood name like "Harlem", "South Bronx", or "Crown Heights"
// city_district: these are usually boroughs or districts within a city that serve some official purpose e.g. "Brooklyn" or "Hackney" or "Bratislava IV"
// city: any human settlement including cities, towns, villages, hamlets, localities, etc.
// island: named islands e.g. "Maui"
// state_district: usually a second-level administrative division or county.
// state: a first-level administrative division. Scotland, Northern Ireland, Wales, and England in the UK are mapped to "state" as well (convention used in OSM, GeoPlanet, etc.)
// country_region: informal subdivision of a country without any political status
// country: sovereign nations and their dependent territories, anything with an ISO-3166 code.
// world_region: currently only used for appending “West Indies” after the country name, a pattern frequently used in the English-speaking Caribbean e.g. “Jamaica, West Indies”
type ParsedAddress struct {
	House         string `bun:"parsed_house_venue" json:"house" xml:"parsed_house"`
	Category      string `bun:"parsed_category" json:"category" xml:"parsed_category"`
	Near          string `bun:"parsed_near" json:"near" xml:"parsed_near"`
	HouseNumber   string `bun:"parsed_house_number" json:"house_number" xml:"parsed_house_number"`
	Road          string `bun:"parsed_road" json:"road" xml:"parsed_road"`
	Unit          string `bun:"parsed_unit" json:"unit" xml:"parsed_unit"`
	Level         string `bun:"parsed_level" json:"level" xml:"parsed_level"`
	Staircase     string `bun:"parsed_staircase" json:"staircase" xml:"parsed_staircase"`
	Entrance      string `bun:"parsed_entrance" json:"entrance" xml:"parsed_entrance"`
	POBox         string `bun:"parsed_po_box" json:"po_box" xml:"parsed_po_box"`
	Postcode      string `bun:"parsed_postcode" json:"postcode" xml:"parsed_postcode"`
	Suburb        string `bun:"parsed_suburb" json:"suburb" xml:"parsed_suburb"`
	CityDistrict  string `bun:"parsed_city_district" json:"city_district" xml:"parsed_city_district"`
	City          string `bun:"parsed_city" json:"city" xml:"parsed_city"`
	Island        string `bun:"parsed_island" json:"island" xml:"parsed_island"`
	StateDistrict string `bun:"parsed_state_district" json:"state_district" xml:"parsed_state_district"`
	State         string `bun:"parsed_state" json:"state" xml:"parsed_state"`
	CountryRegion string `bun:"parsed_country_region" json:"pcountry_region" xml:"parsed_country_region"`
	Country       string `bun:"parsed_country" json:"country" xml:"parsed_country"`
	WorldRegion   string `bun:"parsed_world_region" json:"world_region" xml:"parsed_world_region"`
}

type ParsedAddressComponent struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type ParsedAddressComponents []ParsedAddressComponent

func MapToLedgerEntityAddress(components ParsedAddressComponents) *glmodel.LedgerEntityAddress {
	addr := &glmodel.LedgerEntityAddress{}
	for _, c := range components {
		switch c.Label {
		case "house":
			addr.House = c.Value
		case "category":
			addr.Category = c.Value
		case "near":
			addr.Near = c.Value
		case "house_number":
			addr.HouseNumber = c.Value
		case "road":
			addr.Road = c.Value
		case "unit":
			addr.Unit = c.Value
		case "level":
			addr.Level = c.Value
		case "staircase":
			addr.Staircase = c.Value
		case "entrance":
			addr.Entrance = c.Value
		case "po_box":
			addr.PoBox = c.Value
		case "postcode":
			addr.Postcode = c.Value
		case "suburb":
			addr.Suburb = c.Value
		case "city_district":
			addr.CityDistrict = c.Value
		case "city":
			addr.City = c.Value
		case "island":
			addr.Island = c.Value
		case "state_district":
			addr.StateDistrict = c.Value
		case "state":
			addr.State = c.Value
		case "country_region":
			addr.CountryRegion = c.Value
		case "country":
			addr.Country = c.Value
		case "world_region":
			addr.WorldRegion = c.Value
		}
	}
	return addr
}
