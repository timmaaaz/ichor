package labelapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/labels/labelapp"
)

func parseQueryParams(r *http.Request) (labelapp.QueryParams, error) {
	values := r.URL.Query()
	return labelapp.QueryParams{
		Page:    values.Get("page"),
		Rows:    values.Get("rows"),
		OrderBy: values.Get("orderBy"),
		Type:    values.Get("type"),
	}, nil
}
