package lineitemfulfillmentstatusapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/order/lineitemfulfillmentstatusapp"
)

func parseQueryParams(r *http.Request) (lineitemfulfillmentstatusapp.QueryParams, error) {
	values := r.URL.Query()

	filter := lineitemfulfillmentstatusapp.QueryParams{
		Page:        values.Get("page"),
		Rows:        values.Get("rows"),
		OrderBy:     values.Get("orderBy"),
		ID:          values.Get("id"),
		Name:        values.Get("name"),
		Description: values.Get("description"),
	}

	return filter, nil
}
