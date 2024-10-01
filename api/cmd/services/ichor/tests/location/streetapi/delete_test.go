package street_test

import (
	"fmt"
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

func delete200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "asuser",
			URL:        fmt.Sprintf("/v1/streets/%s", sd.Streets[1].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodDelete,
			StatusCode: http.StatusNoContent,
		},
		{
			Name:       "asadmin",
			URL:        fmt.Sprintf("/v1/streets/%s", sd.Streets[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodDelete,
			StatusCode: http.StatusNoContent,
		},
	}

	return table
}
