package validassetdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/assets/validassetbus"
)

func applyFilter(filter validassetbus.QueryFilter, data map[string]interface{}, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["id"] = *filter.ID
		wc = append(wc, "id = :id")
	}

	if filter.TypeID != nil {
		data["type_id"] = *filter.TypeID
		wc = append(wc, "type_id = :type_id")
	}

	if filter.Name != nil {
		data["name"] = "%" + *filter.Name + "%"
		wc = append(wc, "name ILIKE :name")
	}

	if filter.EstPrice != nil {
		data["est_price"] = *filter.EstPrice
		wc = append(wc, "est_price = :est_price")
	}

	if filter.Price != nil {
		data["price"] = *filter.Price
		wc = append(wc, "price = :price")
	}

	if filter.MaintenanceInterval != nil {
		data["maintenance_interval"] = *filter.MaintenanceInterval
		wc = append(wc, "maintenance_interval = :maintenance_interval")
	}

	if filter.LifeExpectancy != nil {
		data["life_expectancy"] = *filter.LifeExpectancy
		wc = append(wc, "life_expectancy = :life_expectancy")
	}

	if filter.SerialNumber != nil {
		data["serial_number"] = *filter.SerialNumber
		wc = append(wc, "serial_number = :serial_number")
	}

	if filter.ModelNumber != nil {
		data["model_number"] = *filter.ModelNumber
		wc = append(wc, "model_number = :model_number")
	}

	if filter.IsEnabled != nil {
		data["is_enabled"] = *filter.IsEnabled
		wc = append(wc, "is_enabled = :is_enabled")
	}

	if filter.StartCreatedDate != nil {
		data["start_created_date"] = *filter.StartCreatedDate
		wc = append(wc, "created_date >= :start_created_date")
	}

	if filter.EndCreatedDate != nil {
		data["end_created_date"] = *filter.EndCreatedDate
		wc = append(wc, "created_date <= :end_created_date")
	}

	if filter.StartUpdatedDate != nil {
		data["start_updated_date"] = *filter.StartUpdatedDate
		wc = append(wc, "updated_date >= :start_updated_date")
	}

	if filter.EndUpdatedDate != nil {
		data["end_updated_date"] = *filter.EndUpdatedDate
		wc = append(wc, "updated_date <= :end_updated_date")
	}

	if filter.CreatedBy != nil {
		data["created_by"] = *filter.CreatedBy
		wc = append(wc, "created_by = :created_by")
	}

	if filter.UpdatedBy != nil {
		data["updated_by"] = *filter.UpdatedBy
		wc = append(wc, "updated_by = :updated_by")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
