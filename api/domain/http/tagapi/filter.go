package tagapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/tagapp"
)

func parseQueryParams(r *http.Request) (tagapp.QueryParams, error) {
	values := r.URL.Query()

	filter := tagapp.QueryParams{
		Page:        values.Get("page"),
		Rows:        values.Get("rows"),
		OrderBy:     values.Get("orderBy"),
		ID:          values.Get("asset_condition_id"),
		Name:        values.Get("name"),
		Description: values.Get("description"),
	}

	return filter, nil
}
