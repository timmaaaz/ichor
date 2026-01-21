package costhistoryapp

import (
	"github.com/timmaaaz/ichor/business/domain/products/costhistorybus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("amount", order.ASC)

var orderByFields = map[string]string{
	"id":             costhistorybus.OrderByCostHistoryID,
	"product_id":     costhistorybus.OrderByProductID,
	"cost_type":      costhistorybus.OrderByCostType,
	"amount":         costhistorybus.OrderByAmount,
	"currency_id":    costhistorybus.OrderByCurrencyID,
	"effective_date": costhistorybus.OrderByEffectiveDate,
	"end_date":       costhistorybus.OrderByEndDate,
	"created_date":   costhistorybus.OrderByCreatedDate,
	"updated_date":   costhistorybus.OrderByUpdatedDate,
}
