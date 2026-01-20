package currencyapi_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

func delete200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/core/currencies/" + sd.Currencies[1].ID.String(),
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusNoContent,
			Method:     http.MethodDelete,
			GotResp:    nil,
			ExpResp:    nil,
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
