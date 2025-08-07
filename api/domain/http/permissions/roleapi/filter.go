package roleapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/permissions/roleapp"
)

func parseQueryParams(r *http.Request) (roleapp.QueryParams, error) {
	values := r.URL.Query()

	filter := roleapp.QueryParams{
		Page:        values.Get("page"),
		Rows:        values.Get("rows"),
		OrderBy:     values.Get("orderBy"),
		ID:          values.Get("id"),
		Name:        values.Get("name"),
		Description: values.Get("description"),
	}

	return filter, nil
}
