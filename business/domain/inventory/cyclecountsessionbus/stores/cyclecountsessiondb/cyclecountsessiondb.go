package cyclecountsessiondb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountsessionbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for cycle count session database access.
type Store struct {
	log *logger.Logger
	db  sqlx.ExtContext
}

// NewStore constructs the API for data access.
func NewStore(log *logger.Logger, db *sqlx.DB) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

// NewWithTx constructs a new Store value replacing the sqlx.DB
// value with a sqlx.Tx value that is currently inside a transaction.
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (cyclecountsessionbus.Storer, error) {
	ec, err := sqldb.GetExtContext(tx)
	if err != nil {
		return nil, err
	}

	return &Store{
		log: s.log,
		db:  ec,
	}, nil
}

// Create inserts a new cycle count session into the database.
func (s *Store) Create(ctx context.Context, session cyclecountsessionbus.CycleCountSession) error {
	const q = `
	INSERT INTO inventory.cycle_count_sessions
		(id, name, status, created_by, created_date, updated_date, completed_date)
	VALUES
		(:id, :name, :status, :created_by, :created_date, :updated_date, :completed_date)
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBCycleCountSession(session)); err != nil {
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", cyclecountsessionbus.ErrForeignKeyViolation)
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", cyclecountsessionbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update modifies an existing cycle count session in the database.
func (s *Store) Update(ctx context.Context, session cyclecountsessionbus.CycleCountSession) error {
	const q = `
	UPDATE inventory.cycle_count_sessions
	SET
		name           = :name,
		status         = :status,
		updated_date   = :updated_date,
		completed_date = :completed_date
	WHERE
		id = :id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBCycleCountSession(session)); err != nil {
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", cyclecountsessionbus.ErrForeignKeyViolation)
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", cyclecountsessionbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes a cycle count session from the database.
func (s *Store) Delete(ctx context.Context, session cyclecountsessionbus.CycleCountSession) error {
	const q = `
	DELETE FROM inventory.cycle_count_sessions
	WHERE id = :id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBCycleCountSession(session)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of cycle count sessions from the database.
func (s *Store) Query(ctx context.Context, filter cyclecountsessionbus.QueryFilter, orderBy order.By, page page.Page) ([]cyclecountsessionbus.CycleCountSession, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
		id, name, status, created_by, created_date, updated_date, completed_date
	FROM
		inventory.cycle_count_sessions
	`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbSessions []cycleCountSession
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbSessions); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	sessions, err := toBusCycleCountSessions(dbSessions)
	if err != nil {
		return nil, fmt.Errorf("tobuscyclecountsessions: %w", err)
	}

	return sessions, nil
}

// Count returns the total number of cycle count sessions matching the filter.
func (s *Store) Count(ctx context.Context, filter cyclecountsessionbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
	SELECT
		COUNT(1) AS count
	FROM
		inventory.cycle_count_sessions
	`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("namedqueryrow: %w", err)
	}

	return count.Count, nil
}

// QueryByID retrieves a single cycle count session by its ID.
func (s *Store) QueryByID(ctx context.Context, sessionID uuid.UUID) (cyclecountsessionbus.CycleCountSession, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: sessionID.String(),
	}

	const q = `
	SELECT
		id, name, status, created_by, created_date, updated_date, completed_date
	FROM
		inventory.cycle_count_sessions
	WHERE
		id = :id
	`

	var dbSession cycleCountSession
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbSession); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return cyclecountsessionbus.CycleCountSession{}, cyclecountsessionbus.ErrNotFound
		}
		return cyclecountsessionbus.CycleCountSession{}, fmt.Errorf("namedqueryrow: %w", err)
	}

	session, err := toBusCycleCountSession(dbSession)
	if err != nil {
		return cyclecountsessionbus.CycleCountSession{}, fmt.Errorf("tobuscyclecountsession: %w", err)
	}

	return session, nil
}
