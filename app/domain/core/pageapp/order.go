package pageapp

import (
	"github.com/timmaaaz/ichor/business/domain/core/pagebus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy(pagebus.OrderBySortOrder, order.ASC)

var orderByFields = map[string]string{
	"id":        pagebus.OrderByID,
	"path":      pagebus.OrderByPath,
	"name":      pagebus.OrderByName,
	"module":    pagebus.OrderByModule,
	"sortOrder": pagebus.OrderBySortOrder,
	"isActive":  pagebus.OrderByIsActive,
}
