package productdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/productbus"
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (productbus.Storer, error) {
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

// Create inserts a new product into the database.
func (s *Store) Create(ctx context.Context, brand productbus.Product) error {
	const q = `
    INSERT INTO products ( 
		id, sku, brand_id, category_id, name, description, model_number, upc_code, status, 
		is_active, is_perishable, handling_instructions, units_per_case, created_date, updated_date
    ) VALUES (
		:id, :sku, :brand_id, :category_id, :name, :description, :model_number, :upc_code, :status, 
		:is_active, :is_perishable, :handling_instructions, :units_per_case, :created_date, :updated_date
	)
    `

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBProduct(brand)); err != nil {
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", productbus.ErrForeignKeyViolation)
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", productbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

// Update replaces a product document in the database.
func (s *Store) Update(ctx context.Context, prod productbus.Product) error {
	const q = `
	UPDATE
		products
	SET
		id = :id,
		sku = :sku,
		brand_id = :brand_id,
		category_id = :category_id,
		name = :name,
		description = :description,
		model_number = :model_number,
		upc_code = :upc_code,
		status = :status,
		is_active = :is_active,
		is_perishable = :is_perishable,
		handling_instructions = :handling_instructions,
		units_per_case = :units_per_case,
		updated_date = :updated_date
	WHERE
		id = :id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBProduct(prod)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", productbus.ErrUniqueEntry)
		}
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", productbus.ErrForeignKeyViolation)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

// Delete removes an product from the database.
func (s *Store) Delete(ctx context.Context, product productbus.Product) error {
	const q = `
	DELETE FROM
		products
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBProduct(product)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of products from the database.
func (s *Store) Query(ctx context.Context, filter productbus.QueryFilter, orderBy order.By, page page.Page) ([]productbus.Product, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
    SELECT
		id, sku, brand_id, category_id, name, description, model_number, upc_code, status, 
		is_active, is_perishable, handling_instructions, units_per_case, created_date, updated_date
    FROM
        products`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var prod []product
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &prod); err != nil {
		return nil, fmt.Errorf("namedselectcontext: %w", err)
	}

	return toBusProducts(prod), nil
}

// Count returns the number of assets in the database.
func (s *Store) Count(ctx context.Context, filter productbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
    SELECT
        COUNT(1) AS count
    FROM
        products`

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
func (s *Store) QueryByID(ctx context.Context, userBrandID uuid.UUID) (productbus.Product, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: userBrandID.String(),
	}

	const q = `
    SELECT
       	id, sku, brand_id, category_id, name, description, model_number, upc_code, status, 
		is_active, is_perishable, handling_instructions, units_per_case, created_date, updated_date
    FROM
        products
    WHERE
        id = :id
    `
	var ci product

	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &ci); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return productbus.Product{}, productbus.ErrNotFound
		}
		return productbus.Product{}, fmt.Errorf("querystruct: %w", err)
	}

	return toBusProduct(ci), nil
}
