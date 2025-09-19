package inventorytransactionapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/inventory/inventorytransactionapp"
)

func parseQueryParams(r *http.Request) (inventorytransactionapp.QueryParams, error) {
	values := r.URL.Query()

	filter := inventorytransactionapp.QueryParams{
		Page:                   values.Get("page"),
		Rows:                   values.Get("rows"),
		OrderBy:                values.Get("orderBy"),
		InventoryTransactionID: values.Get("transaction_id"),
		ProductID:              values.Get("product_id"),
		LocationID:             values.Get("location_id"),
		UserID:                 values.Get("user_id"),
		Quantity:               values.Get("quantity"),
		TransactionType:        values.Get("transaction_type"),
		ReferenceNumber:        values.Get("reference_number"),
		TransactionDate:        values.Get("transaction_date"),
		CreatedDate:            values.Get("created_date"),
		UpdatedDate:            values.Get("updated_date"),
	}

	return filter, nil
}
