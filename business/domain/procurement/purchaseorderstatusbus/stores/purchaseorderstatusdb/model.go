package purchaseorderstatusdb

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderstatusbus"
)

type purchaseOrderStatus struct {
	ID          uuid.UUID `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	SortOrder   int       `db:"sort_order"`
}

func toDBPurchaseOrderStatus(bus purchaseorderstatusbus.PurchaseOrderStatus) purchaseOrderStatus {
	return purchaseOrderStatus{
		ID:          bus.ID,
		Name:        bus.Name,
		Description: bus.Description,
		SortOrder:   bus.SortOrder,
	}
}

func toBusPurchaseOrderStatus(db purchaseOrderStatus) purchaseorderstatusbus.PurchaseOrderStatus {
	return purchaseorderstatusbus.PurchaseOrderStatus{
		ID:          db.ID,
		Name:        db.Name,
		Description: db.Description,
		SortOrder:   db.SortOrder,
	}
}

func toBusPurchaseOrderStatuses(dbs []purchaseOrderStatus) []purchaseorderstatusbus.PurchaseOrderStatus {
	statuses := make([]purchaseorderstatusbus.PurchaseOrderStatus, len(dbs))
	for i, db := range dbs {
		statuses[i] = toBusPurchaseOrderStatus(db)
	}
	return statuses
}
