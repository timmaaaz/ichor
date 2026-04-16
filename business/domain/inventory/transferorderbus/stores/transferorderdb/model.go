package transferorderdb

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus"
)

type transferOrder struct {
	TransferID      uuid.UUID      `db:"id"`
	TransferNumber  sql.NullString `db:"transfer_number"`
	ProductID       uuid.UUID      `db:"product_id"`
	FromLocationID  uuid.UUID     `db:"from_location_id"`
	ToLocationID    uuid.UUID     `db:"to_location_id"`
	RequestedByID   uuid.UUID     `db:"requested_by"`
	ApprovedByID    uuid.NullUUID `db:"approved_by"`
	RejectedByID    uuid.NullUUID `db:"rejected_by_id"`
	ApprovalReason  sql.NullString `db:"approval_reason"`
	RejectionReason sql.NullString `db:"rejection_reason"`
	ClaimedByID     uuid.NullUUID `db:"claimed_by"`
	ClaimedAt       sql.NullTime  `db:"claimed_at"`
	CompletedByID   uuid.NullUUID `db:"completed_by"`
	CompletedAt     sql.NullTime  `db:"completed_at"`
	Quantity        int           `db:"quantity"`
	Status          string        `db:"status"`
	TransferDate    time.Time     `db:"transfer_date"`
	CreatedDate     time.Time     `db:"created_date"`
	UpdatedDate     time.Time     `db:"updated_date"`
}

func toBusTransferOrder(db transferOrder) transferorderbus.TransferOrder {
	var transferNumber *string
	if db.TransferNumber.Valid {
		s := db.TransferNumber.String
		transferNumber = &s
	}

	to := transferorderbus.TransferOrder{
		TransferID:     db.TransferID,
		TransferNumber: transferNumber,
		ProductID:      db.ProductID,
		FromLocationID: db.FromLocationID,
		ToLocationID:   db.ToLocationID,
		RequestedByID:  db.RequestedByID,
		Quantity:       db.Quantity,
		Status:         db.Status,
		TransferDate:   db.TransferDate,
		CreatedDate:    db.CreatedDate,
		UpdatedDate:    db.UpdatedDate,
	}

	if db.ApprovalReason.Valid {
		to.ApprovalReason = db.ApprovalReason.String
	}

	if db.RejectionReason.Valid {
		to.RejectionReason = db.RejectionReason.String
	}

	if db.ApprovedByID.Valid {
		to.ApprovedByID = &db.ApprovedByID.UUID
	}

	if db.RejectedByID.Valid {
		to.RejectedByID = &db.RejectedByID.UUID
	}
	if db.ClaimedByID.Valid {
		to.ClaimedByID = &db.ClaimedByID.UUID
	}
	if db.ClaimedAt.Valid {
		t := db.ClaimedAt.Time
		to.ClaimedAt = &t
	}
	if db.CompletedByID.Valid {
		to.CompletedByID = &db.CompletedByID.UUID
	}
	if db.CompletedAt.Valid {
		t := db.CompletedAt.Time
		to.CompletedAt = &t
	}

	return to
}

func toBusTransferOrders(dbs []transferOrder) []transferorderbus.TransferOrder {
	app := make([]transferorderbus.TransferOrder, len(dbs))
	for i, db := range dbs {
		app[i] = toBusTransferOrder(db)
	}
	return app
}

func toDBTransferOrder(bus transferorderbus.TransferOrder) transferOrder {
	var transferNumber sql.NullString
	if bus.TransferNumber != nil {
		transferNumber = sql.NullString{String: *bus.TransferNumber, Valid: true}
	}

	db := transferOrder{
		TransferID:     bus.TransferID,
		TransferNumber: transferNumber,
		ProductID:      bus.ProductID,
		FromLocationID: bus.FromLocationID,
		ToLocationID:   bus.ToLocationID,
		RequestedByID:  bus.RequestedByID,
		Quantity:       bus.Quantity,
		Status:         bus.Status,
		TransferDate:   bus.TransferDate,
		CreatedDate:    bus.CreatedDate,
		UpdatedDate:    bus.UpdatedDate,
	}

	if bus.ApprovalReason != "" {
		db.ApprovalReason = sql.NullString{String: bus.ApprovalReason, Valid: true}
	}

	if bus.RejectionReason != "" {
		db.RejectionReason = sql.NullString{String: bus.RejectionReason, Valid: true}
	}

	if bus.ApprovedByID != nil {
		db.ApprovedByID = uuid.NullUUID{UUID: *bus.ApprovedByID, Valid: true}
	}

	if bus.RejectedByID != nil {
		db.RejectedByID = uuid.NullUUID{UUID: *bus.RejectedByID, Valid: true}
	}
	if bus.ClaimedByID != nil {
		db.ClaimedByID = uuid.NullUUID{UUID: *bus.ClaimedByID, Valid: true}
	}
	if bus.ClaimedAt != nil {
		db.ClaimedAt = sql.NullTime{Time: *bus.ClaimedAt, Valid: true}
	}
	if bus.CompletedByID != nil {
		db.CompletedByID = uuid.NullUUID{UUID: *bus.CompletedByID, Valid: true}
	}
	if bus.CompletedAt != nil {
		db.CompletedAt = sql.NullTime{Time: *bus.CompletedAt, Valid: true}
	}

	return db
}
