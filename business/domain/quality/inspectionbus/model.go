package inspectionbus

import (
	"time"

	"github.com/google/uuid"
)

type Inspection struct {
	InspectionID       uuid.UUID
	ProductID          uuid.UUID
	InspectorID        uuid.UUID
	LotID              uuid.UUID
	Status             string
	Notes              string
	InspectionDate     time.Time
	NextInspectionDate time.Time
	UpdatedDate        time.Time
	CreatedDate        time.Time
}

type NewInspection struct {
	ProductID          uuid.UUID
	InspectorID        uuid.UUID
	LotID              uuid.UUID
	Status             string
	Notes              string
	InspectionDate     time.Time
	NextInspectionDate time.Time
}

type UpdateInspection struct {
	ProductID          *uuid.UUID
	InspectorID        *uuid.UUID
	LotID              *uuid.UUID
	Status             *string
	Notes              *string
	InspectionDate     *time.Time
	NextInspectionDate *time.Time
}
