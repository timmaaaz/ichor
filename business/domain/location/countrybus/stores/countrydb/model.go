package countrydb

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/location/countrybus"
)

type country struct {
	ID     uuid.UUID `db:"id"`
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
