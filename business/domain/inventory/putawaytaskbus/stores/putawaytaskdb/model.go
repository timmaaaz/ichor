package putawaytaskdb

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/putawaytaskbus"
	"github.com/timmaaaz/ichor/business/sdk/sqldb/nulltypes"
)

// putAwayTask mirrors the inventory.put_away_tasks DB row.
// Nullable columns use sql.NullString / sql.NullTime.
type putAwayTask struct {
	ID              uuid.UUID      `db:"id"`
	ProductID       uuid.UUID      `db:"product_id"`
	LocationID      uuid.UUID      `db:"location_id"`
	Quantity        int            `db:"quantity"`
	ReferenceNumber string         `db:"reference_number"`
	Status          string         `db:"status"`
	AssignedTo      sql.NullString `db:"assigned_to"`
	AssignedAt      sql.NullTime   `db:"assigned_at"`
	CompletedBy     sql.NullString `db:"completed_by"`
	CompletedAt     sql.NullTime   `db:"completed_at"`
	CreatedBy       uuid.UUID      `db:"created_by"`
	CreatedDate     time.Time      `db:"created_date"`
	UpdatedDate     time.Time      `db:"updated_date"`
}

func toBusPutAwayTask(db putAwayTask) (putawaytaskbus.PutAwayTask, error) {
	var assignedAt time.Time
	if db.AssignedAt.Valid {
		assignedAt = db.AssignedAt.Time
	}

	var completedAt time.Time
	if db.CompletedAt.Valid {
		completedAt = db.CompletedAt.Time
	}

	status, err := putawaytaskbus.ParseStatus(db.Status)
	if err != nil {
		return putawaytaskbus.PutAwayTask{}, fmt.Errorf("parse status %q: %w", db.Status, err)
	}

	return putawaytaskbus.PutAwayTask{
		ID:              db.ID,
		ProductID:       db.ProductID,
		LocationID:      db.LocationID,
		Quantity:        db.Quantity,
		ReferenceNumber: db.ReferenceNumber,
		Status:          status,
		AssignedTo:      nulltypes.FromNullableUUID(db.AssignedTo),
		AssignedAt:      assignedAt,
		CompletedBy:     nulltypes.FromNullableUUID(db.CompletedBy),
		CompletedAt:     completedAt,
		CreatedBy:       db.CreatedBy,
		CreatedDate:     db.CreatedDate,
		UpdatedDate:     db.UpdatedDate,
	}, nil
}

func toBusPutAwayTasks(dbs []putAwayTask) ([]putawaytaskbus.PutAwayTask, error) {
	tasks := make([]putawaytaskbus.PutAwayTask, len(dbs))
	for i, db := range dbs {
		t, err := toBusPutAwayTask(db)
		if err != nil {
			return nil, err
		}
		tasks[i] = t
	}
	return tasks, nil
}

func toDBPutAwayTask(bus putawaytaskbus.PutAwayTask) putAwayTask {
	var assignedAt sql.NullTime
	if !bus.AssignedAt.IsZero() {
		assignedAt = sql.NullTime{Time: bus.AssignedAt.UTC(), Valid: true}
	}

	var completedAt sql.NullTime
	if !bus.CompletedAt.IsZero() {
		completedAt = sql.NullTime{Time: bus.CompletedAt.UTC(), Valid: true}
	}

	return putAwayTask{
		ID:              bus.ID,
		ProductID:       bus.ProductID,
		LocationID:      bus.LocationID,
		Quantity:        bus.Quantity,
		ReferenceNumber: bus.ReferenceNumber,
		Status:          bus.Status.String(),
		AssignedTo:      nulltypes.ToNullableUUID(bus.AssignedTo),
		AssignedAt:      assignedAt,
		CompletedBy:     nulltypes.ToNullableUUID(bus.CompletedBy),
		CompletedAt:     completedAt,
		CreatedBy:       bus.CreatedBy,
		CreatedDate:     bus.CreatedDate,
		UpdatedDate:     bus.UpdatedDate,
	}
}
