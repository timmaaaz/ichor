package rolepagedb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/core/rolepagebus"
)

func applyFilter(filter rolepagebus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["id"] = *filter.ID
		wc = append(wc, "id = :id")
	}

	if filter.RoleID != nil {
		data["role_id"] = *filter.RoleID
		wc = append(wc, "role_id = :role_id")
	}

	if filter.PageID != nil {
		data["page_id"] = *filter.PageID
		wc = append(wc, "page_id = :page_id")
	}

	if filter.CanAccess != nil {
		data["can_access"] = *filter.CanAccess
		wc = append(wc, "can_access = :can_access")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
