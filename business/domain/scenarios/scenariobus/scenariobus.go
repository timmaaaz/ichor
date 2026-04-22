// Package scenariobus owns three tables — scenarios, scenario_fixtures,
// scenarios_active — as one aggregate. They share ON DELETE CASCADE /
// SET NULL constraints (migration 2.34), a single transactional boundary
// (Load/Reset mutate all three atomically), and no independent external
// lifecycle (fixtures are WORM, active is a singleton). See
// docs/superpowers/plans/floor-physical-warehouse-testing/NOTES.md
// 2026-04-20 "Why scenarios + scenario_fixtures + scenarios_active share
// one slice" for the full rationale.
package scenariobus

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound          = errors.New("scenario not found")
	ErrUniqueName        = errors.New("scenario name already exists")
	ErrNoActiveScenario  = errors.New("no active scenario set")
)

// Storer declares the behavior needed to persist and retrieve scenario data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)

	// scenarios table
	Create(ctx context.Context, s Scenario) error
	Update(ctx context.Context, s Scenario) error
	Delete(ctx context.Context, s Scenario) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Scenario, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, id uuid.UUID) (Scenario, error)
	QueryByName(ctx context.Context, name string) (Scenario, error)

	// scenario_fixtures table
	CreateFixture(ctx context.Context, f ScenarioFixture) error
	QueryFixturesByScenario(ctx context.Context, scenarioID uuid.UUID) ([]ScenarioFixture, error)

	// scenarios_active table (singleton row)
	QueryActive(ctx context.Context) (uuid.UUID, error) // returns uuid.Nil if no active scenario
	SetActive(ctx context.Context, id uuid.UUID) error  // UPSERT

	// bulk load/reset
	ApplyFixtures(ctx context.Context, target uuid.UUID) error
	DeleteScopedRows(ctx context.Context, scenarioID uuid.UUID) error
}

// Business manages the set of APIs for scenario access.
type Business struct {
	log      *logger.Logger
	delegate *delegate.Delegate
	storer   Storer
	beginner sqldb.Beginner
}

// NewBusiness constructs a scenario business API for use.
func NewBusiness(log *logger.Logger, d *delegate.Delegate, storer Storer, beginner sqldb.Beginner) *Business {
	return &Business{log: log, delegate: d, storer: storer, beginner: beginner}
}

// NewWithTx constructs a new Business value replacing the Storer value with
// a Storer value that is currently inside a transaction.
func (b *Business) NewWithTx(tx sqldb.CommitRollbacker) (*Business, error) {
	storer, err := b.storer.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	return &Business{
		log:      b.log,
		delegate: b.delegate,
		storer:   storer,
		beginner: b.beginner,
	}, nil
}

// Create inserts a new scenario.
func (b *Business) Create(ctx context.Context, ns NewScenario) (Scenario, error) {
	now := time.Now()
	s := Scenario{
		ID:          uuid.New(),
		Name:        ns.Name,
		Description: ns.Description,
		CreatedDate: now,
		UpdatedDate: now,
	}

	if err := b.storer.Create(ctx, s); err != nil {
		return Scenario{}, fmt.Errorf("create: %w", err)
	}

	if err := b.delegate.Call(ctx, ActionCreatedData(s)); err != nil {
		b.log.Error(ctx, "scenariobus: delegate call failed", "action", ActionCreated, "err", err)
	}

	return s, nil
}

// SeedCreate inserts a fully-formed Scenario (caller supplies the ID).
// Seed-only — preserves deterministic UUIDs across reseeds.
func (b *Business) SeedCreate(ctx context.Context, s Scenario) error {
	if s.CreatedDate.IsZero() {
		s.CreatedDate = time.Now()
	}
	if s.UpdatedDate.IsZero() {
		s.UpdatedDate = s.CreatedDate
	}
	if err := b.storer.Create(ctx, s); err != nil {
		return fmt.Errorf("seedcreate: %w", err)
	}
	return nil
}

// SeedCreateFixture inserts a fully-formed ScenarioFixture.
// Seed-only — fixtures are never created through the API; only the seeder writes them.
func (b *Business) SeedCreateFixture(ctx context.Context, f ScenarioFixture) error {
	if f.CreatedDate.IsZero() {
		f.CreatedDate = time.Now()
	}
	if err := b.storer.CreateFixture(ctx, f); err != nil {
		return fmt.Errorf("seedcreatefixture: %w", err)
	}
	return nil
}

// Update applies a partial patch to an existing scenario.
func (b *Business) Update(ctx context.Context, s Scenario, us UpdateScenario) (Scenario, error) {
	before := s

	if us.Name != nil {
		s.Name = *us.Name
	}
	if us.Description != nil {
		s.Description = *us.Description
	}
	s.UpdatedDate = time.Now()

	if err := b.storer.Update(ctx, s); err != nil {
		return Scenario{}, fmt.Errorf("update: %w", err)
	}

	if err := b.delegate.Call(ctx, ActionUpdatedData(before, s)); err != nil {
		b.log.Error(ctx, "scenariobus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return s, nil
}

// Delete removes a scenario. ON DELETE CASCADE handles fixture rows.
func (b *Business) Delete(ctx context.Context, s Scenario) error {
	if err := b.storer.Delete(ctx, s); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	if err := b.delegate.Call(ctx, ActionDeletedData(s)); err != nil {
		b.log.Error(ctx, "scenariobus: delegate call failed", "action", ActionDeleted, "err", err)
	}

	return nil
}

// QueryByID retrieves a scenario by its ID.
func (b *Business) QueryByID(ctx context.Context, id uuid.UUID) (Scenario, error) {
	s, err := b.storer.QueryByID(ctx, id)
	if err != nil {
		return Scenario{}, fmt.Errorf("querybyid: %w", err)
	}
	return s, nil
}

// QueryByName retrieves a scenario by its name.
// Used by the seeder for name → ID lookups at runtime.
func (b *Business) QueryByName(ctx context.Context, name string) (Scenario, error) {
	s, err := b.storer.QueryByName(ctx, name)
	if err != nil {
		return Scenario{}, fmt.Errorf("querybyname: %w", err)
	}
	return s, nil
}

// Count returns the number of scenarios matching the filter.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	count, err := b.storer.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}
	return count, nil
}

// Query retrieves a page of scenarios matching the filter.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, pg page.Page) ([]Scenario, error) {
	scenarios, err := b.storer.Query(ctx, filter, orderBy, pg)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	return scenarios, nil
}

// Active returns the currently active scenario or ErrNotFound if none is set.
func (b *Business) Active(ctx context.Context) (Scenario, error) {
	activeID, err := b.storer.QueryActive(ctx)
	if err != nil {
		return Scenario{}, fmt.Errorf("queryactive: %w", err)
	}
	if activeID == uuid.Nil {
		return Scenario{}, ErrNotFound
	}

	s, err := b.storer.QueryByID(ctx, activeID)
	if err != nil {
		return Scenario{}, fmt.Errorf("querybyid active: %w", err)
	}
	return s, nil
}

// SetActive updates the scenarios_active singleton to point at the given ID.
// The ID must exist in the scenarios table.
func (b *Business) SetActive(ctx context.Context, id uuid.UUID) error {
	// Verify the scenario exists before setting it active.
	if _, err := b.storer.QueryByID(ctx, id); err != nil {
		return fmt.Errorf("setactive querybyid: %w", err)
	}

	if err := b.storer.SetActive(ctx, id); err != nil {
		return fmt.Errorf("setactive: %w", err)
	}
	return nil
}

// Load executes a full scenario swap inside a single database transaction:
//  1. Reads the current active scenario from scenarios_active.
//  2. Deletes all scoped rows from the 18 floor-scoped tables for the
//     current active scenario (if one is set).
//  3. Inserts fixture rows for the target scenario via ApplyFixtures.
//  4. Updates scenarios_active to the target scenario.
//
// Delegate events are NOT fired from Load — this is a bulk mutation that
// bypasses individual workflow automation triggers by design. The operation
// is atomic: either all 18 tables are reset and the new fixtures applied,
// or nothing changes.
func (b *Business) Load(ctx context.Context, id uuid.UUID) error {
	// Verify the target scenario exists before opening a transaction.
	if _, err := b.storer.QueryByID(ctx, id); err != nil {
		return fmt.Errorf("load querybyid: %w", err)
	}

	tx, err := b.beginner.Begin()
	if err != nil {
		return fmt.Errorf("load begin tx: %w", err)
	}

	txBus, err := b.NewWithTx(tx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("load newwithtx: %w", err)
	}

	// Read current active (may be uuid.Nil if fresh/cleared).
	currentActive, err := txBus.storer.QueryActive(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("load queryactive: %w", err)
	}

	// Delete scoped rows for the current active scenario.
	if currentActive != uuid.Nil {
		if err := txBus.storer.DeleteScopedRows(ctx, currentActive); err != nil {
			tx.Rollback()
			return fmt.Errorf("load delete scoped rows: %w", err)
		}
	}

	// Apply fixture rows for the target scenario.
	if err := txBus.storer.ApplyFixtures(ctx, id); err != nil {
		tx.Rollback()
		return fmt.Errorf("load applyfixtures: %w", err)
	}

	// Update the active pointer.
	if err := txBus.storer.SetActive(ctx, id); err != nil {
		tx.Rollback()
		return fmt.Errorf("load setactive: %w", err)
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return fmt.Errorf("load commit: %w", err)
	}

	return nil
}

// Reset re-applies the current active scenario (idempotent re-seed).
// Returns ErrNoActiveScenario if scenarios_active.scenario_id IS NULL.
//
// Delegate events are NOT fired from Reset — same reasoning as Load:
// this is a bulk mutation; individual workflow triggers do not apply.
func (b *Business) Reset(ctx context.Context) error {
	// Read active outside a TX first to give a clean error before acquiring TX.
	activeID, err := b.storer.QueryActive(ctx)
	if err != nil {
		return fmt.Errorf("reset queryactive: %w", err)
	}
	if activeID == uuid.Nil {
		return ErrNoActiveScenario
	}

	// Re-apply via Load (opens its own TX, deletes then re-inserts fixtures).
	if err := b.Load(ctx, activeID); err != nil {
		return fmt.Errorf("reset: %w", err)
	}

	return nil
}

