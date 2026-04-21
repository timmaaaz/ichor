package sqldb

import (
	"bytes"
	"context"
	"strings"

	"github.com/google/uuid"
)

type scenarioKey struct{}

// SetScenarioFilter returns a new context carrying the active scenario id.
// Pass uuid.UUID{} (zero value) to indicate "no scenario" — GetScenarioFilter
// treats zero as absent so callers cannot accidentally filter against an
// all-zeroes scenario id. The mid layer populates this on request entry from
// inventory.scenarios_active; repositories read via ApplyScenarioFilter at
// the Query/QueryByID call site inside each store.
func SetScenarioFilter(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, scenarioKey{}, id)
}

// GetScenarioFilter returns the active scenario id from ctx and a bool
// indicating presence. A zero uuid.UUID is treated as absent. Callers
// that need to TAG writes (not filter reads) — e.g. bus.Create populating
// ScenarioID on new rows — use this directly.
func GetScenarioFilter(ctx context.Context) (uuid.UUID, bool) {
	v, ok := ctx.Value(scenarioKey{}).(uuid.UUID)
	if !ok || v == (uuid.UUID{}) {
		return uuid.UUID{}, false
	}
	return v, true
}

// ApplyScenarioFilter appends a scenario-scoped WHERE or AND clause to buf
// when an active scenario is present in ctx. The clause restricts reads to
// rows that are either baseline (scenario_id IS NULL) or belong to the
// active scenario. When ctx carries no scenario, buf and data are untouched
// so non-scoped queries remain unaffected.
//
// Call site convention: invoke at the end of Query() and QueryByID() in
// each floor-scoped store, immediately after the existing applyFilter(...)
// returns and immediately before the sqldb.NamedQuery* dispatch. This lets
// the helper observe the buffer tail (to choose WHERE vs AND) without
// modifying any existing filter.go signature.
func ApplyScenarioFilter(ctx context.Context, buf *bytes.Buffer, data map[string]any) {
	id, ok := GetScenarioFilter(ctx)
	if !ok {
		return
	}
	data["scenario_id"] = id

	if strings.Contains(buf.String(), " WHERE ") {
		buf.WriteString(" AND (scenario_id IS NULL OR scenario_id = :scenario_id)")
		return
	}
	buf.WriteString(" WHERE (scenario_id IS NULL OR scenario_id = :scenario_id)")
}
