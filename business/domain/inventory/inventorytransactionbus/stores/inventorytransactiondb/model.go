package inventorytransactiondb

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorytransactionbus"
)

type inventoryTransaction struct {
	InventoryTransactionID uuid.UUID      `db:"id"`
	ProductID              uuid.UUID      `db:"product_id"`
	LocationID             uuid.UUID      `db:"location_id"`
	UserID                 uuid.UUID      `db:"user_id"`
	LotID                  uuid.NullUUID  `db:"lot_id"`
	Quantity               int            `db:"quantity"`
	TransactionType        string         `db:"transaction_type"`
	ReferenceNumber        string         `db:"reference_number"`
	TransactionDate        time.Time      `db:"transaction_date"`
	CreatedDate            time.Time      `db:"created_date"`
	UpdatedDate            time.Time      `db:"updated_date"`
}

func toBusInventoryTransaction(db inventoryTransaction) inventorytransactionbus.InventoryTransaction {
	var lotID *uuid.UUID
	if db.LotID.Valid {
		id := db.LotID.UUID
		lotID = &id
	}

	return inventorytransactionbus.InventoryTransaction{
		InventoryTransactionID: db.InventoryTransactionID,
		ProductID:              db.ProductID,
		LocationID:             db.LocationID,
		UserID:                 db.UserID,
		LotID:                  lotID,
		Quantity:               db.Quantity,
		TransactionType:        db.TransactionType,
		ReferenceNumber:        db.ReferenceNumber,
		TransactionDate:        db.TransactionDate,
		CreatedDate:            db.CreatedDate,
		UpdatedDate:            db.UpdatedDate,
	}
}

func toBusInventoryTransactions(dbs []inventoryTransaction) []inventorytransactionbus.InventoryTransaction {
	bus := make([]inventorytransactionbus.InventoryTransaction, len(dbs))

	for i, db := range dbs {
		bus[i] = toBusInventoryTransaction(db)
	}
	return bus
}

func toDBInventoryTransaction(bus inventorytransactionbus.InventoryTransaction) inventoryTransaction {
	var lotID uuid.NullUUID
	if bus.LotID != nil {
		lotID = uuid.NullUUID{UUID: *bus.LotID, Valid: true}
	}

	return inventoryTransaction{
		InventoryTransactionID: bus.InventoryTransactionID,
		ProductID:              bus.ProductID,
		LocationID:             bus.LocationID,
		UserID:                 bus.UserID,
		LotID:                  lotID,
		Quantity:               bus.Quantity,
		TransactionType:        bus.TransactionType,
		ReferenceNumber:        bus.ReferenceNumber,
		TransactionDate:        bus.TransactionDate,
		CreatedDate:            bus.CreatedDate,
		UpdatedDate:            bus.UpdatedDate,
	}
}
