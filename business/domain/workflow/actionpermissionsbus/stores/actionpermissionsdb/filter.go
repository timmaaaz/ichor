package actionpermissionsdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/workflow/actionpermissionsbus"
)

func applyFilter(filter actionpermissionsbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["id"] = *filter.ID
		wc = append(wc, "id = :id")
	}

	if filter.RoleID != nil {
		data["role_id"] = *filter.RoleID
		wc = append(wc, "role_id = :role_id")
	}

	if filter.ActionType != nil {
		data["action_type"] = *filter.ActionType
		wc = append(wc, "action_type = :action_type")
	}

	if filter.IsAllowed != nil {
		data["is_allowed"] = *filter.IsAllowed
		wc = append(wc, "is_allowed = :is_allowed")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
