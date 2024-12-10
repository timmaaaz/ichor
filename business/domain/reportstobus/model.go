package reportstobus

import "github.com/google/uuid"

type ReportsTo struct {
	ID         uuid.UUID
	ReporterID uuid.UUID
	BossID     uuid.UUID
}

type NewReportsTo struct {
	ReporterID uuid.UUID
	BossID     uuid.UUID
}

type UpdateReportsTo struct {
	ReporterID *uuid.UUID
	BossID     *uuid.UUID
}
