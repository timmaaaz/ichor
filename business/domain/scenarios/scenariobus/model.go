package scenariobus

import (
	"time"

	"github.com/google/uuid"
)

// Scenario represents a named test scenario for floor warehouse testing.
type Scenario struct {
	ID          uuid.UUID
	Name        string
	Description string
	CreatedDate time.Time
	UpdatedDate time.Time
}

// NewScenario is the input for creating a new scenario.
type NewScenario struct {
	Name        string
	Description string
}

// UpdateScenario carries optional patch fields.
type UpdateScenario struct {
	Name        *string
	Description *string
}

// ScenarioFixture represents a single fixture row belonging to a scenario.
// Fixtures are write-once (WORM) — only the seeder inserts them; no API
// creates or modifies fixture rows.
type ScenarioFixture struct {
	ID          uuid.UUID
	ScenarioID  uuid.UUID
	TargetTable string
	PayloadJSON []byte
	CreatedDate time.Time
}

// NewScenarioFixture is the input for seeding a new fixture row.
type NewScenarioFixture struct {
	ScenarioID  uuid.UUID
	TargetTable string
	PayloadJSON []byte
}

// WorkerZoneBinding represents a single user→zones assignment applied to
// core.users.assigned_zones at scenario Load time. Bus-layer value type;
// yamlload.WorkerBinding is the YAML-layer parallel.
type WorkerZoneBinding struct {
	Username string
	Zones    []string
}

// SettingOverride represents one row of config.scenario_setting_overrides.
// Persisted at seed time (see seed_scenarios.go) so the settings resolver
// LEFT JOIN sees a stable per-scenario override layer.
type SettingOverride struct {
	ScenarioID uuid.UUID
	Key        string
	Value      string
}
