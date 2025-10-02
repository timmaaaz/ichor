package regionapp

import (
	"encoding/json"

	"github.com/timmaaaz/ichor/business/domain/geography/regionbus"
)

type QueryParams struct {
	Page      string
	Rows      string
	OrderBy   string
	ID        string
	Name      string
	CountryID string
	Code      string
}

// Region is a struct that represents a region in the database.
type Region struct {
	ID        string `json:"id"`
	CountryID string `json:"country_id"`
	Name      string `json:"name"`
	Code      string `json:"code"`
}

// Encode implements the encoder interface.
func (app Region) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppRegion(rgn regionbus.Region) Region {
	return Region{
		ID:        rgn.ID.String(),
		CountryID: rgn.CountryID.String(),
		Name:      rgn.Name,
		Code:      rgn.Code,
	}
}

func ToAppRegions(rgns []regionbus.Region) []Region {
	app := make([]Region, len(rgns))
	for i, rgn := range rgns {
		app[i] = ToAppRegion(rgn)
	}
	return app
}
