package assettypeapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/assets/assettypeapp"
)

func parseQueryParams(r *http.Request) (assettypeapp.QueryParams, error) {
	values := r.URL.Query()

	filter := assettypeapp.QueryParams{
		Page:        values.Get("page"),
		Rows:        values.Get("rows"),
		OrderBy:     values.Get("orderBy"),
		ID:          values.Get("id"),
		Name:        values.Get("name"),
		Description: values.Get("description"),
	}

	return filter, nil
}
