// Package scenariodb provides the Postgres implementation of scenariobus.Storer.
package scenariodb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/scenarios/scenariobus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/business/sdk/sqldb/dbarray"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for scenario database access.
type Store struct {
	log *logger.Logger
	db  sqlx.ExtContext
}

// NewStore constructs the api for data access.
func NewStore(log *logger.Logger, db *sqlx.DB) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

// NewWithTx constructs a new Store value replacing the sqlx DB value with one
// currently inside a transaction.
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (scenariobus.Storer, error) {
	ec, err := sqldb.GetExtContext(tx)
	if err != nil {
		return nil, err
	}

	return &Store{
		log: s.log,
		db:  ec,
	}, nil
}

// =============================================================================
// scenarios table
// =============================================================================

// Create inserts a new scenario row.
func (s *Store) Create(ctx context.Context, sc scenariobus.Scenario) error {
	const q = `
	INSERT INTO inventory.scenarios (
		id, name, description, created_date, updated_date
	) VALUES (
		:id, :name, :description, :created_date, :updated_date
	)`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBScenario(sc)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", scenariobus.ErrUniqueName)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

// Update replaces an existing scenario row.
func (s *Store) Update(ctx context.Context, sc scenariobus.Scenario) error {
	const q = `
	UPDATE
		inventory.scenarios
	SET
		name         = :name,
		description  = :description,
		updated_date = :updated_date
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBScenario(sc)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", scenariobus.ErrUniqueName)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

// Delete removes a scenario row. ON DELETE CASCADE handles fixture rows.
func (s *Store) Delete(ctx context.Context, sc scenariobus.Scenario) error {
	const q = `
	DELETE FROM
		inventory.scenarios
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBScenario(sc)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

// Query retrieves a page of scenario rows.
func (s *Store) Query(ctx context.Context, filter scenariobus.QueryFilter, orderBy order.By, pg page.Page) ([]scenariobus.Scenario, error) {
	data := map[string]any{
		"offset":        (pg.Number() - 1) * pg.RowsPerPage(),
		"rows_per_page": pg.RowsPerPage(),
	}

	const q = `
	SELECT
		id, name, description, created_date, updated_date
	FROM
		inventory.scenarios`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var rows []dbScenario
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &rows); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusScenarios(rows), nil
}

// Count returns the number of scenario rows matching the filter.
func (s *Store) Count(ctx context.Context, filter scenariobus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
	SELECT
		COUNT(1) AS count
	FROM
		inventory.scenarios`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("namedquerystruct: %w", err)
	}

	return count.Count, nil
}

// QueryByID retrieves a single scenario row by ID.
func (s *Store) QueryByID(ctx context.Context, id uuid.UUID) (scenariobus.Scenario, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: id.String(),
	}

	const q = `
	SELECT
		id, name, description, created_date, updated_date
	FROM
		inventory.scenarios
	WHERE
		id = :id`

	var row dbScenario
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &row); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return scenariobus.Scenario{}, scenariobus.ErrNotFound
		}
		return scenariobus.Scenario{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusScenario(row), nil
}

// QueryByName retrieves a single scenario row by name.
func (s *Store) QueryByName(ctx context.Context, name string) (scenariobus.Scenario, error) {
	data := struct {
		Name string `db:"name"`
	}{
		Name: name,
	}

	const q = `
	SELECT
		id, name, description, created_date, updated_date
	FROM
		inventory.scenarios
	WHERE
		name = :name`

	var row dbScenario
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &row); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return scenariobus.Scenario{}, scenariobus.ErrNotFound
		}
		return scenariobus.Scenario{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusScenario(row), nil
}

// =============================================================================
// scenario_fixtures table
// =============================================================================

// CreateFixture inserts a new scenario fixture row.
func (s *Store) CreateFixture(ctx context.Context, f scenariobus.ScenarioFixture) error {
	const q = `
	INSERT INTO inventory.scenario_fixtures (
		id, scenario_id, target_table, payload_json, created_date
	) VALUES (
		:id, :scenario_id, :target_table, :payload_json, :created_date
	)`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBScenarioFixture(f)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

// QueryFixturesByScenario returns all fixture rows for the given scenario.
func (s *Store) QueryFixturesByScenario(ctx context.Context, scenarioID uuid.UUID) ([]scenariobus.ScenarioFixture, error) {
	data := struct {
		ScenarioID string `db:"scenario_id"`
	}{
		ScenarioID: scenarioID.String(),
	}

	const q = `
	SELECT
		id, scenario_id, target_table, payload_json, created_date
	FROM
		inventory.scenario_fixtures
	WHERE
		scenario_id = :scenario_id
	ORDER BY
		target_table, created_date`

	var rows []dbScenarioFixture
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &rows); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusScenarioFixtures(rows), nil
}

// =============================================================================
// scenarios_active table (singleton)
// =============================================================================

// QueryActive returns the active scenario ID from the singleton row.
// Returns uuid.Nil if the singleton row does not exist or scenario_id IS NULL.
func (s *Store) QueryActive(ctx context.Context) (uuid.UUID, error) {
	const q = `
	SELECT
		scenario_id
	FROM
		inventory.scenarios_active
	WHERE
		singleton_key = 0`

	var row dbScenariosActive
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, struct{}{}, &row); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			// No singleton row yet — no active scenario.
			return uuid.Nil, nil
		}
		return uuid.Nil, fmt.Errorf("namedquerystruct: %w", err)
	}

	if row.ScenarioID == nil {
		return uuid.Nil, nil
	}
	return *row.ScenarioID, nil
}

// SetActive UPSERTs the singleton scenarios_active row to point at id.
func (s *Store) SetActive(ctx context.Context, id uuid.UUID) error {
	data := struct {
		ScenarioID string `db:"scenario_id"`
	}{
		ScenarioID: id.String(),
	}

	const q = `
	INSERT INTO inventory.scenarios_active (singleton_key, scenario_id, updated_date)
	VALUES (0, :scenario_id, NOW())
	ON CONFLICT (singleton_key)
	DO UPDATE SET
		scenario_id  = EXCLUDED.scenario_id,
		updated_date = NOW()`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

// =============================================================================
// Bulk load/reset helpers
// =============================================================================

// scopedTables is the ordered list of the 18 floor-scoped tables that carry
// a scenario_id column (migration 2.35). Order matters for FK constraints —
// more dependent child tables are listed before their parents.
var scopedTables = []string{
	"workflow.approval_requests",
	"inventory.cycle_count_items",
	"inventory.cycle_count_sessions",
	"inventory.quality_inspections",
	"inventory.put_away_tasks",
	"inventory.pick_tasks",
	"inventory.serial_numbers",
	"inventory.lot_locations",
	"inventory.lot_trackings",
	"inventory.inventory_adjustments",
	"inventory.inventory_transactions",
	"inventory.transfer_orders",
	"procurement.purchase_order_line_items",
	"procurement.purchase_orders",
	"sales.order_fulfillment_statuses",
	"sales.order_line_items",
	"sales.orders",
	"inventory.inventory_items",
}

// DeleteScopedRows removes all rows with the given scenario_id from every
// one of the 18 floor-scoped tables. Called inside a transaction by Load.
func (s *Store) DeleteScopedRows(ctx context.Context, scenarioID uuid.UUID) error {
	data := map[string]any{
		"scenario_id": scenarioID,
	}

	for _, table := range scopedTables {
		q := fmt.Sprintf(`DELETE FROM %s WHERE scenario_id = :scenario_id`, table)
		if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
			return fmt.Errorf("deletescopedrows %s: %w", table, err)
		}
	}
	return nil
}

// ApplyFixtures inserts rows from scenario_fixtures into their target tables
// using jsonb_populate_record so column defaults are honoured and the Go code
// remains table-agnostic. Each target_table in the fixture set is handled with
// a single INSERT … SELECT statement.
func (s *Store) ApplyFixtures(ctx context.Context, target uuid.UUID) error {
	// Get the distinct target tables for this scenario.
	tableData := struct {
		ScenarioID string `db:"scenario_id"`
	}{
		ScenarioID: target.String(),
	}

	const distinctQ = `
	SELECT DISTINCT target_table
	FROM inventory.scenario_fixtures
	WHERE scenario_id = :scenario_id`

	var tableRows []struct {
		TargetTable string `db:"target_table"`
	}
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, distinctQ, tableData, &tableRows); err != nil {
		return fmt.Errorf("applyfixtures distinct tables: %w", err)
	}

	for _, tr := range tableRows {
		// INSERT … SELECT per target table. The payload_json column holds JSONB;
		// jsonb_populate_record maps each row to the target table's columns.
		// Named params :scenario_id and :table_name are bound by sqlx.
		//
		// CAST(NULL AS %s) is used instead of NULL::%s because sqlx's named-
		// parameter parser treats "::" as ":" + ":name" and rejects the cast
		// with "syntax error at or near ':'". CAST(...) is semantically
		// identical and avoids the collision.
		q := fmt.Sprintf(`
		INSERT INTO %s
		SELECT (jsonb_populate_record(CAST(NULL AS %s), payload_json)).*
		FROM inventory.scenario_fixtures
		WHERE scenario_id = :scenario_id AND target_table = :table_name`,
			tr.TargetTable, tr.TargetTable)

		data := map[string]any{
			"scenario_id": target,
			"table_name":  tr.TargetTable,
		}
		if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
			return fmt.Errorf("applyfixtures insert %s: %w", tr.TargetTable, err)
		}
	}

	return nil
}

// =============================================================================
// Lever overrides + worker zones
// =============================================================================

// ApplyLeverOverrides upserts rows into config.scenario_setting_overrides.
// Called at seed time to persist per-scenario config-lever values. The value
// column is TEXT so no JSON marshalling is required.
func (s *Store) ApplyLeverOverrides(ctx context.Context, overrides []scenariobus.SettingOverride) error {
	if len(overrides) == 0 {
		return nil
	}

	const q = `
	INSERT INTO config.scenario_setting_overrides
	    (scenario_id, key, value)
	VALUES
	    (:scenario_id, :key, :value)
	ON CONFLICT (scenario_id, key) DO UPDATE
	    SET value = EXCLUDED.value`

	for _, o := range overrides {
		row := struct {
			ScenarioID string `db:"scenario_id"`
			Key        string `db:"key"`
			Value      string `db:"value"`
		}{
			ScenarioID: o.ScenarioID.String(),
			Key:        o.Key,
			Value:      o.Value,
		}
		if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, row); err != nil {
			return fmt.Errorf("apply lever override %s: %w", o.Key, err)
		}
	}
	return nil
}

// ApplyWorkerZones updates core.users.assigned_zones for each binding.
// Called inside the Load transaction so zone assignments are atomic with
// fixture application. Returns an error if a username is not found.
func (s *Store) ApplyWorkerZones(ctx context.Context, zones []scenariobus.WorkerZoneBinding) error {
	if len(zones) == 0 {
		return nil
	}

	const q = `
	UPDATE core.users
	SET    assigned_zones = :zones
	WHERE  username = :username`

	for _, z := range zones {
		row := struct {
			Zones    dbarray.String `db:"zones"`
			Username string         `db:"username"`
		}{
			Zones:    dbarray.String(z.Zones),
			Username: z.Username,
		}
		n, err := sqldb.NamedExecContextWithCount(ctx, s.log, s.db, q, row)
		if err != nil {
			return fmt.Errorf("apply worker zones %s: %w", z.Username, err)
		}
		if n == 0 {
			return fmt.Errorf("apply worker zones: username %q not found", z.Username)
		}
	}
	return nil
}
