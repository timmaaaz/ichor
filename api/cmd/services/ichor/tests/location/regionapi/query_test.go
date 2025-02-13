package region_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/location/regionapp"
	"github.com/timmaaaz/ichor/app/sdk/query"
)

func query200(sd apitest.SeedData) []apitest.Table {
	expected := []regionapp.Region{
		{
			Name: "Alabama",
			Code: "AL",
		},
		{
			Name: "Alaska",
			Code: "AK",
		},
		{
			Name: "Arizona",
			Code: "AZ",
		},
		{
			Name: "Arkansas",
			Code: "AR",
		},
		{
			Name: "California",
			Code: "CA",
		},
	}

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/location/regions?page=1&rows=5&orderBy=name,ASC",
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &query.Result[regionapp.Region]{},
			ExpResp: &query.Result[regionapp.Region]{
				Page:        1,
				RowsPerPage: 5,
				Total:       50,
				Items:       expected,
			},
			CmpFunc: func(got any, exp any) string {
				gotResult := got.(*query.Result[regionapp.Region])
				expResult := exp.(*query.Result[regionapp.Region])

				for i := range expResult.Items {
					expResult.Items[i].ID = gotResult.Items[i].ID
					expResult.Items[i].CountryID = gotResult.Items[i].CountryID
				}

				return cmp.Diff(gotResult, expResult)

			},
		},
	}

	return table
}

func queryByID200(sd apitest.SeedData) []apitest.Table {
	expected := regionapp.Region{
		Name: "Alabama",
		Code: "AL",
	}

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/location/regions/%s", sd.Regions[0].ID),
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &regionapp.Region{},
			ExpResp:    &expected,
			CmpFunc: func(got any, exp any) string {
				gotResp := got.(*regionapp.Region)
				expResp := exp.(*regionapp.Region)

				expResp.ID = gotResp.ID
				expResp.CountryID = gotResp.CountryID

				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
