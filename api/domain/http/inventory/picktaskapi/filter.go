package picktaskapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/inventory/picktaskapp"
)

func parseQueryParams(r *http.Request) (picktaskapp.QueryParams, error) {
	values := r.URL.Query()

	qp := picktaskapp.QueryParams{
		Page:                 values.Get("page"),
		Rows:                 values.Get("rows"),
		OrderBy:              values.Get("orderBy"),
		ID:                   values.Get("id"),
		SalesOrderID:         values.Get("sales_order_id"),
		SalesOrderLineItemID: values.Get("sales_order_line_item_id"),
		ProductID:            values.Get("product_id"),
		LocationID:           values.Get("location_id"),
		Status:               values.Get("status"),
		AssignedTo:           values.Get("assigned_to"),
		CreatedBy:            values.Get("created_by"),
	}

	return qp, nil
}
