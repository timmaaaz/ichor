package assettypedb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/assets/assettypebus"
)

func applyFilter(filter assettypebus.QueryFilter, data map[string]interface{}, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["asset_type_id"] = *filter.ID
		wc = append(wc, "asset_type_id = :asset_type_id")
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
