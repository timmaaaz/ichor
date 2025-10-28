package rolepageapp

import (
	"github.com/timmaaaz/ichor/business/domain/core/rolepagebus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy(rolepagebus.OrderByID, order.ASC)

var orderByFields = map[string]string{
	"id":        rolepagebus.OrderByID,
	"roleId":    rolepagebus.OrderByRoleID,
	"pageId":    rolepagebus.OrderByPageID,
	"canAccess": rolepagebus.OrderByCanAccess,
}
