package purchaseorderlineitemstatusapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/procurement/purchaseorderlineitemstatusapp"
)

func parseQueryParams(r *http.Request) (purchaseorderlineitemstatusapp.QueryParams, error) {
	values := r.URL.Query()

	filter := purchaseorderlineitemstatusapp.QueryParams{
		Page:        values.Get("page"),
		Rows:        values.Get("rows"),
		OrderBy:     values.Get("orderBy"),
		ID:          values.Get("id"),
		Name:        values.Get("name"),
		Description: values.Get("description"),
	}

	return filter, nil
}
