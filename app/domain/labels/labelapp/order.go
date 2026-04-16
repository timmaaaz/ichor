package labelapp

import (
	"github.com/timmaaaz/ichor/business/domain/labels/labelbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy(labelbus.OrderByCode, order.ASC)

var orderByFields = map[string]string{
	"id":           labelbus.OrderByID,
	"code":         labelbus.OrderByCode,
	"type":         labelbus.OrderByType,
	"created_date": labelbus.OrderByCreatedDate,
}
