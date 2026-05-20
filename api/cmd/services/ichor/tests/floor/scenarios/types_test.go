package scenarios_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

type family string

const (
	familyReceive    family = "receive"
	familyPick       family = "pick"
	familyTransfer   family = "transfer"
	familyCycleCount family = "cycle-count"
	familyProfile    family = "profile"
)

// ScenarioRow is one row in the TestFloorScenarios table.
// Family-derived rows use the matching walkXxx helper; rows with
// Family == "" must populate Custom (e.g., profile-* scenarios).
type ScenarioRow struct {
	Name     string
	Category string // "generic" | "customer:<name>" — for -run filtering
	Family   family
	LotFlow  bool   // family-specific flag (receive/pick: emit lot prompts)
	Custom   func(t *testing.T, h *apitest.Test, db *sqlx.DB, scenarioID uuid.UUID)
}

// ReceiveInputs is the workflow-input bundle for familyReceive walks.
type ReceiveInputs struct {
	POID      uuid.UUID
	LineItems []ReceiveLineItem
}

type ReceiveLineItem struct {
	LineID            uuid.UUID
	ProductID         uuid.UUID
	UPC               string
	ExpectedQty       int
	LotTracked        bool
	SerialTracked     bool
	SupplierProductID uuid.UUID // for lot_trackings wiring
}

// PickInputs is the workflow-input bundle for familyPick walks.
type PickInputs struct {
	SOID        uuid.UUID
	Allocations []PickAllocation
}

type PickAllocation struct {
	PickTaskID uuid.UUID
	ProductID    uuid.UUID
	UPC          string
	LocationCode string // canonical, e.g., "STG-A01"
	Qty          int
	LotTracked   bool
}

// TransferInputs is the workflow-input bundle for familyTransfer walks.
type TransferInputs struct {
	TransferID uuid.UUID
	FromCode   string
	ToCode     string
	ProductID  uuid.UUID
	UPC        string
	Quantity   int
	LotTracked bool
}

// CycleCountInputs is the workflow-input bundle for familyCycleCount walks.
type CycleCountInputs struct {
	LocationCode string
	LocationID   uuid.UUID
	Items        []CycleCountItem
	VarianceMode string // "over" | "under" | "none" — derived from scenario name
}

type CycleCountItem struct {
	ProductID   uuid.UUID
	UPC         string
	ExpectedQty int
	ActualQty   int // for variance scenarios, differs from ExpectedQty
}

// ProfileFlags captures lever_overrides activated by a profile scenario.
// Populated by discoverProfileFlags; consumed by profile Custom handlers.
type ProfileFlags struct {
	LeverOverrides map[string]string
}
