package inventoryitemdb

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
)

type inventoryItem struct {
	ItemID                uuid.UUID `db:"id"`
	ProductID             uuid.UUID `db:"product_id"`
	LocationID            uuid.UUID `db:"location_id"`
	Quantity              int       `db:"quantity"`
	ReservedQuantity      int       `db:"reserved_quantity"`
	AllocatedQuantity     int       `db:"allocated_quantity"`
	MinimumStock          int       `db:"minimum_stock"`
	MaximumStock          int       `db:"maximum_stock"`
	ReorderPoint          int       `db:"reorder_point"`
	EconomicOrderQuantity int       `db:"economic_order_quantity"`
	SafetyStock           int       `db:"safety_stock"`
	AvgDailyUsage         int       `db:"avg_daily_usage"`
	CreatedDate           time.Time `db:"created_date"`
	UpdatedDate           time.Time `db:"updated_date"`
}

func toBusInventoryItem(db inventoryItem) inventoryitembus.InventoryItem {
	return inventoryitembus.InventoryItem{
		ItemID:                db.ItemID,
		ProductID:             db.ProductID,
		LocationID:            db.LocationID,
		Quantity:              db.Quantity,
		ReservedQuantity:      db.ReservedQuantity,
		AllocatedQuantity:     db.AllocatedQuantity,
		MinimumStock:          db.MinimumStock,
		MaximumStock:          db.MaximumStock,
		ReorderPoint:          db.ReorderPoint,
		EconomicOrderQuantity: db.EconomicOrderQuantity,
		SafetyStock:           db.SafetyStock,
		AvgDailyUsage:         db.AvgDailyUsage,
		CreatedDate:           db.CreatedDate.UTC(),
		UpdatedDate:           db.UpdatedDate.UTC(),
	}
}

func toBusInventoryItems(dbs []inventoryItem) []inventoryitembus.InventoryItem {
	bus := make([]inventoryitembus.InventoryItem, len(dbs))

	for i, db := range dbs {
		bus[i] = toBusInventoryItem(db)
	}
	return bus
}

func toDBInventoryItem(bus inventoryitembus.InventoryItem) inventoryItem {
	return inventoryItem{
		ItemID:                bus.ItemID,
		ProductID:             bus.ProductID,
		LocationID:            bus.LocationID,
		Quantity:              bus.Quantity,
		ReservedQuantity:      bus.ReservedQuantity,
		AllocatedQuantity:     bus.AllocatedQuantity,
		MinimumStock:          bus.MinimumStock,
		MaximumStock:          bus.MaximumStock,
		ReorderPoint:          bus.ReorderPoint,
		EconomicOrderQuantity: bus.EconomicOrderQuantity,
		SafetyStock:           bus.SafetyStock,
		AvgDailyUsage:         bus.AvgDailyUsage,
		CreatedDate:           bus.CreatedDate.UTC(),
		UpdatedDate:           bus.UpdatedDate.UTC(),
	}
}
