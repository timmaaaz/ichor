package orgunitcolumnaccessdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/permissions/orgunitcolumnaccessbus"
)

func applyFilter(filter orgunitcolumnaccessbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["org_unit_column_access_id"] = *filter.ID
		buf.WriteString(" WHERE org_unit_column_access_id = :org_unit_column_access_id")
	}

	if filter.OrganizationalUnitID != nil {
		data["organizational_unit_id"] = *filter.OrganizationalUnitID
		buf.WriteString(" WHERE organizational_unit_id = :organizational_unit_id")
	}

	if filter.TableName != nil {
		data["table_name"] = *filter.TableName
		buf.WriteString(" WHERE table_name = :table_name")
	}

	if filter.ColumnName != nil {
		data["column_name"] = *filter.ColumnName
		buf.WriteString(" WHERE column_name = :column_name")
	}

	if filter.CanRead != nil {
		data["can_read"] = *filter.CanRead
		buf.WriteString(" WHERE can_read = :can_read")
	}

	if filter.CanUpdate != nil {
		data["can_update"] = *filter.CanUpdate
		buf.WriteString(" WHERE can_update = :can_update")
	}

	if filter.CanInheritPermissions != nil {
		data["can_inherit_permissions"] = *filter.CanInheritPermissions
		buf.WriteString(" WHERE can_inherit_permissions = :can_inherit_permissions")
	}

	if filter.CanRollupData != nil {
		data["can_rollup_data"] = *filter.CanRollupData
		buf.WriteString(" WHERE can_rollup_data = :can_rollup_data")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
