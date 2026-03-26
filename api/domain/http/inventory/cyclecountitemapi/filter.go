package cyclecountitemapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/inventory/cyclecountitemapp"
)

func parseQueryParams(r *http.Request) (cyclecountitemapp.QueryParams, error) {
	values := r.URL.Query()

	qp := cyclecountitemapp.QueryParams{
		Page:       values.Get("page"),
		Rows:       values.Get("rows"),
		OrderBy:    values.Get("orderBy"),
		ID:         values.Get("id"),
		SessionID:  values.Get("session_id"),
		ProductID:  values.Get("product_id"),
		LocationID: values.Get("location_id"),
		Status:     values.Get("status"),
		CountedBy:  values.Get("counted_by"),
	}

	return qp, nil
}
