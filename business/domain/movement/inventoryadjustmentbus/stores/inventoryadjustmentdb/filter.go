package inventoryadjustmentdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/movement/inventoryadjustmentbus"
)

func applyFilter(filter inventoryadjustmentbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.InventoryAdjustmentID != nil {
		data["adjustment_id"] = *filter.InventoryAdjustmentID
		wc = append(wc, "adjustment_id = :adjustment_id")
	}

	if filter.ProductID != nil {
		data["product_id"] = *filter.ProductID
		wc = append(wc, "product_id = :product_id")
	}

	if filter.LocationID != nil {
		data["location_id"] = *filter.LocationID
		wc = append(wc, "location_id = :location_id")
	}

	if filter.AdjustedBy != nil {
		data["adjusted_by"] = *filter.AdjustedBy
		wc = append(wc, "adjusted_by = :adjusted_by")
	}

	if filter.ApprovedBy != nil {
		data["approved_by"] = *filter.ApprovedBy
		wc = append(wc, "approved_by = :approved_by")
	}

	if filter.QuantityChange != nil {
		data["quantity_change"] = *filter.QuantityChange
		wc = append(wc, "quantity_change = :quantity_change")
	}

	if filter.ReasonCode != nil {
		data["reason_code"] = *filter.ReasonCode
		wc = append(wc, "reason_code = :reason_code")
	}

	if filter.Notes != nil {
		data["notes"] = *filter.Notes
		wc = append(wc, "notes = :notes")
	}

	if filter.AdjustmentDate != nil {
		data["adjustment_date"] = *filter.AdjustmentDate
		wc = append(wc, "adjustment_date = :adjustment_date")
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
