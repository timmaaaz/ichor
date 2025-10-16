package formapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/config/formapp"
)

func parseQueryParams(r *http.Request) (formapp.QueryParams, error) {
	values := r.URL.Query()

	qp := formapp.QueryParams{
		Page:    values.Get("page"),
		Rows:    values.Get("rows"),
		OrderBy: values.Get("orderBy"),
		ID:      values.Get("id"),
		Name:    values.Get("name"),
	}

	return qp, nil
}