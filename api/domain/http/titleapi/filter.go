package titleapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/titleapp"
)

func parseQueryParams(r *http.Request) (titleapp.QueryParams, error) {
	values := r.URL.Query()

	filter := titleapp.QueryParams{
		Page:        values.Get("page"),
		Rows:        values.Get("rows"),
		OrderBy:     values.Get("orderBy"),
		ID:          values.Get("title_id"),
		Name:        values.Get("name"),
		Description: values.Get("description"),
	}

	return filter, nil
}
