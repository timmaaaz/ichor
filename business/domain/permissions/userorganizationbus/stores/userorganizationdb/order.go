package userorganizationdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/permissions/userorganizationbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	userorganizationbus.OrderByID:                   "user_organization_id",
	userorganizationbus.OrderByUserID:               "user_id",
	userorganizationbus.OrderByOrganizationalUnitID: "organizational_unit_id",
	userorganizationbus.OrderByRoleID:               "role_id",
	userorganizationbus.OrderByIsUnitManager:        "is_unit_manager",
	userorganizationbus.OrderByStartDate:            "start_date",
	userorganizationbus.OrderByEndDate:              "end_date",
	userorganizationbus.OrderByCreatedBy:            "created_by",
	userorganizationbus.OrderByCreatedAt:            "created_at",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
