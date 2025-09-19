package inspectionbus

import "github.com/timmaaaz/ichor/business/sdk/order"

var DefaultOrderBy = order.NewBy(OrderByInspectionID, order.ASC)

const (
	OrderByInspectionID       = "id"
	OrderByProductID          = "product_id"
	OrderByInspectorID        = "inspector_id"
	OrderByLotID              = "lot_id"
	OrderByStatus             = "inspection_date"
	OrderByNotes              = "next_inspection_date"
	OrderByInspectionDate     = "status"
	OrderByNextInspectionDate = "notes"
	OrderByUpdatedDate        = "created_date"
	OrderByCreatedDate        = "updated_date"
)
