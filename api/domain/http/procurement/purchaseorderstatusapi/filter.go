package purchaseorderstatusapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/procurement/purchaseorderstatusapp"
)

func parseQueryParams(r *http.Request) (purchaseorderstatusapp.QueryParams, error) {
	values := r.URL.Query()

	filter := purchaseorderstatusapp.QueryParams{
		Page:        values.Get("page"),
		Rows:        values.Get("rows"),
		OrderBy:     values.Get("orderBy"),
		ID:          values.Get("id"),
		Name:        values.Get("name"),
		Description: values.Get("description"),
	}

	return filter, nil
}
