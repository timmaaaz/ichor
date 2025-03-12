package street_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/location/streetapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/location/streets/%s", sd.Streets[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &streetapp.UpdateStreet{
				CityID:     dbtest.StringPointer(sd.Cities[0].ID),
				Line1:      dbtest.StringPointer("Updated Street"),
				Line2:      dbtest.StringPointer("Updated Street Line 2"),
				PostalCode: dbtest.StringPointer("54321"),
			},
			GotResp: &streetapp.Street{},
			ExpResp: &streetapp.Street{
				CityID:     sd.Cities[0].ID,
				Line1:      "Updated Street",
				Line2:      "Updated Street Line 2",
				PostalCode: "54321",
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

func update400(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "missing city",
			URL:        fmt.Sprintf("/v1/location/streets/%s", sd.Streets[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &streetapp.UpdateStreet{
				CityID: dbtest.StringPointer("asdf"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "parse: invalid UUID length: 4"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}
