package approvalstatusapp

import (
	"github.com/timmaaaz/ichor/business/domain/approvalstatusbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("name", order.ASC)

var orderByFields = map[string]string{
	"approval_status_id": approvalstatusbus.OrderByID,
	"icon_id":            approvalstatusbus.OrderByIconID,
	"name":               approvalstatusbus.OrderByName,
}
