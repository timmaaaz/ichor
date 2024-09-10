package regionapi

import (
	"net/http"

	"bitbucket.org/superiortechnologies/ichor/app/domain/location/regionapp"
)

func parseQueryParams(r *http.Request) (regionapp.QueryParams, error) {
	values := r.URL.Query()

	filter := regionapp.QueryParams{
		Page:      values.Get("page"),
		Rows:      values.Get("row"),
		OrderBy:   values.Get("orderBy"),
		ID:        values.Get("region_id"),
		CountryID: values.Get("country_id"),
		Name:      values.Get("name"),
		Code:      values.Get("code"),
	}

	return filter, nil
}
