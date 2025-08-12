package orderfulfillmentstatusapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/order/orderfulfillmentstatusapp"
)

func parseQueryParams(r *http.Request) (orderfulfillmentstatusapp.QueryParams, error) {
	values := r.URL.Query()

	filter := orderfulfillmentstatusapp.QueryParams{
		Page:        values.Get("page"),
		Rows:        values.Get("rows"),
		OrderBy:     values.Get("orderBy"),
		ID:          values.Get("id"),
		Name:        values.Get("name"),
		Description: values.Get("description"),
	}

	return filter, nil
}
