package inspectionapp

import (
	"github.com/timmaaaz/ichor/business/domain/quality/inspectionbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("inspection_id", order.ASC)

var orderByFields = map[string]string{
	"inspection_id":        inspectionbus.OrderByInspectionID,
	"product_id":           inspectionbus.OrderByProductID,
	"inspector_id":         inspectionbus.OrderByInspectorID,
	"lot_id":               inspectionbus.OrderByLotID,
	"status":               inspectionbus.OrderByStatus,
	"notes":                inspectionbus.OrderByNotes,
	"inspection_date":      inspectionbus.OrderByInspectionDate,
	"next_inspection_date": inspectionbus.OrderByNextInspectionDate,
	"updated_date":         inspectionbus.OrderByUpdatedDate,
	"created_date":         inspectionbus.OrderByCreatedDate,
}
