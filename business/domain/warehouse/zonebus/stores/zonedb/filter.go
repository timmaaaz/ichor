package zonedb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/warehouse/zonebus"
)

func applyFilter(filter zonebus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["warehouse_id"] = *filter.ID
		wc = append(wc, "warehouse_id = :warehouse_id")
	}

	if filter.WarehouseID != nil {
		data["warehouse_id"] = *filter.WarehouseID
		wc = append(wc, "warehouse_id = :warehouse_id")
	}

	if filter.Name != nil {
		data["name"] = *filter.Name
		wc = append(wc, "name ILIKE :name")
	}

	if filter.Description != nil {
		data["description"] = *filter.Description
		wc = append(wc, "description ILIKE :description")
	}

	if filter.IsActive != nil {
		data["is_active"] = *filter.IsActive
		wc = append(wc, "is_active = :is_active")
	}

	if filter.CreatedBy != nil {
		data["created_by"] = *filter.CreatedBy
		wc = append(wc, "created_by = :created_by")
	}

	if filter.UpdatedBy != nil {
		data["updated_by"] = *filter.UpdatedBy
		wc = append(wc, "updated_by = :updated_by")
	}

	if filter.StartCreatedDate != nil {
		data["start_date_created"] = filter.StartCreatedDate.UTC()
		wc = append(wc, "date_created >= :start_date_created")
	}

	if filter.EndCreatedDate != nil {
		data["end_date_created"] = filter.EndCreatedDate.UTC()
		wc = append(wc, "date_created <= :end_date_created")
	}

	if filter.StartUpdatedDate != nil {
		data["start_date_updated"] = filter.StartUpdatedDate.UTC()
		wc = append(wc, "date_updated >= :start_date_updated")
	}

	if filter.EndUpdatedDate != nil {
		data["end_date_updated"] = filter.EndUpdatedDate.UTC()
		wc = append(wc, "date_updated <= :end_date_updated")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
