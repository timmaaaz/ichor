package userorganizationbus

import "github.com/timmaaaz/ichor/business/sdk/order"

// DefaultOrderBy represents the default way we sort.
var DefaultOrderBy = order.NewBy(OrderByUserID, order.ASC)

// Set of fields that the results can be ordered by.
const (
	OrderByID             = "user_organization_id"
	OrderByUserID         = "user_id"
	OrderByOrganizationID = "organization_id"
	OrderByRoleID         = "role_id"
	OrderByIsUnitManager  = "is_unit_manager"
	OrderByStartDate      = "start_date"
	OrderByEndDate        = "end_date"
	OrderByCreatedBy      = "created_by"
	OrderByCreatedAt      = "created_at"
)
