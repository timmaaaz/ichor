package inspectiondb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/quality/inspectionbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	inspectionbus.OrderByInspectionID:       "inspection_id",
	inspectionbus.OrderByProductID:          "product_id",
	inspectionbus.OrderByInspectorID:        "inspector_id",
	inspectionbus.OrderByLotID:              "lot_id",
	inspectionbus.OrderByStatus:             "inspection_date",
	inspectionbus.OrderByNotes:              "next_inspection_date",
	inspectionbus.OrderByInspectionDate:     "status",
	inspectionbus.OrderByNextInspectionDate: "notes",
	inspectionbus.OrderByUpdatedDate:        "created_date",
	inspectionbus.OrderByCreatedDate:        "updated_date",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
