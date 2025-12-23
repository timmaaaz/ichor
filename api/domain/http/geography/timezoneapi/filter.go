package timezoneapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/geography/timezoneapp"
)

func parseQueryParams(r *http.Request) (timezoneapp.QueryParams, error) {
	values := r.URL.Query()

	filter := timezoneapp.QueryParams{
		Page:        values.Get("page"),
		Rows:        values.Get("rows"),
		OrderBy:     values.Get("orderBy"),
		ID:          values.Get("id"),
		Name:        values.Get("name"),
		DisplayName: values.Get("displayName"),
		IsActive:    values.Get("isActive"),
	}

	return filter, nil
}
