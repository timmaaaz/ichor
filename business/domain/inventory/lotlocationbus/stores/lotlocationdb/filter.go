package lotlocationdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/inventory/lotlocationbus"
)

func applyFilter(filter lotlocationbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["id"] = *filter.ID
		wc = append(wc, "id = :id")
	}

	if filter.LotID != nil {
		data["lot_id"] = *filter.LotID
		wc = append(wc, "lot_id = :lot_id")
	}

	if filter.LocationID != nil {
		data["location_id"] = *filter.LocationID
		wc = append(wc, "location_id = :location_id")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
