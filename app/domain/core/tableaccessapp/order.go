package tableaccessapp

import (
	"github.com/timmaaaz/ichor/business/domain/core/tableaccessbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy(tableaccessbus.OrderByID, order.ASC)

var orderByFields = map[string]string{
	"id":         tableaccessbus.OrderByID,
	"role_id":    tableaccessbus.OrderByRoleID,
	"table_name": tableaccessbus.OrderByTableName,
	"can_create": tableaccessbus.OrderByCanCreate,
	"can_read":   tableaccessbus.OrderByCanRead,
	"can_update": tableaccessbus.OrderByCanUpdate,
	"can_delete": tableaccessbus.OrderByCanDelete,
}
