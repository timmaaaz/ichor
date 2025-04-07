package inventoryadjustmentapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/movement/inventoryadjustmentapp"
)

func parseQueryParams(r *http.Request) (inventoryadjustmentapp.QueryParams, error) {
	values := r.URL.Query()

	filter := inventoryadjustmentapp.QueryParams{
		Page:                  values.Get("page"),
		Rows:                  values.Get("rows"),
		OrderBy:               values.Get("orderBy"),
		InventoryAdjustmentID: values.Get("adjustment_id"),
		ProductID:             values.Get("product_id"),
		LocationID:            values.Get("location_id"),
		AdjustedBy:            values.Get("adjusted_by"),
		ApprovedBy:            values.Get("approved_by"),
		QuantityChange:        values.Get("quantity_change"),
		ReasonCode:            values.Get("reason_code"),
		Notes:                 values.Get("notes"),
		AdjustmentDate:        values.Get("adjustment_date"),
		CreatedDate:           values.Get("created_date"),
		UpdatedDate:           values.Get("updated_date"),
	}

	return filter, nil
}
