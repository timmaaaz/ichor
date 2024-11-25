package fulfillmentstatusdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/fulfillmentstatusbus"
)

func applyFilter(filter fulfillmentstatusbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["fulfillment_status_id"] = *filter.ID
		wc = append(wc, "fulfillment_status_id = :fulfillment_status_id")
	}

	if filter.IconID != nil {
		data["icon_id"] = *filter.IconID
		wc = append(wc, "icon_id = :icon_id")
	}

	if filter.Name != nil {
		data["name"] = "%" + *filter.Name + "%"
		wc = append(wc, "name LIKE :name")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
