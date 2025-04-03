package inspectiondb

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/quality/inspectionbus"
)

type inspection struct {
	InspectionID       uuid.UUID `db:"inspection_id"`
	ProductID          uuid.UUID `db:"product_id"`
	InspectorID        uuid.UUID `db:"inspector_id"`
	LotID              uuid.UUID `db:"lot_id"`
	Status             string    `db:"status"`
	Notes              string    `db:"notes"`
	InspectionDate     time.Time `db:"inspection_date"`
	NextInspectionDate time.Time `db:"next_inspection_date"`
	UpdatedDate        time.Time `db:"updated_date"`
	CreatedDate        time.Time `db:"created_date"`
}

func toBusInspection(db inspection) inspectionbus.Inspection {
	return inspectionbus.Inspection{
		InspectionID:       db.InspectionID,
		ProductID:          db.ProductID,
		InspectorID:        db.InspectorID,
		LotID:              db.LotID,
		Status:             db.Status,
		Notes:              db.Notes,
		InspectionDate:     db.InspectionDate,
		NextInspectionDate: db.NextInspectionDate,
		UpdatedDate:        db.UpdatedDate,
		CreatedDate:        db.CreatedDate,
	}
}

func toBusInspections(db []inspection) []inspectionbus.Inspection {
	busInspections := make([]inspectionbus.Inspection, len(db))
	for i, dbInspection := range db {
		busInspections[i] = toBusInspection(dbInspection)
	}
	return busInspections
}

func toDBInspection(bus inspectionbus.Inspection) inspection {
	return inspection{
		InspectionID:       bus.InspectionID,
		ProductID:          bus.ProductID,
		InspectorID:        bus.InspectorID,
		LotID:              bus.LotID,
		Status:             bus.Status,
		Notes:              bus.Notes,
		InspectionDate:     bus.InspectionDate,
		NextInspectionDate: bus.NextInspectionDate,
		UpdatedDate:        bus.UpdatedDate,
		CreatedDate:        bus.CreatedDate,
	}
}
