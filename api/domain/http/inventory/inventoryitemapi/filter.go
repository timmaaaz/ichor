package inventoryitemapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/inventory/inventoryitemapp"
)

func parseQueryParams(r *http.Request) (inventoryitemapp.QueryParams, error) {
	values := r.URL.Query()

	filter := inventoryitemapp.QueryParams{
		Page:                  values.Get("page"),
		Rows:                  values.Get("rows"),
		OrderBy:               values.Get("orderBy"),
		ItemID:                values.Get("item_id"),
		ProductID:             values.Get("product_id"),
		LocationID:            values.Get("location_id"),
		Quantity:              values.Get("quantity"),
		ReservedQuantity:      values.Get("reserved_quantity"),
		AllocatedQuantity:     values.Get("allocated_quantity"),
		MinimumStock:          values.Get("minimum_stock"),
		MaximumStock:          values.Get("maximum_stock"),
		ReorderPoint:          values.Get("reorder_point"),
		EconomicOrderQuantity: values.Get("economic_order_quantity"),
		SafetyStock:           values.Get("safety_stock"),
		AvgDailyUsage:         values.Get("avg_daily_usage"),
		CreatedDate:           values.Get("created_date"),
		UpdatedDate:           values.Get("updated_date"),
	}

	return filter, nil
}
