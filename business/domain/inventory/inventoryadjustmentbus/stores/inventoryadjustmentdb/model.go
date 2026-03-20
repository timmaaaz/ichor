package inventoryadjustmentdb

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryadjustmentbus"
)

type inventoryAdjustment struct {
	InventoryAdjustmentID uuid.UUID  `db:"id"`
	ProductID             uuid.UUID  `db:"product_id"`
	LocationID            uuid.UUID  `db:"location_id"`
	AdjustedBy            uuid.UUID  `db:"adjusted_by"`
	ApprovedBy            *uuid.UUID `db:"approved_by"`
	ApprovalStatus        string     `db:"approval_status"`
	ApprovalReason        sql.NullString `db:"approval_reason"`
	RejectedBy            *uuid.UUID     `db:"rejected_by"`
	RejectionReason       sql.NullString `db:"rejection_reason"`
	QuantityChange        int        `db:"quantity_change"`
	ReasonCode            string     `db:"reason_code"`
	Notes                 string     `db:"notes"`
	AdjustmentDate        time.Time  `db:"adjustment_date"`
	CreatedDate           time.Time  `db:"created_date"`
	UpdatedDate           time.Time  `db:"updated_date"`
}

func toBusInventoryAdjustment(ia inventoryAdjustment) inventoryadjustmentbus.InventoryAdjustment {
	bus := inventoryadjustmentbus.InventoryAdjustment{
		InventoryAdjustmentID: ia.InventoryAdjustmentID,
		ProductID:             ia.ProductID,
		LocationID:            ia.LocationID,
		AdjustedBy:            ia.AdjustedBy,
		ApprovedBy:            ia.ApprovedBy,
		ApprovalStatus:        ia.ApprovalStatus,
		RejectedBy:            ia.RejectedBy,
		QuantityChange:        ia.QuantityChange,
		ReasonCode:            ia.ReasonCode,
		Notes:                 ia.Notes,
		AdjustmentDate:        ia.AdjustmentDate,
		CreatedDate:           ia.CreatedDate,
		UpdatedDate:           ia.UpdatedDate,
	}

	if ia.ApprovalReason.Valid {
		bus.ApprovalReason = ia.ApprovalReason.String
	}

	if ia.RejectionReason.Valid {
		bus.RejectionReason = ia.RejectionReason.String
	}

	return bus
}

func toBusInventoryAdjustments(ias []inventoryAdjustment) []inventoryadjustmentbus.InventoryAdjustment {
	bus := make([]inventoryadjustmentbus.InventoryAdjustment, len(ias))
	for i, ia := range ias {
		bus[i] = toBusInventoryAdjustment(ia)
	}
	return bus
}

func toDBInventoryAdjustment(ia inventoryadjustmentbus.InventoryAdjustment) inventoryAdjustment {
	db := inventoryAdjustment{
		InventoryAdjustmentID: ia.InventoryAdjustmentID,
		ProductID:             ia.ProductID,
		LocationID:            ia.LocationID,
		AdjustedBy:            ia.AdjustedBy,
		ApprovedBy:            ia.ApprovedBy,
		ApprovalStatus:        ia.ApprovalStatus,
		RejectedBy:            ia.RejectedBy,
		QuantityChange:        ia.QuantityChange,
		ReasonCode:            ia.ReasonCode,
		Notes:                 ia.Notes,
		AdjustmentDate:        ia.AdjustmentDate,
		CreatedDate:           ia.CreatedDate,
		UpdatedDate:           ia.UpdatedDate,
	}

	if ia.ApprovalReason != "" {
		db.ApprovalReason = sql.NullString{String: ia.ApprovalReason, Valid: true}
	}

	if ia.RejectionReason != "" {
		db.RejectionReason = sql.NullString{String: ia.RejectionReason, Valid: true}
	}

	return db
}
