package actionpermissionsapp

import (
	"github.com/timmaaaz/ichor/business/domain/workflow/actionpermissionsbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy(actionpermissionsbus.OrderByActionType, order.ASC)

var orderByFields = map[string]string{
	"id":         actionpermissionsbus.OrderByID,
	"roleId":     actionpermissionsbus.OrderByRoleID,
	"actionType": actionpermissionsbus.OrderByActionType,
	"createdAt":  actionpermissionsbus.OrderByCreatedAt,
}
