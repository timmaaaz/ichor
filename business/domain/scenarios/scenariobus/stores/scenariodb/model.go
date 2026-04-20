package scenariodb

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/scenarios/scenariobus"
)

type dbScenario struct {
	ID          uuid.UUID `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	CreatedDate time.Time `db:"created_date"`
	UpdatedDate time.Time `db:"updated_date"`
}

func toDBScenario(bus scenariobus.Scenario) dbScenario {
	return dbScenario{
		ID:          bus.ID,
		Name:        bus.Name,
		Description: bus.Description,
		CreatedDate: bus.CreatedDate.UTC(),
		UpdatedDate: bus.UpdatedDate.UTC(),
	}
}

func toBusScenario(db dbScenario) scenariobus.Scenario {
	return scenariobus.Scenario{
		ID:          db.ID,
		Name:        db.Name,
		Description: db.Description,
		CreatedDate: db.CreatedDate.In(time.Local),
		UpdatedDate: db.UpdatedDate.In(time.Local),
	}
}

func toBusScenarios(dbs []dbScenario) []scenariobus.Scenario {
	out := make([]scenariobus.Scenario, len(dbs))
	for i, d := range dbs {
		out[i] = toBusScenario(d)
	}
	return out
}

// dbScenarioFixture maps to inventory.scenario_fixtures.
type dbScenarioFixture struct {
	ID          uuid.UUID `db:"id"`
	ScenarioID  uuid.UUID `db:"scenario_id"`
	TargetTable string    `db:"target_table"`
	PayloadJSON []byte    `db:"payload_json"`
	CreatedDate time.Time `db:"created_date"`
}

func toDBScenarioFixture(bus scenariobus.ScenarioFixture) dbScenarioFixture {
	return dbScenarioFixture{
		ID:          bus.ID,
		ScenarioID:  bus.ScenarioID,
		TargetTable: bus.TargetTable,
		PayloadJSON: bus.PayloadJSON,
		CreatedDate: bus.CreatedDate.UTC(),
	}
}

func toBusScenarioFixture(db dbScenarioFixture) scenariobus.ScenarioFixture {
	return scenariobus.ScenarioFixture{
		ID:          db.ID,
		ScenarioID:  db.ScenarioID,
		TargetTable: db.TargetTable,
		PayloadJSON: db.PayloadJSON,
		CreatedDate: db.CreatedDate.In(time.Local),
	}
}

func toBusScenarioFixtures(dbs []dbScenarioFixture) []scenariobus.ScenarioFixture {
	out := make([]scenariobus.ScenarioFixture, len(dbs))
	for i, d := range dbs {
		out[i] = toBusScenarioFixture(d)
	}
	return out
}

// dbScenariosActive maps to inventory.scenarios_active singleton row.
type dbScenariosActive struct {
	ScenarioID *uuid.UUID `db:"scenario_id"`
}
