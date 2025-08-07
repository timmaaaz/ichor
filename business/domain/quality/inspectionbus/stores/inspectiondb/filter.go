package inspectiondb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/quality/inspectionbus"
)

func applyFilter(filter inspectionbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.InspectionID != nil {
		data["id"] = *filter.InspectionID
		wc = append(wc, "id = :id")
	}

	if filter.ProductID != nil {
		data["product_id"] = *filter.ProductID
		wc = append(wc, "product_id = :product_id")
	}

	if filter.InspectorID != nil {
		data["inspector_id"] = *filter.InspectorID
		wc = append(wc, "inspector_id = :inspector_id")
	}

	if filter.LotID != nil {
		data["lot_id"] = *filter.LotID
		wc = append(wc, "lot_id = :lot_id")
	}

	if filter.Status != nil {
		data["status"] = *filter.Status
		wc = append(wc, "status = :status")
	}

	if filter.Notes != nil {
		data["notes"] = *filter.Notes
		wc = append(wc, "notes = :notes")
	}

	if filter.InspectionDate != nil {
		data["inspection_date"] = *filter.InspectionDate
		wc = append(wc, "inspection_date = :inspection_date")
	}

	if filter.NextInspectionDate != nil {
		data["next_inspection_date"] = *filter.NextInspectionDate
		wc = append(wc, "next_inspection_date = :next_inspection_date")
	}

	if filter.UpdatedDate != nil {
		data["updated_date"] = *filter.UpdatedDate
		wc = append(wc, "updated_date = :updated_date")
	}

	if filter.CreatedDate != nil {
		data["created_date"] = *filter.CreatedDate
		wc = append(wc, "created_date = :created_date")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
