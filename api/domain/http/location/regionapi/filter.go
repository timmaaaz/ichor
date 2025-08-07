package regionapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/location/regionapp"
)

func parseQueryParams(r *http.Request) (regionapp.QueryParams, error) {
	values := r.URL.Query()

	filter := regionapp.QueryParams{
		Page:      values.Get("page"),
		Rows:      values.Get("rows"),
		OrderBy:   values.Get("orderBy"),
		ID:        values.Get("id"),
		CountryID: values.Get("country_id"),
		Name:      values.Get("name"),
		Code:      values.Get("code"),
	}

	return filter, nil
}
