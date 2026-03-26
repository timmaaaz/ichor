package cyclecountitemdb

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountitembus"
	"github.com/timmaaaz/ichor/business/sdk/sqldb/nulltypes"
)

// cycleCountItem mirrors the inventory.cycle_count_items DB row.
type cycleCountItem struct {
	ID              uuid.UUID      `db:"id"`
	SessionID       uuid.UUID      `db:"session_id"`
	ProductID       uuid.UUID      `db:"product_id"`
	LocationID      uuid.UUID      `db:"location_id"`
	SystemQuantity  int            `db:"system_quantity"`
	CountedQuantity sql.NullInt64  `db:"counted_quantity"`
	Variance        sql.NullInt64  `db:"variance"`
	Status          string         `db:"status"`
	CountedBy       sql.NullString `db:"counted_by"`
	CountedDate     sql.NullTime   `db:"counted_date"`
	CreatedDate     time.Time      `db:"created_date"`
	UpdatedDate     time.Time      `db:"updated_date"`
}

func toBusCycleCountItem(db cycleCountItem) (cyclecountitembus.CycleCountItem, error) {
	status, err := cyclecountitembus.ParseStatus(db.Status)
	if err != nil {
		return cyclecountitembus.CycleCountItem{}, fmt.Errorf("parse status %q: %w", db.Status, err)
	}

	var countedDate time.Time
	if db.CountedDate.Valid {
		countedDate = db.CountedDate.Time
	}

	return cyclecountitembus.CycleCountItem{
		ID:              db.ID,
		SessionID:       db.SessionID,
		ProductID:       db.ProductID,
		LocationID:      db.LocationID,
		SystemQuantity:  db.SystemQuantity,
		CountedQuantity: nulltypes.Int64Ptr(db.CountedQuantity),
		Variance:        nulltypes.Int64Ptr(db.Variance),
		Status:          status,
		CountedBy:       nulltypes.FromNullableUUID(db.CountedBy),
		CountedDate:     countedDate,
		CreatedDate:     db.CreatedDate,
		UpdatedDate:     db.UpdatedDate,
	}, nil
}

func toBusCycleCountItems(dbs []cycleCountItem) ([]cyclecountitembus.CycleCountItem, error) {
	items := make([]cyclecountitembus.CycleCountItem, len(dbs))
	for i, db := range dbs {
		item, err := toBusCycleCountItem(db)
		if err != nil {
			return nil, err
		}
		items[i] = item
	}
	return items, nil
}

func toDBCycleCountItem(bus cyclecountitembus.CycleCountItem) cycleCountItem {
	var countedDate sql.NullTime
	if !bus.CountedDate.IsZero() {
		countedDate = sql.NullTime{Time: bus.CountedDate.UTC(), Valid: true}
	}

	return cycleCountItem{
		ID:              bus.ID,
		SessionID:       bus.SessionID,
		ProductID:       bus.ProductID,
		LocationID:      bus.LocationID,
		SystemQuantity:  bus.SystemQuantity,
		CountedQuantity: nulltypes.ToNullInt64(bus.CountedQuantity),
		Variance:        nulltypes.ToNullInt64(bus.Variance),
		Status:          bus.Status.String(),
		CountedBy:       nulltypes.ToNullableUUID(bus.CountedBy),
		CountedDate:     countedDate,
		CreatedDate:     bus.CreatedDate,
		UpdatedDate:     bus.UpdatedDate,
	}
}
