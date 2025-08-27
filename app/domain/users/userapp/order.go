package userapp

import (
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("id", order.ASC)

var orderByFields = map[string]string{
	"id":             userbus.OrderByID,
	"requested_by":   userbus.OrderByRequestedBy,
	"approved_by":    userbus.OrderByApprovedBy,
	"title_id":       userbus.OrderByTitleID,
	"office_id":      userbus.OrderByOfficeID,
	"username":       userbus.OrderByUsername,
	"first_name":     userbus.OrderByFirstName,
	"last_name":      userbus.OrderByLastName,
	"email":          userbus.OrderByEmail,
	"roles":          userbus.OrderByRoles,
	"system_roles":   userbus.OrderBySystemRoles,
	"enabled":        userbus.OrderByEnabled,
	"birthday":       userbus.OrderByBirthday,
	"date_hired":     userbus.OrderByDateHired,
	"date_requested": userbus.OrderByDateRequested,
	"date_approved":  userbus.OrderByDateApproved,
	"created_date":   userbus.OrderByCreatedDate,
	"date_modified":  userbus.OrderByDateModified,
}
