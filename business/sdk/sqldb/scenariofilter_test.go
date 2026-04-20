package sqldb_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
)

func TestApplyScenarioFilter_NoCtx(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{}

	// No scenario in context — helper must emit nothing and leave data untouched.
	sqldb.ApplyScenarioFilter(context.Background(), &buf, data)

	if buf.Len() != 0 {
		t.Fatalf("expected empty buffer with no scenario in ctx, got %q", buf.String())
	}
	if _, ok := data["scenario_id"]; ok {
		t.Fatalf("expected no scenario_id key in data map, got %v", data)
	}
}

func TestApplyScenarioFilter_WithCtx(t *testing.T) {
	id := uuid.New()
	ctx := sqldb.SetScenarioFilter(context.Background(), id)

	var buf bytes.Buffer
	data := map[string]any{}

	sqldb.ApplyScenarioFilter(ctx, &buf, data)

	out := buf.String()
	if !strings.Contains(out, "scenario_id") {
		t.Fatalf("expected scenario_id reference in output, got %q", out)
	}
	if got := data["scenario_id"]; got != id {
		t.Fatalf("expected data[scenario_id] == %s, got %v", id, got)
	}
}

func TestApplyScenarioFilter_AppendsToExistingWhere(t *testing.T) {
	id := uuid.New()
	ctx := sqldb.SetScenarioFilter(context.Background(), id)

	// Simulate a buffer that already has WHERE + one clause appended by the
	// caller's applyFilter. Helper must emit " AND (...)" in that case.
	var buf bytes.Buffer
	buf.WriteString(" WHERE warehouse_id = :warehouse_id")
	data := map[string]any{"warehouse_id": uuid.New()}

	sqldb.ApplyScenarioFilter(ctx, &buf, data)

	out := buf.String()
	if !strings.Contains(out, " AND (scenario_id IS NULL OR scenario_id = :scenario_id)") {
		t.Fatalf("expected AND-joined scenario clause, got %q", out)
	}
}

func TestApplyScenarioFilter_StartsWhereWhenEmpty(t *testing.T) {
	id := uuid.New()
	ctx := sqldb.SetScenarioFilter(context.Background(), id)

	var buf bytes.Buffer
	data := map[string]any{}

	sqldb.ApplyScenarioFilter(ctx, &buf, data)

	out := buf.String()
	if !strings.Contains(out, " WHERE (scenario_id IS NULL OR scenario_id = :scenario_id)") {
		t.Fatalf("expected fresh WHERE scenario clause, got %q", out)
	}
}

func TestGetScenarioFilter_ZeroUUIDTreatedAsAbsent(t *testing.T) {
	// A zero uuid.UUID in ctx is indistinguishable from "no scenario"; the
	// helper should return (zero, false) so callers don't accidentally inject
	// a WHERE clause bound to the all-zeroes UUID.
	ctx := sqldb.SetScenarioFilter(context.Background(), uuid.UUID{})
	id, ok := sqldb.GetScenarioFilter(ctx)
	if ok {
		t.Fatalf("expected (_, false) for zero UUID, got (%s, true)", id)
	}
}
