package orderlineitemsapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/sales/orderlineitemsapp"
)

func parseQueryParams(r *http.Request) (orderlineitemsapp.QueryParams, error) {
	values := r.URL.Query()

	filter := orderlineitemsapp.QueryParams{
		Page:                          values.Get("page"),
		Rows:                          values.Get("rows"),
		OrderBy:                       values.Get("orderBy"),
		ID:                            values.Get("id"),
		OrderID:                       values.Get("order_id"),
		ProductID:                     values.Get("product_id"),
		Description:                   values.Get("description"),
		Quantity:                      values.Get("quantity"),
		UnitPrice:                     values.Get("unit_price"),
		Discount:                      values.Get("discount"),
		DiscountType:                  values.Get("discount_type"),
		LineTotal:                     values.Get("line_total"),
		LineItemFulfillmentStatusesID: values.Get("line_item_fulfillment_statuses_id"),
		CreatedBy:                     values.Get("created_by"),
		StartCreatedDate:              values.Get("start_created_date"),
		EndCreatedDate:                values.Get("end_created_date"),
		UpdatedBy:                     values.Get("updated_by"),
		StartUpdatedDate:              values.Get("start_updated_date"),
		EndUpdatedDate:                values.Get("end_updated_date"),
	}

	return filter, nil
}
