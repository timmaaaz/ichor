package inventorylocationapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/inventory/inventorylocationapp"
)

func parseQueryParams(r *http.Request) (inventorylocationapp.QueryParams, error) {
	values := r.URL.Query()

	filter := inventorylocationapp.QueryParams{
		Page:               values.Get("page"),
		Rows:               values.Get("rows"),
		OrderBy:            values.Get("orderBy"),
		LocationID:         values.Get("location_id"),
		WarehouseID:        values.Get("warehouse_id"),
		ZoneID:             values.Get("zone_id"),
		Aisle:              values.Get("aisle"),
		Rack:               values.Get("rack"),
		Shelf:              values.Get("shelf"),
		Bin:                values.Get("bin"),
		LocationCode:       values.Get("location_code"),
		LocationCodeExact:  values.Get("location_code_exact"),
		IsPickLocation:     values.Get("is_pick_location"),
		IsReserveLocation:  values.Get("is_reserve_location"),
		MaxCapacity:        values.Get("max_capacity"),
		CurrentUtilization: values.Get("current_utilization"),
		CreatedDate:        values.Get("created_date"),
		UpdatedDate:        values.Get("updated_date"),
	}

	return filter, nil
}
