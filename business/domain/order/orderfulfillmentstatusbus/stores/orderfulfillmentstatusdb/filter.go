package orderfulfillmentstatusdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/order/orderfulfillmentstatusbus"
)

// TODO: Switch these over to use string.builder?

func applyFilter(filter orderfulfillmentstatusbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["id"] = *filter.ID
		wc = append(wc, "id = :id")
	}

	if filter.Name != nil {
		data["name"] = *filter.Name
		wc = append(wc, "name = :name")
	}

	if filter.Description != nil {
		data["description"] = *filter.Description
		wc = append(wc, "description = :description")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
