package pageactionapp

import (
	"github.com/timmaaaz/ichor/business/domain/config/pageactionbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy(pageactionbus.OrderByActionOrder, order.ASC)

// Set of fields that the results can be ordered by.
const (
	OrderByID           = "id"
	OrderByPageConfigID = "pageConfigId"
	OrderByActionType   = "actionType"
	OrderByActionOrder  = "actionOrder"
)

var orderByFields = map[string]string{
	OrderByID:           pageactionbus.OrderByID,
	OrderByPageConfigID: pageactionbus.OrderByPageConfigID,
	OrderByActionType:   pageactionbus.OrderByActionType,
	OrderByActionOrder:  pageactionbus.OrderByActionOrder,
}
