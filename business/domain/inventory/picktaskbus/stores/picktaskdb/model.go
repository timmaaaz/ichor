package picktaskdb

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/picktaskbus"
	"github.com/timmaaaz/ichor/business/sdk/sqldb/nulltypes"
)

// pickTask mirrors the inventory.pick_tasks DB row.
type pickTask struct {
	ID                   uuid.UUID      `db:"id"`
	TaskNumber           sql.NullString `db:"task_number"`
	SalesOrderID         uuid.UUID      `db:"sales_order_id"`
	SalesOrderLineItemID uuid.UUID      `db:"sales_order_line_item_id"`
	ProductID            uuid.UUID      `db:"product_id"`
	LotID                uuid.NullUUID  `db:"lot_id"`
	SerialID             uuid.NullUUID  `db:"serial_id"`
	LocationID           uuid.UUID      `db:"location_id"`
	QuantityToPick       int            `db:"quantity_to_pick"`
	QuantityPicked       int            `db:"quantity_picked"`
	Status               string         `db:"status"`
	AssignedTo           sql.NullString `db:"assigned_to"`
	AssignedAt           sql.NullTime   `db:"assigned_at"`
	CompletedBy          sql.NullString `db:"completed_by"`
	CompletedAt          sql.NullTime   `db:"completed_at"`
	ShortPickReason      sql.NullString `db:"short_pick_reason"`
	CreatedBy            uuid.UUID      `db:"created_by"`
	CreatedDate          time.Time      `db:"created_date"`
	UpdatedDate          time.Time      `db:"updated_date"`
}

func toBusPickTask(db pickTask) (picktaskbus.PickTask, error) {
	var taskNumber *string
	if db.TaskNumber.Valid {
		s := db.TaskNumber.String
		taskNumber = &s
	}

	var lotID *uuid.UUID
	if db.LotID.Valid {
		id := db.LotID.UUID
		lotID = &id
	}

	var serialID *uuid.UUID
	if db.SerialID.Valid {
		id := db.SerialID.UUID
		serialID = &id
	}

	var assignedAt time.Time
	if db.AssignedAt.Valid {
		assignedAt = db.AssignedAt.Time
	}

	var completedAt time.Time
	if db.CompletedAt.Valid {
		completedAt = db.CompletedAt.Time
	}

	var shortPickReason string
	if db.ShortPickReason.Valid {
		shortPickReason = db.ShortPickReason.String
	}

	status, err := picktaskbus.ParseStatus(db.Status)
	if err != nil {
		return picktaskbus.PickTask{}, fmt.Errorf("parse status %q: %w", db.Status, err)
	}

	return picktaskbus.PickTask{
		ID:                   db.ID,
		TaskNumber:           taskNumber,
		SalesOrderID:         db.SalesOrderID,
		SalesOrderLineItemID: db.SalesOrderLineItemID,
		ProductID:            db.ProductID,
		LotID:                lotID,
		SerialID:             serialID,
		LocationID:           db.LocationID,
		QuantityToPick:       db.QuantityToPick,
		QuantityPicked:       db.QuantityPicked,
		Status:               status,
		AssignedTo:           nulltypes.FromNullableUUID(db.AssignedTo),
		AssignedAt:           assignedAt,
		CompletedBy:          nulltypes.FromNullableUUID(db.CompletedBy),
		CompletedAt:          completedAt,
		ShortPickReason:      shortPickReason,
		CreatedBy:            db.CreatedBy,
		CreatedDate:          db.CreatedDate,
		UpdatedDate:          db.UpdatedDate,
	}, nil
}

func toBusPickTasks(dbs []pickTask) ([]picktaskbus.PickTask, error) {
	tasks := make([]picktaskbus.PickTask, len(dbs))
	for i, db := range dbs {
		t, err := toBusPickTask(db)
		if err != nil {
			return nil, err
		}
		tasks[i] = t
	}
	return tasks, nil
}

func toDBPickTask(bus picktaskbus.PickTask) pickTask {
	var taskNumber sql.NullString
	if bus.TaskNumber != nil {
		taskNumber = sql.NullString{String: *bus.TaskNumber, Valid: true}
	}

	var lotID uuid.NullUUID
	if bus.LotID != nil {
		lotID = uuid.NullUUID{UUID: *bus.LotID, Valid: true}
	}

	var serialID uuid.NullUUID
	if bus.SerialID != nil {
		serialID = uuid.NullUUID{UUID: *bus.SerialID, Valid: true}
	}

	var assignedAt sql.NullTime
	if !bus.AssignedAt.IsZero() {
		assignedAt = sql.NullTime{Time: bus.AssignedAt.UTC(), Valid: true}
	}

	var completedAt sql.NullTime
	if !bus.CompletedAt.IsZero() {
		completedAt = sql.NullTime{Time: bus.CompletedAt.UTC(), Valid: true}
	}

	var shortPickReason sql.NullString
	if bus.ShortPickReason != "" {
		shortPickReason = sql.NullString{String: bus.ShortPickReason, Valid: true}
	}

	return pickTask{
		ID:                   bus.ID,
		TaskNumber:           taskNumber,
		SalesOrderID:         bus.SalesOrderID,
		SalesOrderLineItemID: bus.SalesOrderLineItemID,
		ProductID:            bus.ProductID,
		LotID:                lotID,
		SerialID:             serialID,
		LocationID:           bus.LocationID,
		QuantityToPick:       bus.QuantityToPick,
		QuantityPicked:       bus.QuantityPicked,
		Status:               bus.Status.String(),
		AssignedTo:           nulltypes.ToNullableUUID(bus.AssignedTo),
		AssignedAt:           assignedAt,
		CompletedBy:          nulltypes.ToNullableUUID(bus.CompletedBy),
		CompletedAt:          completedAt,
		ShortPickReason:      shortPickReason,
		CreatedBy:            bus.CreatedBy,
		CreatedDate:          bus.CreatedDate,
		UpdatedDate:          bus.UpdatedDate,
	}
}
