package city_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/geography/cityapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/location/cities/%s", sd.Cities[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &cityapp.UpdateCity{
				Name:     dbtest.StringPointer("Updated City"),
				RegionID: dbtest.StringPointer(sd.Regions[0].ID.String()),
			},
			GotResp: &cityapp.City{},
			ExpResp: &cityapp.City{
				Name:     "Updated City",
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

func update400(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "bad-region",
			URL:        fmt.Sprint("/v1/location/cities/asdf"),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &cityapp.UpdateCity{
				RegionID: dbtest.StringPointer("asdf"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "invalid UUID length: 4"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}
