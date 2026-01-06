package reportstobus

import "github.com/google/uuid"

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

type ReportsTo struct {
	ID         uuid.UUID `json:"id"`
	ReporterID uuid.UUID `json:"reporter_id"`
	BossID     uuid.UUID `json:"boss_id"`
}

type NewReportsTo struct {
	ReporterID uuid.UUID `json:"reporter_id"`
	BossID     uuid.UUID `json:"boss_id"`
}

type UpdateReportsTo struct {
	ReporterID *uuid.UUID `json:"reporter_id,omitempty"`
	BossID     *uuid.UUID `json:"boss_id,omitempty"`
}
