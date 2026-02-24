package putawaytaskapp

import (
	"github.com/timmaaaz/ichor/business/domain/inventory/putawaytaskbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("id", order.ASC)

var orderByFields = map[string]string{
	putawaytaskbus.OrderByID:              "id",
	putawaytaskbus.OrderByProductID:       "product_id",
	putawaytaskbus.OrderByLocationID:      "location_id",
	putawaytaskbus.OrderByQuantity:        "quantity",
	putawaytaskbus.OrderByReferenceNumber: "reference_number",
	putawaytaskbus.OrderByStatus:          "status",
	putawaytaskbus.OrderByAssignedTo:      "assigned_to",
	putawaytaskbus.OrderByCreatedBy:       "created_by",
	putawaytaskbus.OrderByCreatedDate:     "created_date",
	putawaytaskbus.OrderByUpdatedDate:     "updated_date",
}
