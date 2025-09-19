package approvalapp

import (
	"github.com/timmaaaz/ichor/business/domain/hr/approvalbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("name", order.ASC)

var orderByFields = map[string]string{
	"id":      approvalbus.OrderByID,
	"icon_id": approvalbus.OrderByIconID,
	"name":    approvalbus.OrderByName,
}
