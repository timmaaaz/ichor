package cityapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/location/cityapp"
)

func parseQueryParams(r *http.Request) (cityapp.QueryParams, error) {
	values := r.URL.Query()

	filter := cityapp.QueryParams{
		Page:     values.Get("page"),
		Rows:     values.Get("row"),
		OrderBy:  values.Get("orderBy"),
		ID:       values.Get("city_id"),
		RegionID: values.Get("region_id"),
		Name:     values.Get("name"),
	}

	return filter, nil
}
