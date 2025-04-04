package serialnumberapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/lots/serialnumberapp"
)

func parseQueryParams(r *http.Request) (serialnumberapp.QueryParams, error) {
	values := r.URL.Query()

	filter := serialnumberapp.QueryParams{
		Page:         values.Get("page"),
		Rows:         values.Get("rows"),
		OrderBy:      values.Get("orderBy"),
		LotID:        values.Get("lot_id"),
		SerialNumber: values.Get("serial_number"),
		SerialID:     values.Get("serial_id"),
		ProductID:    values.Get("product_id"),
		LocationID:   values.Get("location_id"),
		Status:       values.Get("status"),
		CreatedDate:  values.Get("created_date"),
		UpdatedDate:  values.Get("updated_date"),
	}

	return filter, nil
}
