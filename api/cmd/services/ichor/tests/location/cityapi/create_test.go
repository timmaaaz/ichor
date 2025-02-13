package city_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/location/cityapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/location/cities",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &cityapp.NewCity{
				Name:     "New City",
				RegionID: sd.Regions[0].ID.String(),
			},
			GotResp: &cityapp.City{},
			ExpResp: &cityapp.City{
				Name:     "New City",
				RegionID: sd.Regions[0].ID.String(),
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*cityapp.City)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*cityapp.City)
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
			Name:       "missing region",
			URL:        "/v1/location/cities",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &cityapp.NewCity{
				Name: "New City",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"regionID\",\"error\":\"regionID is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing name",
			URL:        "/v1/location/cities",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &cityapp.NewCity{
				RegionID: sd.Regions[0].ID.String(),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"name\",\"error\":\"name is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
