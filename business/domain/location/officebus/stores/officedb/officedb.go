package officedb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/location/officebus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for streets database access.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (officebus.Storer, error) {
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

// Create inserts a new office into the database.
func (s *Store) Create(ctx context.Context, at officebus.Office) error {
	const q = `
    INSERT INTO hr.offices (
        id, name, street_id
    ) VALUES (
        :id, :name, :street_id
    )
    `
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBOffice(at)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", officebus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update modifies data about an office in the database.
func (s *Store) Update(ctx context.Context, at officebus.Office) error {
	const q = `
    UPDATE 
        hr.offices
    SET
        name = :name,
        street_id = :street_id
    WHERE
        id = :id
    `
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBOffice(at)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", officebus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes an office from the database.
func (s *Store) Delete(ctx context.Context, at officebus.Office) error {
	const q = `
    DELETE FROM
        hr.offices
    WHERE
        id = :id
    `
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBOffice(at)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of existing offices from the database.
func (s *Store) Query(ctx context.Context, filter officebus.QueryFilter, orderBy order.By, page page.Page) ([]officebus.Office, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
    SELECT
        id, name, street_id
    FROM
        hr.offices`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbOs []office
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbOs); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusOffices(dbOs), nil
}

// Count returns the total number of offices in the DB.
func (s *Store) Count(ctx context.Context, filter officebus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
    SELECT
        COUNT(1) AS count
    FROM
        hr.offices`

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

// QueryByID retrieves a single office by its id.
func (s *Store) QueryByID(ctx context.Context, id uuid.UUID) (officebus.Office, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: id.String(),
	}

	const q = `
    SELECT
        id, name, street_id
    FROM
        hr.offices
    WHERE
        id = :id
    `

	var dbO office
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbO); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return officebus.Office{}, fmt.Errorf("db: %w", officebus.ErrNotFound)
		}
		return officebus.Office{}, fmt.Errorf("querystruct: %w", err)
	}

	return toBusOffice(dbO), nil
}
