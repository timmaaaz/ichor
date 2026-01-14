package paymenttermapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/core/paymenttermapp"
)

func parseQueryParams(r *http.Request) (paymenttermapp.QueryParams, error) {
	values := r.URL.Query()

	filter := paymenttermapp.QueryParams{
		Page:        values.Get("page"),
		Rows:        values.Get("rows"),
		OrderBy:     values.Get("orderBy"),
		ID:          values.Get("id"),
		Name:        values.Get("name"),
		Description: values.Get("description"),
	}

	return filter, nil
}
