package purchaseorderlineitemstatusdb

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitemstatusbus"
)

type purchaseOrderLineItemStatus struct {
	ID          uuid.UUID `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	SortOrder   int       `db:"sort_order"`
}

func toDBPurchaseOrderLineItemStatus(bus purchaseorderlineitemstatusbus.PurchaseOrderLineItemStatus) purchaseOrderLineItemStatus {
	return purchaseOrderLineItemStatus{
		ID:          bus.ID,
		Name:        bus.Name,
		Description: bus.Description,
		SortOrder:   bus.SortOrder,
	}
}

func toBusPurchaseOrderLineItemStatus(db purchaseOrderLineItemStatus) purchaseorderlineitemstatusbus.PurchaseOrderLineItemStatus {
	return purchaseorderlineitemstatusbus.PurchaseOrderLineItemStatus{
		ID:          db.ID,
		Name:        db.Name,
		Description: db.Description,
		SortOrder:   db.SortOrder,
	}
}

func toBusPurchaseOrderLineItemStatuses(dbs []purchaseOrderLineItemStatus) []purchaseorderlineitemstatusbus.PurchaseOrderLineItemStatus {
	statuses := make([]purchaseorderlineitemstatusbus.PurchaseOrderLineItemStatus, len(dbs))
	for i, db := range dbs {
		statuses[i] = toBusPurchaseOrderLineItemStatus(db)
	}
	return statuses
}