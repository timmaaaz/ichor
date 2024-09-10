package countryapi

import (
	"net/http"

	"bitbucket.org/superiortechnologies/ichor/app/domain/location/countryapp"
)

func parseQueryParams(r *http.Request) (countryapp.QueryParams, error) {
	values := r.URL.Query()

	filter := countryapp.QueryParams{
		Page:    values.Get("page"),
		Rows:    values.Get("row"),
		OrderBy: values.Get("orderBy"),
		ID:      values.Get("country_id"),
		Name:    values.Get("name"),
		Alpha2:  values.Get("alpha_2"),
		Alpha3:  values.Get("alpha_3"),
	}

	return filter, nil
}
