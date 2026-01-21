package currencyapp

import (
	"github.com/timmaaaz/ichor/business/domain/core/currencybus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy(currencybus.OrderBySortOrder, order.ASC)

var orderByFields = map[string]string{
	"id":             currencybus.OrderByID,
	"code":           currencybus.OrderByCode,
	"name":           currencybus.OrderByName,
	"sort_order":     currencybus.OrderBySortOrder,
	"is_active":      currencybus.OrderByIsActive,
	"decimal_places": currencybus.OrderByDecimalPlaces,
}
