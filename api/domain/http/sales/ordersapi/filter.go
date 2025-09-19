package ordersapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/sales/ordersapp"
)

func parseQueryParams(r *http.Request) (ordersapp.QueryParams, error) {
	values := r.URL.Query()

	filter := ordersapp.QueryParams{
		Page:                values.Get("page"),
		Rows:                values.Get("rows"),
		OrderBy:             values.Get("orderBy"),
		ID:                  values.Get("id"),
		Number:              values.Get("number"),
		CustomerID:          values.Get("customer_id"),
		FulfillmentStatusID: values.Get("fulfillment_status_id"),
		CreatedBy:           values.Get("created_by"),
		UpdatedBy:           values.Get("updated_by"),
		StartDueDate:        values.Get("start_due_date"),
		EndDueDate:          values.Get("end_due_date"),
		StartCreatedDate:    values.Get("start_created_date"),
		EndCreatedDate:      values.Get("end_created_date"),
		StartUpdatedDate:    values.Get("start_updated_date"),
		EndUpdatedDate:      values.Get("end_updated_date"),
	}

	return filter, nil
}
