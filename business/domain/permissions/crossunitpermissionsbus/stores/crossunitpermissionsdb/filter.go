package crossunitpermissionsdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/permissions/crossunitpermissionsbus"
)

func applyFilter(filter crossunitpermissionsbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["cross_unit_permission_id"] = *filter.ID
		buf.WriteString(" WHERE cross_unit_permission_id = :cross_unit_permission_id")
	}

	if filter.SourceUnitID != nil {
		data["source_unit_id"] = *filter.SourceUnitID
		buf.WriteString(" WHERE source_unit_id = :source_unit_id")
	}

	if filter.TargetUnitID != nil {
		data["target_unit_id"] = *filter.TargetUnitID
		buf.WriteString(" WHERE target_unit_id = :target_unit_id")
	}

	if filter.CanRead != nil {
		data["can_read"] = *filter.CanRead
		buf.WriteString(" WHERE can_read = :can_read")
	}

	if filter.CanUpdate != nil {
		data["can_update"] = *filter.CanUpdate
		buf.WriteString(" WHERE can_update = :can_update")
	}

	if filter.GrantedBy != nil {
		data["granted_by"] = *filter.GrantedBy
		buf.WriteString(" WHERE granted_by = :granted_by")
	}

	if filter.StartValidFrom != nil {
		data["valid_from"] = *filter.StartValidFrom
		buf.WriteString(" WHERE valid_from >= :valid_from")
	}

	if filter.EndValidFrom != nil {
		data["valid_from"] = *filter.EndValidFrom
		buf.WriteString(" WHERE valid_from <= :valid_from")
	}

	if filter.StartValidUntil != nil {
		data["valid_until"] = *filter.StartValidUntil
		buf.WriteString(" WHERE valid_until >= :valid_until")
	}

	if filter.Reason != nil {
		data["reason"] = *filter.Reason
		buf.WriteString(" WHERE reason = :reason")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
