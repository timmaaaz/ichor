package crossunitpermissionsbus

import "github.com/timmaaaz/ichor/business/sdk/order"

// DefaultOrderBy represents the default way we sort.
var DefaultOrderBy = order.NewBy(OrderByID, order.ASC)

// Set of fields that the results can be ordered by.
const (
	OrderByID           = "cross_unit_permission_id"
	OrderBySourceUnitID = "source_unit_id"
	OrderByTargetUnitID = "target_unit_id"
	OrderByCanRead      = "can_read"
	OrderByCanUpdate    = "can_update"
	OrderByGrantedBy    = "granted_by"
	OrderByValidFrom    = "valid_from"
	OrderByValidUntil   = "valid_until"
	OrderByReason       = "reason"
)
