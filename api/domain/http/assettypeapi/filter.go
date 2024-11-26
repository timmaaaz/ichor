package assettypeapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/assettypeapp"
)

func parseQueryParams(r *http.Request) (assettypeapp.QueryParams, error) {
	values := r.URL.Query()

	filter := assettypeapp.QueryParams{
		Page:    values.Get("page"),
		Rows:    values.Get("row"),
		OrderBy: values.Get("orderBy"),
		ID:      values.Get("approval_status_id"),
		Name:    values.Get("name"),
	}

	return filter, nil
}
