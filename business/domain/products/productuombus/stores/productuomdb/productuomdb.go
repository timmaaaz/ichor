package productuomdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/products/productuombus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for product_uoms database access.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (productuombus.Storer, error) {
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

// Create inserts a new product UOM into the database.
func (s *Store) Create(ctx context.Context, uom productuombus.ProductUOM) error {
	const q = `
    INSERT INTO products.product_uoms
        (id, product_id, name, abbreviation, conversion_factor, is_base, is_approximate, notes, created_date, updated_date)
    VALUES
        (:id, :product_id, :name, :abbreviation, :conversion_factor, :is_base, :is_approximate, :notes, :created_date, :updated_date)`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBProductUOM(uom)); err != nil {
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", productuombus.ErrForeignKeyViolation)
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", productuombus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

// Update modifies a product UOM in the database.
func (s *Store) Update(ctx context.Context, uom productuombus.ProductUOM) error {
	const q = `
    UPDATE products.product_uoms
    SET
        name              = :name,
        abbreviation      = :abbreviation,
        conversion_factor = :conversion_factor,
        is_base           = :is_base,
        is_approximate    = :is_approximate,
        notes             = :notes,
        updated_date      = :updated_date
    WHERE
        id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBProductUOM(uom)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", productuombus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

// Delete removes a product UOM from the database.
func (s *Store) Delete(ctx context.Context, uom productuombus.ProductUOM) error {
	const q = `DELETE FROM products.product_uoms WHERE id = :id`

	data := struct {
		ID uuid.UUID `db:"id"`
	}{ID: uom.ID}

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

// Count returns the number of product UOMs in the store.
func (s *Store) Count(ctx context.Context, filter productuombus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `SELECT count(1) AS count FROM products.product_uoms`

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

// Query retrieves a list of product UOMs from the database.
func (s *Store) Query(ctx context.Context, filter productuombus.QueryFilter, orderBy order.By, page page.Page) ([]productuombus.ProductUOM, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `SELECT * FROM products.product_uoms`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}
	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbs []productUOM
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbs); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}
	return toBusProductUOMs(dbs), nil
}

// QueryByID retrieves a single product UOM by its ID.
func (s *Store) QueryByID(ctx context.Context, uomID uuid.UUID) (productuombus.ProductUOM, error) {
	data := struct {
		ID uuid.UUID `db:"id"`
	}{ID: uomID}

	const q = `SELECT * FROM products.product_uoms WHERE id = :id`

	var db productUOM
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &db); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return productuombus.ProductUOM{}, fmt.Errorf("namedquerystruct: %w", productuombus.ErrNotFound)
		}
		return productuombus.ProductUOM{}, fmt.Errorf("namedquerystruct: %w", err)
	}
	return toBusProductUOM(db), nil
}
