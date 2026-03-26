package cyclecountsessionapp

import (
	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountsessionbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("id", order.ASC)

var orderByFields = map[string]string{
	cyclecountsessionbus.OrderByID:          "id",
	cyclecountsessionbus.OrderByName:        "name",
	cyclecountsessionbus.OrderByStatus:      "status",
	cyclecountsessionbus.OrderByCreatedBy:   "created_by",
	cyclecountsessionbus.OrderByCreatedDate: "created_date",
}
