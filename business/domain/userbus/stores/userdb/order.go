package userdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/userbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	userbus.OrderByID:                 "user_id",
	userbus.OrderByRequestedBy:        "requested_by",
	userbus.OrderByApprovedBy:         "approved_by",
	userbus.OrderByUserApprovalStatus: "user_approval_status",
	userbus.OrderByTitleID:            "title_id",
	userbus.OrderByOfficeID:           "office_id",
	userbus.OrderByUsername:           "username",
	userbus.OrderByFirstName:          "first_name",
	userbus.OrderByLastName:           "last_name",
	userbus.OrderByEmail:              "email",
	userbus.OrderByRoles:              "roles",
	userbus.OrderBySystemRoles:        "system_roles",
	userbus.OrderByEnabled:            "enabled",
	userbus.OrderByBirthday:           "birthday",
	userbus.OrderByDateHired:          "date_hired",
	userbus.OrderByDateRequested:      "date_requested",
	userbus.OrderByDateApproved:       "date_approved",
	userbus.OrderByDateCreated:        "date_created",
	userbus.OrderByDateModified:       "date_modified",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
