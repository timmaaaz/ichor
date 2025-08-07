package inventorytransactiondb

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/movement/inventorytransactionbus"
)

type inventoryTransaction struct {
	InventoryTransactionID uuid.UUID `db:"id"`
	ProductID              uuid.UUID `db:"product_id"`
	LocationID             uuid.UUID `db:"location_id"`
	UserID                 uuid.UUID `db:"user_id"`
	Quantity               int       `db:"quantity"`
	TransactionType        string    `db:"transaction_type"`
	ReferenceNumber        string    `db:"reference_number"`
	TransactionDate        time.Time `db:"transaction_date"`
	CreatedDate            time.Time `db:"created_date"`
	UpdatedDate            time.Time `db:"updated_date"`
}

func toBusInventoryTransaction(db inventoryTransaction) inventorytransactionbus.InventoryTransaction {
	return inventorytransactionbus.InventoryTransaction{
		InventoryTransactionID: db.InventoryTransactionID,
		ProductID:              db.ProductID,
		LocationID:             db.LocationID,
		UserID:                 db.UserID,
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
	return inventoryTransaction{
		InventoryTransactionID: bus.InventoryTransactionID,
		ProductID:              bus.ProductID,
		LocationID:             bus.LocationID,
		UserID:                 bus.UserID,
		Quantity:               bus.Quantity,
		TransactionType:        bus.TransactionType,
		ReferenceNumber:        bus.ReferenceNumber,
		TransactionDate:        bus.TransactionDate,
		CreatedDate:            bus.CreatedDate,
		UpdatedDate:            bus.UpdatedDate,
	}
}
