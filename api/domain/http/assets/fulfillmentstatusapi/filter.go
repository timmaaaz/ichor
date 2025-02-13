package fulfillmentstatusapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/assets/fulfillmentstatusapp"
)

func parseQueryParams(r *http.Request) (fulfillmentstatusapp.QueryParams, error) {
	values := r.URL.Query()

	filter := fulfillmentstatusapp.QueryParams{
		Page:    values.Get("page"),
		Rows:    values.Get("rows"),
		OrderBy: values.Get("orderBy"),
		ID:      values.Get("fulfillment_status_id"),
		Name:    values.Get("name"),
		IconID:  values.Get("icon_id"),
	}

	return filter, nil
}
