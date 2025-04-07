package inventoryadjustmentapp

import (
	"github.com/timmaaaz/ichor/business/domain/movement/inventoryadjustmentbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("adjustment_id", order.ASC)

var orderByFields = map[string]string{
	"adjustment_id":   inventoryadjustmentbus.OrderByInventoryAdjustmentID,
	"product_id":      inventoryadjustmentbus.OrderByProductID,
	"location_id":     inventoryadjustmentbus.OrderByLocationID,
	"adjusted_by":     inventoryadjustmentbus.OrderByAdjustedBy,
	"approved_by":     inventoryadjustmentbus.OrderByApprovedBy,
	"quantity_change": inventoryadjustmentbus.OrderByQuantityChange,
	"reason_code":     inventoryadjustmentbus.OrderByReasonCode,
	"notes":           inventoryadjustmentbus.OrderByNotes,
	"adjustment_date": inventoryadjustmentbus.OrderByAdjustmentDate,
	"created_date":    inventoryadjustmentbus.OrderByCreatedDate,
	"updated_date":    inventoryadjustmentbus.OrderByUpdatedDate,
}
