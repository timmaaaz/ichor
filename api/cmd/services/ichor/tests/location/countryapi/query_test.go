package country_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/location/countryapp"
	"github.com/timmaaaz/ichor/app/sdk/query"
)

func query200(sd apitest.SeedData) []apitest.Table {

	expected := []countryapp.Country{
		{
			Name:   "United Arab Emirates",
			Number: 2,
			Alpha2: "AE",
			Alpha3: "ARE",
		},
		{
			Name:   "United Kingdom",
			Number: 76,
			Alpha2: "GB",
			Alpha3: "GBR",
		},
		{
			Name:   "Tanzania, United Republic Of",
			Number: 225,
			Alpha2: "TZ",
			Alpha3: "TZA",
		},
		{
			Name:   "United States Minor Outlying Islands",
			Number: 228,
			Alpha2: "UM",
			Alpha3: "UMI",
		},
		{
			Name:   "United States",
			Number: 229,
			Alpha2: "US",
			Alpha3: "USA",
		},
	}

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/countries?page=1&rows=10&orderBy=number,ASC&name=United",
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &query.Result[countryapp.Country]{},
			ExpResp: &query.Result[countryapp.Country]{
				Page:        1,
				RowsPerPage: 10,
				Total:       5,
				Items:       expected,
			},
			CmpFunc: func(got any, exp any) string {

				expResult := exp.(*query.Result[countryapp.Country])
				gotResult := got.(*query.Result[countryapp.Country])

				for i := range expResult.Items {
					expResult.Items[i].ID = gotResult.Items[i].ID
				}

				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func countryQueryByID200(sd apitest.SeedData) []apitest.Table {

	fmt.Println(sd.Countries[0])

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/countries/%s", sd.Countries[0].ID),
			Token:      sd.Users[0].Token, // Want to test this with no permission
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &countryapp.Country{},
			ExpResp: &countryapp.Country{
				ID:     sd.Countries[0].ID.String(),
				Number: sd.Countries[0].Number,
				Name:   sd.Countries[0].Name,
				Alpha2: sd.Countries[0].Alpha2,
				Alpha3: sd.Countries[0].Alpha3,
			},
			CmpFunc: func(got any, exp any) string {
				expCountry := exp.(*countryapp.Country)
				expCountry.ID = got.(*countryapp.Country).ID

				return cmp.Diff(got, expCountry)
			},
		},
	}
	return table
}
