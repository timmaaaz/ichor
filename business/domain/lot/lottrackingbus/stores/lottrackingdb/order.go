package lottrackingdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/lot/lottrackingbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	lottrackingbus.OrderByLotID:             "id",
	lottrackingbus.OrderBySupplierProductID: "supplier_product_id",
	lottrackingbus.OrderByLotNumber:         "lot_number",
	lottrackingbus.OrderByManufactureDate:   "manufacture_date",
	lottrackingbus.OrderByExpirationDate:    "expiration_date",
	lottrackingbus.OrderByRecievedDate:      "received_date",
	lottrackingbus.OrderByQuantity:          "quantity",
	lottrackingbus.OrderByQualityStatus:     "quality_status",
	lottrackingbus.OrderByCreatedDate:       "created_date",
	lottrackingbus.OrderByUpdatedDate:       "updated_date",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
