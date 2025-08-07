package countryapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/location/countryapp"
)

func parseQueryParams(r *http.Request) (countryapp.QueryParams, error) {
	values := r.URL.Query()

	filter := countryapp.QueryParams{
		Page:    values.Get("page"),
		Rows:    values.Get("rows"),
		OrderBy: values.Get("orderBy"),
		ID:      values.Get("id"),
		Name:    values.Get("name"),
		Alpha2:  values.Get("alpha_2"),
		Alpha3:  values.Get("alpha_3"),
	}

	return filter, nil
}
