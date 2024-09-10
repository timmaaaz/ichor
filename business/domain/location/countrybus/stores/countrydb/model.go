package countrydb

import (
	"bitbucket.org/superiortechnologies/ichor/business/domain/location/countrybus"
	"github.com/google/uuid"
)

type country struct {
	ID     uuid.UUID `db:"country_id"`
	Number int       `db:"number"`
	Name   string    `db:"name"`
	Alpha2 string    `db:"alpha_2"`
	Alpha3 string    `db:"alpha_3"`
}

func toBusCountry(dbCtry country) countrybus.Country {
	return countrybus.Country{
		ID:     dbCtry.ID,
		Number: dbCtry.Number,
		Name:   dbCtry.Name,
		Alpha2: dbCtry.Alpha2,
		Alpha3: dbCtry.Alpha3,
	}
}

func toBusCountries(dbCountries []country) []countrybus.Country {
	countries := make([]countrybus.Country, len(dbCountries))
	for i, ctry := range dbCountries {
		countries[i] = toBusCountry(ctry)
	}
	return countries
}
