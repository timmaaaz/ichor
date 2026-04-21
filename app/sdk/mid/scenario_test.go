package mid

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/scenarios/scenariobus"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
)

type fakeActiveReader struct {
	scenario scenariobus.Scenario
	err      error
}

func (f fakeActiveReader) Active(ctx context.Context) (scenariobus.Scenario, error) {
	return f.scenario, f.err
}

func TestActiveScenario_PopulatesBothContextKeys(t *testing.T) {
	want := uuid.New()
	bus := fakeActiveReader{scenario: scenariobus.Scenario{ID: want}}

	var gotMid, gotSQLDB uuid.UUID
	var midOK, sqldbOK bool

	next := func(ctx context.Context) Encoder {
		gotMid, midOK = GetScenario(ctx)
		gotSQLDB, sqldbOK = sqldb.GetScenarioFilter(ctx)
		return nil
	}

	if resp := ActiveScenario(context.Background(), bus, next); resp != nil {
		if err := isError(resp); err != nil {
			t.Fatalf("unexpected error response: %v", err)
		}
	}

	if !midOK || gotMid != want {
		t.Errorf("mid scenario key: got (%v, %v); want (%v, true)", gotMid, midOK, want)
	}
	if !sqldbOK || gotSQLDB != want {
		t.Errorf("sqldb scenario key: got (%v, %v); want (%v, true)", gotSQLDB, sqldbOK, want)
	}
}

func TestActiveScenario_NoActiveScenarioPassesThrough(t *testing.T) {
	bus := fakeActiveReader{err: scenariobus.ErrNotFound}

	var midOK, sqldbOK bool
	nextCalled := false

	next := func(ctx context.Context) Encoder {
		nextCalled = true
		_, midOK = GetScenario(ctx)
		_, sqldbOK = sqldb.GetScenarioFilter(ctx)
		return nil
	}

	if resp := ActiveScenario(context.Background(), bus, next); resp != nil {
		if err := isError(resp); err != nil {
			t.Fatalf("unexpected error response: %v", err)
		}
	}

	if !nextCalled {
		t.Fatal("expected next handler to run when no scenario is active")
	}
	if midOK {
		t.Error("mid scenario key must not be populated when no scenario is active")
	}
	if sqldbOK {
		t.Error("sqldb scenario key must not be populated when no scenario is active")
	}
}

func TestActiveScenario_PropagatesOtherErrors(t *testing.T) {
	wantErr := errors.New("db unavailable")
	bus := fakeActiveReader{err: wantErr}

	nextCalled := false
	next := func(ctx context.Context) Encoder {
		nextCalled = true
		return nil
	}

	resp := ActiveScenario(context.Background(), bus, next)

	if nextCalled {
		t.Error("next handler must not run when bus returns a non-ErrNotFound error")
	}
	if resp == nil {
		t.Fatal("expected an error Encoder, got nil")
	}
	if err := isError(resp); err == nil {
		t.Fatal("expected response to encode an error")
	}
}
