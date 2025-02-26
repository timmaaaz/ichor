package userorganizationdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/permissions/userorganizationbus"
)

func applyFilter(filter userorganizationbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["user_organization_id"] = *filter.ID
		buf.WriteString(" WHERE user_organization_id = :user_organization_id")
	}

	if filter.UserID != nil {
		data["user_id"] = *filter.UserID
		buf.WriteString(" WHERE user_id = :user_id")
	}

	if filter.OrganizationalUnitID != nil {
		data["organizational_unit_id"] = *filter.OrganizationalUnitID
		buf.WriteString(" WHERE organizational_unit_id = :organizational_unit_id")
	}

	if filter.RoleID != nil {
		data["role_id"] = *filter.RoleID
		buf.WriteString(" WHERE role_id = :role_id")
	}

	if filter.IsUnitManager != nil {
		data["is_unit_manager"] = *filter.IsUnitManager
		buf.WriteString(" WHERE is_unit_manager = :is_unit_manager")
	}

	if filter.StartDate != nil {
		data["start_date"] = *filter.StartDate
		buf.WriteString(" WHERE start_date = :start_date")
	}

	if filter.EndDate != nil {
		data["end_date"] = *filter.EndDate
		buf.WriteString(" WHERE end_date = :end_date")
	}

	if filter.CreatedBy != nil {
		data["created_by"] = *filter.CreatedBy
		buf.WriteString(" WHERE created_by = :created_by")
	}

	if filter.CreatedAt != nil {
		data["created_at"] = *filter.CreatedAt
		buf.WriteString(" WHERE created_at = :created_at")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
