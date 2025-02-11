package userapprovalstatusdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/userapprovalstatusbus"
)

func applyFilter(filter userapprovalstatusbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["user_approval_status_id"] = *filter.ID
		wc = append(wc, "approval_status_id = :approval_status_id")
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
