package zoneapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/inventory/zoneapp"
)

func parseQueryParams(r *http.Request) (zoneapp.QueryParams, error) {
	values := r.URL.Query()

	filter := zoneapp.QueryParams{
		Page:        values.Get("page"),
		Rows:        values.Get("rows"),
		OrderBy:     values.Get("orderBy"),
		ZoneID:      values.Get("zone_id"),
		WarehouseID: values.Get("warehouse_id"),
		Name:        values.Get("name"),
		Description: values.Get("description"),
		CreatedDate: values.Get("created_date"),
		UpdatedDate: values.Get("updated_date"),
	}

	return filter, nil
}
