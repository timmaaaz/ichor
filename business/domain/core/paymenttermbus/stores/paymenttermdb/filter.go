package paymenttermdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/core/paymenttermbus"
)

func applyFilter(filter paymenttermbus.QueryFilter, data map[string]interface{}, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["id"] = *filter.ID
		wc = append(wc, "id = :id")
	}

	if filter.Name != nil {
		data["name"] = "%" + *filter.Name + "%"
		wc = append(wc, "name ILIKE :name")
	}

	if filter.Description != nil {
		data["description"] = "%" + *filter.Description + "%"
		wc = append(wc, "description ILIKE :description")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
