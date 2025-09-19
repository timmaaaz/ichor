package countryapp

import (
	"encoding/json"

	"github.com/timmaaaz/ichor/business/domain/geography/countrybus"
)

const dateFormat = "2006-01-02"

type QueryParams struct {
	Page    string
	Rows    string
	OrderBy string
	ID      string
	Name    string
	Alpha2  string
	Alpha3  string
}

// Country represents information about an individual country.
type Country struct {
	ID     string `json:"id"`
	Number int    `json:"number"`
	Name   string `json:"name"`
	Alpha2 string `json:"alpha_2"`
	Alpha3 string `json:"alpha_3"`
}

// Encode implements the encoder interface.
func (app Country) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func toAppCountry(ctry countrybus.Country) Country {
	return Country{
		ID:     ctry.ID.String(),
		Number: ctry.Number,
		Name:   ctry.Name,
		Alpha2: ctry.Alpha2,
		Alpha3: ctry.Alpha3,
	}
}

func toAppCountries(ctrys []countrybus.Country) []Country {
	app := make([]Country, len(ctrys))
	for i, ctry := range ctrys {
		app[i] = toAppCountry(ctry)
	}
	return app
}
