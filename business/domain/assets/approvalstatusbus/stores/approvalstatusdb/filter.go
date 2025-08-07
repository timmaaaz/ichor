package approvalstatusdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/assets/approvalstatusbus"
)

func applyFilter(filter approvalstatusbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["id"] = *filter.ID
		wc = append(wc, "id = :id")
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
