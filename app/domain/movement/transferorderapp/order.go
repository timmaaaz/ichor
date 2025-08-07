package transferorderapp

import (
	"github.com/timmaaaz/ichor/business/domain/movement/transferorderbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("id", order.ASC)

var orderByFields = map[string]string{
	transferorderbus.OrderByTransferID:     "id",
	transferorderbus.OrderByProductID:      "product_id",
	transferorderbus.OrderByFromLocationID: "from_location_id",
	transferorderbus.OrderByToLocationID:   "to_location_id",
	transferorderbus.OrderByRequestedByID:  "requested_by_id",
	transferorderbus.OrderByApprovedByID:   "approved_by_id",
	transferorderbus.OrderByQuantity:       "quantity",
	transferorderbus.OrderByStatus:         "status",
	transferorderbus.OrderByTransferDate:   "transfer_date",
	transferorderbus.OrderByCreatedDate:    "created_date",
	transferorderbus.OrderByUpdatedDate:    "updated_date",
}
