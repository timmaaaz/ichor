package lotlocationapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/inventory/lotlocationapp"
)

func parseQueryParams(r *http.Request) (lotlocationapp.QueryParams, error) {
	values := r.URL.Query()

	return lotlocationapp.QueryParams{
		Page:       values.Get("page"),
		Rows:       values.Get("rows"),
		OrderBy:    values.Get("orderBy"),
		ID:         values.Get("id"),
		LotID:      values.Get("lot_id"),
		LocationID: values.Get("location_id"),
	}, nil
}
