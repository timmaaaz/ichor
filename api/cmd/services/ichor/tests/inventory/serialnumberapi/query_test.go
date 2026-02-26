package serialnumber_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"

	"github.com/timmaaaz/ichor/app/domain/inventory/serialnumberapp"
	"github.com/timmaaaz/ichor/app/sdk/query"
)

func query200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/inventory/serial-numbers?page=1&rows=10",
			Token:      sd.Users[0].Token,
			StatusCode: 200,
			Method:     "GET",
			GotResp:    &query.Result[serialnumberapp.SerialNumber]{},
			ExpResp: &query.Result[serialnumberapp.SerialNumber]{
				Page:        1,
				RowsPerPage: 10,
				Total:       20,
				Items:       sd.SerialNumbers[:10],
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func queryByID200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/inventory/serial-numbers/" + sd.SerialNumbers[0].SerialID,
			Token:      sd.Users[0].Token,
			StatusCode: 200,
			Method:     "GET",
			GotResp:    &serialnumberapp.SerialNumber{},
			ExpResp:    &sd.SerialNumbers[0],
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

// serialLocationIDOnly captures just location_id for assertions; warehouse_name
// and zone_name are random seed values we cannot predict deterministically.
type serialLocationIDOnly struct {
	LocationID string `json:"location_id"`
}

func queryLocation200(sd apitest.SeedData) []apitest.Table {
	sn := sd.SerialNumbers[0]
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/inventory/serial-numbers/" + sn.SerialID + "/location",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &serialLocationIDOnly{},
			ExpResp:    &serialLocationIDOnly{LocationID: sn.LocationID},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got.(*serialLocationIDOnly).LocationID, exp.(*serialLocationIDOnly).LocationID)
			},
		},
	}
}
