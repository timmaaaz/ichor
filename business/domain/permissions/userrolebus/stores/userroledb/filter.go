package userroledb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/permissions/userrolebus"
)

func applyFilter(filter userrolebus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["user_role_id"] = *filter.ID
		wc = append(wc, "user_role_id = :user_role_id")
	}

	if filter.UserID != nil {
		data["user_id"] = *filter.UserID
		wc = append(wc, "user_id = :user_id")
	}

	if filter.RoleID != nil {
		data["role_id"] = *filter.RoleID
		wc = append(wc, "role_id = :role_id")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
