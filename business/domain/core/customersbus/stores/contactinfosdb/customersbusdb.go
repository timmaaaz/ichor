package customersdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/customersbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"

	"github.com/jmoiron/sqlx"
)

// Store manages the set of APIs for assets database access.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (customersbus.Storer, error) {
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

// Create inserts a new user asset into the database.
func (s *Store) Create(ctx context.Context, ass customersbus.Customers) error {
	const q = `
    INSERT INTO customers (
		id, name, contact_id, delivery_address_id, notes, created_by, updated_by, created_date, updated_date
    ) VALUES (
		:id, :name, :contact_id, :delivery_address_id, :notes, :created_by, :updated_by, :created_date, :updated_date
	)
    `

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBCustomers(ass)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", customersbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

// Update replaces a user asset document in the database.
func (s *Store) Update(ctx context.Context, ass customersbus.Customers) error {
	const q = `
	UPDATE
		customers
	SET
		id = :id,
		name = :name,
        contact_id = :contact_id,
        delivery_address_id = :delivery_address_id,
        notes = :notes,
        created_by = :created_by,
        updated_by = :updated_by,
        created_date = :created_date,
        updated_date = :updated_date
	WHERE
		id = :id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBCustomers(ass)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", customersbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

// Delete removes an user asset from the database.
func (s *Store) Delete(ctx context.Context, ass customersbus.Customers) error {
	const q = `
	DELETE FROM
		customers
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBCustomers(ass)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of user assets from the database.
func (s *Store) Query(ctx context.Context, filter customersbus.QueryFilter, orderBy order.By, page page.Page) ([]customersbus.Customers, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
    SELECT
		id, name, contact_id, delivery_address_id, notes, created_by, updated_by, created_date, updated_date
    FROM
        customers`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var ci []customers
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &ci); err != nil {
		return nil, fmt.Errorf("namedselectcontext: %w", err)
	}

	return toBusCustomerss(ci), nil
}

// Count returns the number of assets in the database.
func (s *Store) Count(ctx context.Context, filter customersbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
    SELECT
        COUNT(1) AS count
    FROM
        customers`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("namedqueryint: %w", err)
	}

	return count.Count, nil
}

// QueryByID retrieves a single asset from the database by its ID.
func (s *Store) QueryByID(ctx context.Context, userCustomersID uuid.UUID) (customersbus.Customers, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: userCustomersID.String(),
	}

	const q = `
    SELECT
        id, name, contact_id, delivery_address_id, notes, created_by, updated_by, created_date, updated_date
    FROM
        customers
    WHERE
        id = :id
    `
	var ci customers

	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &ci); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return customersbus.Customers{}, customersbus.ErrNotFound
		}
		return customersbus.Customers{}, fmt.Errorf("querystruct: %w", err)
	}

	return toBusCustomers(ci), nil
}
