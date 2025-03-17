package userroleapp

import (
	"github.com/timmaaaz/ichor/business/domain/permissions/userrolebus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy(userrolebus.OrderByUserID, order.ASC)

var orderByFields = map[string]string{
	"user_role_id": userrolebus.OrderByID,
	"user_id":      userrolebus.OrderByUserID,
	"role_id":      userrolebus.OrderByRoleID,
}
