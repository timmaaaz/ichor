package inventoryitemdb

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
)

type inventoryItem struct {
	ID                    uuid.UUID `db:"id"`
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
		ID:                    db.ID,
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

type inventoryItemWithLocation struct {
	ID                    uuid.UUID `db:"id"`
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
	LocationCode          string    `db:"location_code"`
	Aisle                 string    `db:"aisle"`
	Rack                  string    `db:"rack"`
	Shelf                 string    `db:"shelf"`
	Bin                   string    `db:"bin"`
	ZoneName              string    `db:"zone_name"`
	WarehouseName         string    `db:"warehouse_name"`
}

func toBusInventoryItemWithLocation(db inventoryItemWithLocation) inventoryitembus.InventoryItemWithLocation {
	return inventoryitembus.InventoryItemWithLocation{
		InventoryItem: inventoryitembus.InventoryItem{
			ID:                    db.ID,
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
		},
		LocationCode:  db.LocationCode,
		Aisle:         db.Aisle,
		Rack:          db.Rack,
		Shelf:         db.Shelf,
		Bin:           db.Bin,
		ZoneName:      db.ZoneName,
		WarehouseName: db.WarehouseName,
	}
}

func toBusInventoryItemsWithLocation(dbs []inventoryItemWithLocation) []inventoryitembus.InventoryItemWithLocation {
	bus := make([]inventoryitembus.InventoryItemWithLocation, len(dbs))
	for i, db := range dbs {
		bus[i] = toBusInventoryItemWithLocation(db)
	}
	return bus
}

type itemWithProduct struct {
	ProductID    uuid.UUID `db:"product_id"`
	ProductName  string    `db:"product_name"`
	ProductSKU   string    `db:"product_sku"`
	TrackingType string    `db:"tracking_type"`
	Quantity     int       `db:"quantity"`
}

func toBusItemWithProduct(db itemWithProduct) inventoryitembus.ItemWithProduct {
	return inventoryitembus.ItemWithProduct{
		ProductID:    db.ProductID,
		ProductName:  db.ProductName,
		ProductSKU:   db.ProductSKU,
		TrackingType: db.TrackingType,
		Quantity:     db.Quantity,
	}
}

func toBusItemsWithProduct(dbs []itemWithProduct) []inventoryitembus.ItemWithProduct {
	bus := make([]inventoryitembus.ItemWithProduct, len(dbs))
	for i, db := range dbs {
		bus[i] = toBusItemWithProduct(db)
	}
	return bus
}

func toDBInventoryItem(bus inventoryitembus.InventoryItem) inventoryItem {
	return inventoryItem{
		ID:                    bus.ID,
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
