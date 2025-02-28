package crossunitpermissionsdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/permissions/crossunitpermissionsbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	crossunitpermissionsbus.OrderByID:           "cross_unit_permission_id",
	crossunitpermissionsbus.OrderBySourceUnitID: "source_unit_id",
	crossunitpermissionsbus.OrderByTargetUnitID: "target_unit_id",
	crossunitpermissionsbus.OrderByCanRead:      "can_read",
	crossunitpermissionsbus.OrderByCanUpdate:    "can_update",
	crossunitpermissionsbus.OrderByGrantedBy:    "granted_by",
	crossunitpermissionsbus.OrderByValidFrom:    "valid_from",
	crossunitpermissionsbus.OrderByValidUntil:   "valid_until",
	crossunitpermissionsbus.OrderByReason:       "reason",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}
	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
