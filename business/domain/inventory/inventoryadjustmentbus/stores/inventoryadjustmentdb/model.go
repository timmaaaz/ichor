package inventoryadjustmentdb

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryadjustmentbus"
)

type inventoryAdjustment struct {
	InventoryAdjustmentID uuid.UUID `db:"id"`
	ProductID             uuid.UUID `db:"product_id"`
	LocationID            uuid.UUID `db:"location_id"`
	AdjustedBy            uuid.UUID `db:"adjusted_by"`
	ApprovedBy            uuid.UUID `db:"approved_by"`
	QuantityChange        int       `db:"quantity_change"`
	ReasonCode            string    `db:"reason_code"`
	Notes                 string    `db:"notes"`
	AdjustmentDate        time.Time `db:"adjustment_date"`
	CreatedDate           time.Time `db:"created_date"`
	UpdatedDate           time.Time `db:"updated_date"`
}

func toBusInventoryAdjustment(ia inventoryAdjustment) inventoryadjustmentbus.InventoryAdjustment {
	return inventoryadjustmentbus.InventoryAdjustment{
		InventoryAdjustmentID: ia.InventoryAdjustmentID,
		ProductID:             ia.ProductID,
		LocationID:            ia.LocationID,
		AdjustedBy:            ia.AdjustedBy,
		ApprovedBy:            ia.ApprovedBy,
		QuantityChange:        ia.QuantityChange,
		ReasonCode:            ia.ReasonCode,
		Notes:                 ia.Notes,
		AdjustmentDate:        ia.AdjustmentDate,
		CreatedDate:           ia.CreatedDate,
		UpdatedDate:           ia.UpdatedDate,
	}
}

func toBusInventoryAdjustments(ias []inventoryAdjustment) []inventoryadjustmentbus.InventoryAdjustment {
	bus := make([]inventoryadjustmentbus.InventoryAdjustment, len(ias))
	for i, ia := range ias {
		bus[i] = toBusInventoryAdjustment(ia)
	}
	return bus
}

func toDBInventoryAdjustment(ia inventoryadjustmentbus.InventoryAdjustment) inventoryAdjustment {
	return inventoryAdjustment{
		InventoryAdjustmentID: ia.InventoryAdjustmentID,
		ProductID:             ia.ProductID,
		LocationID:            ia.LocationID,
		AdjustedBy:            ia.AdjustedBy,
		ApprovedBy:            ia.ApprovedBy,
		QuantityChange:        ia.QuantityChange,
		ReasonCode:            ia.ReasonCode,
		Notes:                 ia.Notes,
		AdjustmentDate:        ia.AdjustmentDate,
		CreatedDate:           ia.CreatedDate,
		UpdatedDate:           ia.UpdatedDate,
	}
}
