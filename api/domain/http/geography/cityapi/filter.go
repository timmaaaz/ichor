package cityapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/geography/cityapp"
)

func parseQueryParams(r *http.Request) (cityapp.QueryParams, error) {
	values := r.URL.Query()

	filter := cityapp.QueryParams{
		Page:     values.Get("page"),
		Rows:     values.Get("rows"),
		OrderBy:  values.Get("orderBy"),
		ID:       values.Get("id"),
		RegionID: values.Get("region_id"),
		Name:     values.Get("name"),
	}

	return filter, nil
}
