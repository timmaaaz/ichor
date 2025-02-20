package organizationalunitdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/permissions/organizationalunitbus"
)

func applyFilter(filter organizationalunitbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["organizational_unit_id"] = *filter.ID
		wc = append(wc, "organizational_unit_id = :organizational_unit_id")
	}

	if filter.Name != nil {
		data["name"] = "%" + *filter.Name + "%"
		wc = append(wc, "name ILIKE :name")
	}

	if filter.Level != nil {
		data["level"] = *filter.Level
		wc = append(wc, "level = :level")
	}

	if filter.Path != nil {
		data["path"] = "%" + *filter.Path + "%"
		wc = append(wc, "path ILIKE :path")
	}

	if filter.CanInheritPermissions != nil {
		data["can_inherit_permissions"] = *filter.CanInheritPermissions
		wc = append(wc, "can_inherit_permissions = :can_inherit_permissions")
	}

	if filter.CanRollupData != nil {
		data["can_rollup_data"] = *filter.CanRollupData
		wc = append(wc, "can_rollup_data = :can_rollup_data")
	}

	if filter.UnitType != nil {
		data["unit_type"] = *filter.UnitType
		wc = append(wc, "unit_type = :unit_type")
	}

	if filter.IsActive != nil {
		data["is_active"] = *filter.IsActive
		wc = append(wc, "is_active = :is_active")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
