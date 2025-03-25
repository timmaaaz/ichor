package supplierdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/supplier/supplierbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (supplierbus.Storer, error) {
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

// Create inserts a new supplier into the database.
func (s *Store) Create(ctx context.Context, supplier supplierbus.Supplier) error {
	const q = `
	INSERT INTO suppliers (
		supplier_id, contact_id, name, payment_terms, lead_time_days, rating, is_active, created_date, updated_date
	) VALUES (
		:supplier_id, :contact_id, :name, :payment_terms, :lead_time_days, :rating, :is_active, :created_date, :updated_date
	)
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBSupplier(supplier)); err != nil {
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", supplierbus.ErrForeignKeyViolation)
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", supplierbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

// Update updates an existing supplier in the database.
func (s *Store) Update(ctx context.Context, supplier supplierbus.Supplier) error {
	const q = `
	UPDATE
		suppliers
	SET
	    supplier_id = :supplier_id,
        contact_id = :contact_id,
        name = :name,
        payment_terms = :payment_terms,
        lead_time_days = :lead_time_days,
		rating = :rating,
        is_active = :is_active,
        updated_date = :updated_date
	WHERE
		supplier_id = :supplier_id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBSupplier(supplier)); err != nil {
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", supplierbus.ErrForeignKeyViolation)
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", supplierbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

// Delete removes a supplier from the database.
func (s *Store) Delete(ctx context.Context, supplier supplierbus.Supplier) error {
	const q = `
    DELETE FROM
        suppliers
    WHERE
        supplier_id = :supplier_id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBSupplier(supplier)); err != nil {
		return fmt.Errorf("execcontext: %w", err)
	}
	return nil
}

// Query retrieves a list of suppliers from the database.
func (s *Store) Query(ctx context.Context, filter supplierbus.QueryFilter, orderBy order.By, page page.Page) ([]supplierbus.Supplier, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
	    supplier_id, contact_id, name, payment_terms, lead_time_days, rating, is_active, created_date, updated_date
	FROM
		suppliers
	`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbSuppliers []supplier

	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbSuppliers); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusSuppliers(dbSuppliers)
}

// Count returns the number of productcosts in the database.
func (s *Store) Count(ctx context.Context, filter supplierbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
    SELECT
        COUNT(1) AS count
    FROM
        suppliers`

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

// QueryByID retrieves a single supplier from the database by its ID.
func (s *Store) QueryByID(ctx context.Context, supplierID uuid.UUID) (supplierbus.Supplier, error) {
	data := struct {
		ID string `db:"supplier_id"`
	}{
		ID: supplierID.String(),
	}

	const q = `
	SELECT
	    supplier_id, contact_id, name, payment_terms, lead_time_days, rating, is_active, created_date, updated_date
	FROM
		suppliers
	WHERE 
		supplier_id = :supplier_id
	`

	var dbSupplier supplier

	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbSupplier); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return supplierbus.Supplier{}, supplierbus.ErrNotFound
		}
		return supplierbus.Supplier{}, fmt.Errorf("querystruct: %w", err)
	}

	return toBusSupplier(dbSupplier)
}
