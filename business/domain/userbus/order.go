package userbus

import "bitbucket.org/superiortechnologies/ichor/business/sdk/order"

// DefaultOrderBy represents the default way we sort.
var DefaultOrderBy = order.NewBy(OrderByID, order.ASC)

// Set of fields that the results can be ordered by.
const (
	OrderByID            = "user_id"
	OrderByRequestedBy   = "requested_by"
	OrderByApprovedBy    = "approved_by"
	OrderByTitleID       = "title_id"
	OrderByOfficeID      = "office_id"
	OrderByUsername      = "username"
	OrderByFirstName     = "first_name"
	OrderByLastName      = "last_name"
	OrderByEmail         = "email"
	OrderByRoles         = "roles"
	OrderBySystemRoles   = "system_roles"
	OrderByEnabled       = "enabled"
	OrderByBirthday      = "birthday"
	OrderByDateHired     = "date_hired"
	OrderByDateRequested = "date_requested"
	OrderByDateApproved  = "date_approved"
	OrderByDateCreated   = "date_created"
	OrderByDateModified  = "date_modified"
)
