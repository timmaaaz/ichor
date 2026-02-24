package putawaytaskapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/inventory/putawaytaskapp"
)

func parseQueryParams(r *http.Request) (putawaytaskapp.QueryParams, error) {
	values := r.URL.Query()

	qp := putawaytaskapp.QueryParams{
		Page:            values.Get("page"),
		Rows:            values.Get("rows"),
		OrderBy:         values.Get("orderBy"),
		ID:              values.Get("id"),
		ProductID:       values.Get("product_id"),
		LocationID:      values.Get("location_id"),
		Status:          values.Get("status"),
		AssignedTo:      values.Get("assigned_to"),
		CreatedBy:       values.Get("created_by"),
		ReferenceNumber: values.Get("reference_number"),
	}

	return qp, nil
}
