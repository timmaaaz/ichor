package serialnumberdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/inventory/serialnumberbus"
)

func applyFilter(filter serialnumberbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.SerialID != nil {
		data["id"] = *filter.SerialID
		wc = append(wc, "id = :id")
	}

	if filter.LotID != nil {
		data["lot_id"] = *filter.LotID
		wc = append(wc, "lot_id = :lot_id")
	}

	if filter.ProductID != nil {
		data["product_id"] = *filter.ProductID
		wc = append(wc, "product_id = :product_id")
	}

	if filter.LocationID != nil {
		data["location_id"] = *filter.LocationID
		wc = append(wc, "location_id = :location_id")
	}

	if filter.SerialNumber != nil {
		data["serial_number"] = *filter.SerialNumber
		wc = append(wc, "serial_number = :serial_number")
	}

	if filter.Status != nil {
		data["status"] = *filter.Status
		wc = append(wc, "status = :status")
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
