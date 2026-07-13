-- ENTITY DB
DROP TABLE IF EXISTS entity;
CREATE TABLE entity (
    id VARCHAR NOT NULL PRIMARY KEY,
    entity_type VARCHAR NOT NULL, -- INDIVIDUAL, US SOLE Proprietorship, US LLC, US CORPORATION, US GOVT AGENCY, FOREIGN CORP, GENERIC_COMPANY
    entity_name_id VARCHAR NOT NULL,
    mailing_address_id VARCHAR NOT NULL,
    business_address_id VARCHAR,
    latest_verification_id VARCHAR,
    source_type VARCHAR NOT NULL, -- ie PAYMENT, MANUAL_ENTRY_TRANSACTION, 
    source_id VARCHAR NOT NULL, -- ie the paymentId, journal entry transaction ID, etc
    remapped BOOLEAN NOT NULL,
    remapped_to_entity_id VARCHAR,
    create_date TIMESTAMPTZ NOT NULL,
    update_date TIMESTAMPTZ NOT NULL
);
CREATE INDEX entity_entity_type_idx ON entity(entity_type);
CREATE INDEX entity_entity_name_id_idx ON entity(entity_name_id);
CREATE INDEX entity_mailing_address_id_idx ON entity(mailing_address_id);
CREATE INDEX entity_source_type_idx ON entity(source_type, source_id);

DROP TABLE IF EXISTS entity_name;
CREATE TABLE entity_name (
    id VARCHAR NOT NULL PRIMARY KEY,
    individual_given_name VARCHAR,
    individual_middle_name VARCHAR,
    individual_sur_name VARCHAR,
    us_sole_proprietorship_name VARCHAR,
    us_llc_name VARCHAR,
    us_corporation_name VARCHAR,
    us_government_agency_name VARCHAR,
    foreign_corporation_name VARCHAR,
    create_date TIMESTAMPTZ NOT NULL
);

-- house: venue name e.g. "Brooklyn Academy of Music", and building names e.g. "Empire State Building"
-- category: for category queries like "restaurants", etc.
-- near: phrases like "in", "near", etc. used after a category phrase to help with parsing queries like "restaurants in Brooklyn"
-- house_number: usually refers to the external (street-facing) building number. In some countries this may be a compount, hyphenated number which also includes an apartment number, or a block number (a la Japan), but libpostal will just call it the house_number for simplicity.
-- road: street name(s)
-- unit: an apartment, unit, office, lot, or other secondary unit designator
-- level: expressions indicating a floor number e.g. "3rd Floor", "Ground Floor", etc.
-- staircase: numbered/lettered staircase
-- entrance: numbered/lettered entrance
-- po_box: post office box: typically found in non-physical (mail-only) addresses
-- postcode: postal codes used for mail sorting
-- suburb: usually an unofficial neighborhood name like "Harlem", "South Bronx", or "Crown Heights"
-- city_district: these are usually boroughs or districts within a city that serve some official purpose e.g. "Brooklyn" or "Hackney" or "Bratislava IV"
-- city: any human settlement including cities, towns, villages, hamlets, localities, etc.
-- island: named islands e.g. "Maui"
-- state_district: usually a second-level administrative division or county.
-- state: a first-level administrative division. Scotland, Northern Ireland, Wales, and England in the UK are mapped to "state" as well (convention used in OSM, GeoPlanet, etc.)
-- country_region: informal subdivision of a country without any political status
-- country: sovereign nations and their dependent territories, anything with an ISO-3166 code.
-- world_region: currently only used for appending “West Indies” after the country name, a pattern frequently used in the English-speaking Caribbean e.g. “Jamaica, West Indies”
DROP TABLE IF EXISTS entity_address;
CREATE TABLE entity_address (
    id VARCHAR NOT NULL PRIMARY KEY,
    raw_address TEXT NOT NULL,
    parsed_house_venue VARCHAR,
    parsed_category VARCHAR,
    parsed_near VARCHAR,
    parsed_house_number VARCHAR,
    parsed_road VARCHAR,
    parsed_unit VARCHAR,
    parsed_level VARCHAR,
    parsed_staircase VARCHAR,
    parsed_entrance VARCHAR,
    parsed_po_box VARCHAR,
    parsed_postcode VARCHAR,
    parsed_suburb VARCHAR,
    parsed_city_district VARCHAR,
    parsed_city VARCHAR,
    parsed_island VARCHAR,
    parsed_state_district VARCHAR,
    parsed_state VARCHAR,
    parsed_country_region VARCHAR,
    parsed_country VARCHAR,
    parsed_world_region VARCHAR,
    create_date TIMESTAMPTZ NOT NULL,
    parsed_date TIMESTAMPTZ
);

DROP TABLE IF EXISTS entity_verification;
CREATE TABLE entity_verification (
    id VARCHAR NOT NULL PRIMARY KEY,
    entity_id VARCHAR NOT NULL,
    verification_status VARCHAR NOT NULL,
    verification_start_date TIMESTAMPTZ NOT NULL,
    verification_complete_date TIMESTAMPTZ,
    verification_message VARCHAR
);
CREATE INDEX idx_entity_verification_entity_id ON entity_verification(entity_id);