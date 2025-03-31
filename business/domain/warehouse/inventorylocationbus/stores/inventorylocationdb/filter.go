package inventorylocationdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/warehouse/inventorylocationbus"
)

func applyFilter(filter inventorylocationbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.LocationID != nil {
		data["location_id"] = *filter.LocationID
		wc = append(wc, "location_id = :location_id")
	}

	if filter.ZoneID != nil {
		data["zone_id"] = *filter.ZoneID
		wc = append(wc, "zone_id = :zone_id")
	}

	if filter.WarehouseID != nil {
		data["warehouse_id"] = *filter.WarehouseID
		wc = append(wc, "warehouse_id = :warehouse_id")
	}

	if filter.Aisle != nil {
		data["aisle"] = *filter.Aisle
		wc = append(wc, "aisle = :aisle")
	}

	if filter.Rack != nil {
		data["rack"] = *filter.Rack
		wc = append(wc, "rack = :rack")
	}

	if filter.Shelf != nil {
		data["shelf"] = *filter.Shelf
		wc = append(wc, "shelf = :shelf")
	}

	if filter.Bin != nil {
		data["bin"] = *filter.Bin
		wc = append(wc, "bin = :bin")
	}

	if filter.IsPickLocation != nil {
		data["is_pick_location"] = *filter.IsPickLocation
		wc = append(wc, "is_pick_location = :is_pick_location")
	}

	if filter.IsReserveLocation != nil {
		data["is_reserve_location"] = *filter.IsReserveLocation
		wc = append(wc, "is_reserve_location = :is_reserve_location")
	}

	if filter.MaxCapacity != nil {
		data["max_capacity"] = *filter.MaxCapacity
		wc = append(wc, "max_capacity = :max_capacity")
	}

	if filter.CurrentUtilization != nil {
		data["current_utilization"] = *filter.CurrentUtilization
		wc = append(wc, "current_utilization = :current_utilization")
	}

	if filter.CreatedDate != nil {
		data["created_date"] = *filter.CreatedDate
		wc = append(wc, "created_date = :created_date")
	}

	if filter.UpdatedDate != nil {
		data["updated_date"] = *filter.UpdatedDate
		wc = append(wc, "updated_date = :updated_date")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
