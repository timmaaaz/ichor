package tableaccessdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/permissions/tableaccessbus"
)

func applyFilter(filter tableaccessbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["table_access_id"] = *filter.ID
		buf.WriteString(" WHERE table_access_id = :table_access_id")
	}

	if filter.RoleID != nil {
		data["role_id"] = *filter.RoleID
		buf.WriteString(" WHERE role_id = :role_id")
	}

	if filter.TableName != nil {
		data["table_name"] = *filter.TableName
		buf.WriteString(" WHERE table_name = :table_name")
	}

	if filter.CanCreate != nil {
		data["can_create"] = *filter.CanCreate
		buf.WriteString(" WHERE can_create = :can_create")
	}

	if filter.CanRead != nil {
		data["can_read"] = *filter.CanRead
		buf.WriteString(" WHERE can_read = :can_read")
	}

	if filter.CanUpdate != nil {
		data["can_update"] = *filter.CanUpdate
		buf.WriteString(" WHERE can_update = :can_update")
	}

	if filter.CanDelete != nil {
		data["can_delete"] = *filter.CanDelete
		buf.WriteString(" WHERE can_delete = :can_delete")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
