package cyclecountitemapp

import (
	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountitembus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("id", order.ASC)

var orderByFields = map[string]string{
	cyclecountitembus.OrderByID:             "id",
	cyclecountitembus.OrderBySessionID:      "session_id",
	cyclecountitembus.OrderByProductID:      "product_id",
	cyclecountitembus.OrderByLocationID:     "location_id",
	cyclecountitembus.OrderBySystemQuantity: "system_quantity",
	cyclecountitembus.OrderByStatus:         "status",
	cyclecountitembus.OrderByCreatedDate:    "created_date",
}
