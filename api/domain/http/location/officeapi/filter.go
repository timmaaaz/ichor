package officeapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/location/officeapp"
)

func parseQueryParams(r *http.Request) (officeapp.QueryParams, error) {
	values := r.URL.Query()

	filter := officeapp.QueryParams{
		Page:     values.Get("page"),
		Rows:     values.Get("rows"),
		OrderBy:  values.Get("orderBy"),
		ID:       values.Get("id"),
		Name:     values.Get("name"),
		StreetID: values.Get("street_id"),
	}

	return filter, nil
}
