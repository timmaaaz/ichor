package userapprovalstatusapp

import (
	"github.com/timmaaaz/ichor/business/domain/userapprovalstatusbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("name", order.ASC)

var orderByFields = map[string]string{
	"user_approval_status_id": userapprovalstatusbus.OrderByID,
	"icon_id":                 userapprovalstatusbus.OrderByIconID,
	"name":                    userapprovalstatusbus.OrderByName,
}
