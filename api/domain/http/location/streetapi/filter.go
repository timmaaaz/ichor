package streetapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/location/streetapp"
)

func parseQueryParams(r *http.Request) (streetapp.QueryParams, error) {
	values := r.URL.Query()

	filter := streetapp.QueryParams{
		Page:       values.Get("page"),
		Rows:       values.Get("rows"),
		OrderBy:    values.Get("orderBy"),
		ID:         values.Get("street_id"),
		CityID:     values.Get("city_id"),
		Line1:      values.Get("line_1"),
		Line2:      values.Get("line_2"),
		PostalCode: values.Get("postal_code"),
	}

	return filter, nil
}
