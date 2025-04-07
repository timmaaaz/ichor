package transferorderapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/movement/transferorderapp"
)

func parseQueryParams(r *http.Request) (transferorderapp.QueryParams, error) {
	values := r.URL.Query()

	filter := transferorderapp.QueryParams{
		Page:           values.Get("page"),
		Rows:           values.Get("rows"),
		OrderBy:        values.Get("orderBy"),
		TransferID:     values.Get("transfer_id"),
		FromLocationID: values.Get("from_location_id"),
		ToLocationID:   values.Get("to_location_id"),
		RequestedByID:  values.Get("requested_by_id"),
		ApprovedByID:   values.Get("approved_by_id"),
		Status:         values.Get("status"),
		UpdatedDate:    values.Get("updated_date"),
		TransferDate:   values.Get("transfer_date"),
		CreatedDate:    values.Get("created_date"),
		ProductID:      values.Get("product_id"),
		Quantity:       values.Get("quantity"),
	}

	return filter, nil
}
