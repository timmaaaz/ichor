package street_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/location/streetapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/streets",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &streetapp.NewStreet{
				CityID:     sd.Cities[0].ID,
				Line1:      "New Street",
				Line2:      "New Street Line 2",
				PostalCode: "12345",
			},
			GotResp: &streetapp.Street{},
			ExpResp: &streetapp.Street{
				CityID:     sd.Cities[0].ID,
				Line1:      "New Street",
				Line2:      "New Street Line 2",
				PostalCode: "12345",
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*streetapp.Street)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*streetapp.Street)
				expResp.ID = gotResp.ID

				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func create400(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "missing city",
			URL:        "/v1/streets",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &streetapp.NewStreet{
				Line1:      "New Street",
				Line2:      "New Street Line 2",
				PostalCode: "12345",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"city_id\",\"error\":\"city_id is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing line 1",
			URL:        "/v1/streets",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &streetapp.NewStreet{
				CityID:     sd.Cities[0].ID,
				Line2:      "New Street Line 2",
				PostalCode: "12345",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"line_1\",\"error\":\"line_1 is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "line 1 over 100 chars",
			URL:        "/v1/streets",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &streetapp.NewStreet{
				CityID:     sd.Cities[0].ID,
				Line1:      "1234567890 1234567890 1234567890 1234567890 1234567890 1234567890 1234567890 1234567890 1234567890 1234567890",
				Line2:      "New Street Line 2",
				PostalCode: "12345",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"line_1\",\"error\":\"line_1 must be a maximum of 100 characters in length\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "line 2 over 100 chars",
			URL:        "/v1/streets",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &streetapp.NewStreet{
				CityID:     sd.Cities[0].ID,
				Line1:      "New Street",
				Line2:      "1234567890 1234567890 1234567890 1234567890 1234567890 1234567890 1234567890 1234567890 1234567890 1234567890",
				PostalCode: "12345",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"line_2\",\"error\":\"line_2 must be a maximum of 100 characters in length\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
