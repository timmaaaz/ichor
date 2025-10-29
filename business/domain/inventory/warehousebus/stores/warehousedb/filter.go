package warehousedb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/inventory/warehousebus"
)

func applyFilter(filter warehousebus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["id"] = *filter.ID
		wc = append(wc, "id = :id")
	}

	if filter.Code != nil {
		data["code"] = *filter.Code
		wc = append(wc, "code ILIKE :code")
	}

	if filter.Name != nil {
		data["name"] = *filter.Name
		wc = append(wc, "name ILIKE :name")
	}

	if filter.StreetID != nil {
		data["street_id"] = *filter.StreetID
		wc = append(wc, "street_id = :street_id")
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
		data["start_created_date"] = filter.StartCreatedDate.UTC()
		wc = append(wc, "created_date >= :start_created_date")
	}

	if filter.EndCreatedDate != nil {
		data["end_created_date"] = filter.EndCreatedDate.UTC()
		wc = append(wc, "created_date <= :end_created_date")
	}

	if filter.StartUpdatedDate != nil {
		data["start_updated_date"] = filter.StartUpdatedDate.UTC()
		wc = append(wc, "updated_date >= :start_updated_date")
	}

	if filter.EndUpdatedDate != nil {
		data["end_updated_date"] = filter.EndUpdatedDate.UTC()
		wc = append(wc, "updated_date <= :end_updated_date")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
