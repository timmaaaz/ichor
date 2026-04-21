package cyclecountsessiondb

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountsessionbus"
)

// cycleCountSession mirrors the inventory.cycle_count_sessions DB row.
type cycleCountSession struct {
	ID            uuid.UUID    `db:"id"`
	Name          string       `db:"name"`
	Status        string       `db:"status"`
	CreatedBy     uuid.UUID    `db:"created_by"`
	CreatedDate   time.Time    `db:"created_date"`
	UpdatedDate   time.Time    `db:"updated_date"`
	CompletedDate sql.NullTime `db:"completed_date"`
	ScenarioID    *uuid.UUID   `db:"scenario_id"`
}

func toBusCycleCountSession(db cycleCountSession) (cyclecountsessionbus.CycleCountSession, error) {
	status, err := cyclecountsessionbus.ParseStatus(db.Status)
	if err != nil {
		return cyclecountsessionbus.CycleCountSession{}, fmt.Errorf("parse status %q: %w", db.Status, err)
	}

	var completedDate *time.Time
	if db.CompletedDate.Valid {
		t := db.CompletedDate.Time
		completedDate = &t
	}

	return cyclecountsessionbus.CycleCountSession{
		ID:            db.ID,
		Name:          db.Name,
		Status:        status,
		CreatedBy:     db.CreatedBy,
		CreatedDate:   db.CreatedDate,
		UpdatedDate:   db.UpdatedDate,
		CompletedDate: completedDate,
		ScenarioID:    db.ScenarioID,
	}, nil
}

func toBusCycleCountSessions(dbs []cycleCountSession) ([]cyclecountsessionbus.CycleCountSession, error) {
	sessions := make([]cyclecountsessionbus.CycleCountSession, len(dbs))
	for i, db := range dbs {
		s, err := toBusCycleCountSession(db)
		if err != nil {
			return nil, err
		}
		sessions[i] = s
	}
	return sessions, nil
}

func toDBCycleCountSession(bus cyclecountsessionbus.CycleCountSession) cycleCountSession {
	var completedDate sql.NullTime
	if bus.CompletedDate != nil {
		completedDate = sql.NullTime{Time: bus.CompletedDate.UTC(), Valid: true}
	}

	return cycleCountSession{
		ID:            bus.ID,
		Name:          bus.Name,
		Status:        bus.Status.String(),
		CreatedBy:     bus.CreatedBy,
		CreatedDate:   bus.CreatedDate,
		UpdatedDate:   bus.UpdatedDate,
		CompletedDate: completedDate,
		ScenarioID:    bus.ScenarioID,
	}
}
