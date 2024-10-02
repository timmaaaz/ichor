package assetconditiondb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/assetconditionbus"
)

func applyFilter(filter assetconditionbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["asset_condition_id"] = *filter.ID
		wc = append(wc, "asset_condition_id = :asset_condition_id")
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