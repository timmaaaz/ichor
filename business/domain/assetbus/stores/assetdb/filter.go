package assetdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/assetbus"
)

func applyFilter(filter assetbus.QueryFilter, data map[string]interface{}, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["asset_id"] = *filter.ID
		wc = append(wc, "asset_id = :asset_id")
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

	if filter.ModelNumber != nil {
		data["model_number"] = *filter.ModelNumber
		wc = append(wc, "model_number = :model_number")
	}

	if filter.IsEnabled != nil {
		data["is_enabled"] = *filter.IsEnabled
		wc = append(wc, "is_enabled = :is_enabled")
	}

	if filter.StartDateCreated != nil {
		data["start_date_created"] = *filter.StartDateCreated
		wc = append(wc, "date_created >= :start_date_created")
	}

	if filter.EndDateCreated != nil {
		data["end_date_created"] = *filter.EndDateCreated
		wc = append(wc, "date_created <= :end_date_created")
	}

	if filter.StartDateUpdated != nil {
		data["start_date_updated"] = *filter.StartDateUpdated
		wc = append(wc, "date_updated >= :start_date_updated")
	}

	if filter.EndDateUpdated != nil {
		data["end_date_updated"] = *filter.EndDateUpdated
		wc = append(wc, "date_updated <= :end_date_updated")
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
