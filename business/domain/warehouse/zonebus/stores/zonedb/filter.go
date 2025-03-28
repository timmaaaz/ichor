package zonedb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/warehouse/zonebus"
)

func applyFilter(filter zonebus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ZoneID != nil {
		data["zone_id"] = *filter.ZoneID
		wc = append(wc, "zone_id = :zone_id")
	}

	if filter.WarehouseID != nil {
		data["warehouse_id"] = *filter.WarehouseID
		wc = append(wc, "warehouse_id = :warehouse_id")
	}

	if filter.Name != nil {
		data["name"] = "%" + *filter.Name + "%"
		wc = append(wc, "name ILIKE :name")
	}

	if filter.Description != nil {
		data["description"] = "%" + *filter.Description + "%"
		wc = append(wc, "description ILIKE :description")
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
