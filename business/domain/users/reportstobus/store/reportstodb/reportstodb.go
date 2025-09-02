package reportstodb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/users/reportstobus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for reports to database access.
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

// NewWithTx constructs a new Store value replacing the sqlx DB
// value with a sqlx DB value that is currently inside a transaction.
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (reportstobus.Storer, error) {
	ec, err := sqldb.GetExtContext(tx)
	if err != nil {
		return nil, err
	}

	store := Store{
		log: s.log,
		db:  ec,
	}

	return &store, nil
}

// Create inserts a new reports to into the database.
func (s *Store) Create(ctx context.Context, t reportstobus.ReportsTo) error {
	const q = `
    INSERT INTO hr.reports_to (
        id, boss_id, reporter_id
    ) VALUES (
        :id, :boss_id, :reporter_id
    )
    `
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBReportsTo(t)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", reportstobus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update modifies data about an reports to in the database.
func (s *Store) Update(ctx context.Context, t reportstobus.ReportsTo) error {
	const q = `
    UPDATE 
        hr.reports_to
    SET
        boss_id = :boss_id,
        reporter_id = :reporter_id
    WHERE
        id = :id
    `
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBReportsTo(t)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", reportstobus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes an tag from the database.
func (s *Store) Delete(ctx context.Context, at reportstobus.ReportsTo) error {
	const q = `
    DELETE FROM
        hr.reports_to
    WHERE
        id = :id
    `
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBReportsTo(at)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of existing hr.reports_to from the database.
func (s *Store) Query(ctx context.Context, filter reportstobus.QueryFilter, orderBy order.By, page page.Page) ([]reportstobus.ReportsTo, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
    SELECT
        id, boss_id, reporter_id
    FROM
        hr.reports_to `

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbRTs []reportsTo
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbRTs); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusReportsTos(dbRTs), nil
}

// Count returns the total number of hr.reports_to in the DB.
func (s *Store) Count(ctx context.Context, filter reportstobus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
    SELECT
        COUNT(1) AS count
    FROM
        hr.reports_to`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("namedquerysingle: %w", err)
	}

	return count.Count, nil
}

// QueryByID retrieves a single tag by its id.
func (s *Store) QueryByID(ctx context.Context, id uuid.UUID) (reportstobus.ReportsTo, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: id.String(),
	}

	const q = `
    SELECT
        id, boss_id, reporter_id
    FROM
        hr.reports_to
    WHERE
        id = :id
    `

	var dbRT reportsTo
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbRT); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return reportstobus.ReportsTo{}, fmt.Errorf("db: %w", reportstobus.ErrNotFound)
		}
		return reportstobus.ReportsTo{}, fmt.Errorf("querystruct: %w", err)
	}

	return toBusReportsTo(dbRT), nil
}
