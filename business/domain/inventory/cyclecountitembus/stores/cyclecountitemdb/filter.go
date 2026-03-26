package cyclecountitemdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountitembus"
)

func applyFilter(filter cyclecountitembus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["id"] = *filter.ID
		wc = append(wc, "id = :id")
	}

	if filter.SessionID != nil {
		data["session_id"] = *filter.SessionID
		wc = append(wc, "session_id = :session_id")
	}

	if filter.ProductID != nil {
		data["product_id"] = *filter.ProductID
		wc = append(wc, "product_id = :product_id")
	}

	if filter.LocationID != nil {
		data["location_id"] = *filter.LocationID
		wc = append(wc, "location_id = :location_id")
	}

	if filter.Status != nil {
		data["status"] = filter.Status.String()
		wc = append(wc, "status = :status")
	}

	if filter.CountedBy != nil {
		data["counted_by"] = *filter.CountedBy
		wc = append(wc, "counted_by = :counted_by")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
