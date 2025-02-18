package contactinfoapi_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"

	"github.com/timmaaaz/ichor/app/domain/core/contactinfoapp"
	"github.com/timmaaaz/ichor/app/sdk/query"
)

func query200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/core/contactinfo?page=1&rows=10",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &query.Result[contactinfoapp.ContactInfo]{},
			ExpResp: &query.Result[contactinfoapp.ContactInfo]{
				Page:        1,
				RowsPerPage: 10,
				Total:       15,
				Items:       sd.ContactInfo[:10],
			},
			CmpFunc: func(got any, exp any) string {
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
			URL:        "/v1/core/contactinfo/" + sd.ContactInfo[0].ID,
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &contactinfoapp.ContactInfo{},
			ExpResp:    &sd.ContactInfo[0],
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}
