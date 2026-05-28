package glmodel

import (
	"github.com/uptrace/bun"
	"time"
)

type Entity struct {
	bun.BaseModel        `bun:"table:entity"`
	ID                   string    `bun:"id,pk" json:"id" xml:"id"`
	EntityType           string    `bun:"entity_type" json:"entity_type" xml:"entity_type"`
	EntityNameID         string    `bun:"entity_name_id" json:"entity_name_id" xml:"entity_name_id"`
	MailingAddressID     string    `bun:"mailing_address_id" json:"mailing_address_id" xml:"mailing_address_id"`
	BusinessAddressID    string    `bun:"business_address_id" json:"business_address_id" xml:"business_address_id"`
	LatestVerificationID string    `bun:"latest_verification_id" json:"latest_verification_id" xml:"latest_verification_id"`
	SourceType           string    `bun:"source_type" json:"source_type" xml:"source_type"`
	SourceID             string    `bun:"source_id" json:"source_id" xml:"source_id"`
	Remapped             bool      `bun:"remapped" json:"remapped" xml:"remapped"`
	RemappedToEntityID   string    `json:"remapped_to_entity_id" xml:"remapped_to_entity_id"`
	CreateDate           time.Time `bun:"create_date" json:"create_date" xml:"create_date"`
	UpdateDate           time.Time `bun:"update_date" json:"update_date" xml:"update_date"`
}

type EntityName struct {
	bun.BaseModel            `bun:"table:entity_name"`
	ID                       string    `bun:"id,pk" json:"id" xml:"id"`
	IndividualGivenName      string    `bun:"individual_given_name" json:"individual_given_name" xml:"individual_given_name"`
	IndividualSurName        string    `bun:"individual_sur_name" json:"individual_sur_name" xml:"individual_sur_name"`
	IndividualMiddleName     string    `bun:"individual_middle_name" json:"individual_middle_name" xml:"individual_middle_name"`
	USSoleProprietorshipName string    `bun:"us_sole_proprietorship_name" json:"us_sole_proprietorship_name" xml:"us_sole_proprietorship_name"`
	USLLCName                string    `bun:"us_llc_name" json:"us_llc_name" xml:"us_llc_name"`
	USCorporationName        string    `bun:"us_corporation_name" json:"us_corporation_name" xml:"us_corporation_name"`
	USGovernmentAgencyName   string    `bun:"us_government_agency_name" json:"us_government_agency_name" xml:"us_government_agency_name"`
	ForeignCorporationName   string    `bun:"foreign_corporation_name" json:"foreign_corporation_name" xml:"foreign_corporation_name"`
	CreateDate               time.Time `bun:"create_date" json:"create_date" xml:"create_date"`
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
type EntityAddress struct {
	bun.BaseModel `bun:"table:entity_address"`
	ID            string `bun:"id,pk" json:"id" xml:"id"`
	RawAddress    string `bun:"raw_address" json:"raw_address" xml:"raw_address"`

	House         string `bun:"parsed_house_venue" json:"parsed_house_venue" xml:"parsed_house"`
	Category      string `bun:"parsed_category" json:"parsed_category" xml:"parsed_category"`
	Near          string `bun:"parsed_near" json:"parsed_near" xml:"parsed_near"`
	HouseNumber   string `bun:"parsed_house_number" json:"parsed_house_number" xml:"parsed_house_number"`
	Road          string `bun:"parsed_road" json:"parsed_road" xml:"parsed_road"`
	Unit          string `bun:"parsed_unit" json:"parsed_unit" xml:"parsed_unit"`
	Level         string `bun:"parsed_level" json:"parsed_level" xml:"parsed_level"`
	Staircase     string `bun:"parsed_staircase" json:"parsed_staircase" xml:"parsed_staircase"`
	Entrance      string `bun:"parsed_entrance" json:"parsed_entrance" xml:"parsed_entrance"`
	POBox         string `bun:"parsed_po_box" json:"parsed_po_box" xml:"parsed_po_box"`
	Postcode      string `bun:"parsed_postcode" json:"parsed_postcode" xml:"parsed_postcode"`
	Suburb        string `bun:"parsed_suburb" json:"parsed_suburb" xml:"parsed_suburb"`
	CityDistrict  string `bun:"parsed_city_district" json:"parsed_city_district" xml:"parsed_city_district"`
	City          string `bun:"parsed_city" json:"parsed_city" xml:"parsed_city"`
	Island        string `bun:"parsed_island" json:"parsed_island" xml:"parsed_island"`
	StateDistrict string `bun:"parsed_state_district" json:"parsed_state_district" xml:"parsed_state_district"`
	State         string `bun:"parsed_state" json:"parsed_state" xml:"parsed_state"`
	CountryRegion string `bun:"parsed_country_region" json:"parsed_country_region" xml:"parsed_country_region"`
	Country       string `bun:"parsed_country" json:"parsed_country" xml:"parsed_country"`
	WorldRegion   string `bun:"parsed_world_region" json:"parsed_world_region" xml:"parsed_world_region"`

	CreateDate time.Time `bun:"create_date" json:"create_date" xml:"create_date"`
	ParsedDate time.Time `bun:"parsed_date" json:"parsed_date" xml:"parsed_date"`
}

const VERIFICATION_STATUS_PENDING = "PENDING"
const VERIFICATION_STATUS_VERIFIED = "VERIFIED"
const VERIFICATION_STATUS_FAILED = "FAILED"

type EntityVerification struct {
	bun.BaseModel            `bun:"table:entity_verification"`
	ID                       string    `bun:"id,pk" json:"id" xml:"id"`
	EntityID                 string    `bun:"entity_id" json:"entity_id" xml:"entity_id"`
	VerificationStatus       string    `bun:"verification_status" json:"verification_status" xml:"verification_status"`
	VerificationStartDate    time.Time `bun:"verification_start_date" json:"verification_start_date" xml:"verification_start_date"`
	VerificationCompleteDate time.Time `bun:"verification_complete_date" json:"verification_complete_date" xml:"verification_complete_date"`
	VerificationMessage      string    `bun:"verification_message" json:"verification_message" xml:"verification_message"`
}
