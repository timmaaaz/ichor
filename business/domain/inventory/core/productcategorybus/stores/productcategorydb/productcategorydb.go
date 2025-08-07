package productcategorydb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/business/domain/inventory/core/productcategorybus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"

	"github.com/jmoiron/sqlx"
)

// Store manages the set of APIs for product category s database access.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (productcategorybus.Storer, error) {
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

// Create inserts a new product category into the database.
func (s *Store) Create(ctx context.Context, pc productcategorybus.ProductCategory) error {
	const q = `
    INSERT INTO product_categories (
        id, name, description, created_date, updated_date
    ) VALUES (
		:id, :name, :description, :created_date, :updated_date
	)
    `

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBProductCategory(pc)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", productcategorybus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

// Update replaces a product category  in the database.
func (s *Store) Update(ctx context.Context, pc productcategorybus.ProductCategory) error {
	const q = `
	UPDATE
		product_categories
	SET
		name = :name,
        description = :description,
        updated_date = :updated_date
	WHERE
		id = :id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBProductCategory(pc)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", productcategorybus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

// Delete removes an product category from the database.
func (s *Store) Delete(ctx context.Context, pc productcategorybus.ProductCategory) error {
	const q = `
	DELETE FROM
		product_categories
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBProductCategory(pc)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of product categories from the database.
func (s *Store) Query(ctx context.Context, filter productcategorybus.QueryFilter, orderBy order.By, page page.Page) ([]productcategorybus.ProductCategory, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
    SELECT
		id, name, description, created_date, updated_date
    FROM
        product_categories`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var ci []productCategory
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &ci); err != nil {
		return nil, fmt.Errorf("namedselectcontext: %w", err)
	}

	return toBusProductCategories(ci), nil
}

// Count returns the number of product categories in the database.
func (s *Store) Count(ctx context.Context, filter productcategorybus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
    SELECT
        COUNT(1) AS count
    FROM
        product_categories`

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

// QueryByID retrieves a single product category  from the database by its ID.
func (s *Store) QueryByID(ctx context.Context, userProductCategoryID uuid.UUID) (productcategorybus.ProductCategory, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: userProductCategoryID.String(),
	}

	const q = `
    SELECT
        id, name, description, created_date, updated_date
    FROM
        product_categories
    WHERE
        id = :id
    `
	var ci productCategory

	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &ci); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return productcategorybus.ProductCategory{}, productcategorybus.ErrNotFound
		}
		return productcategorybus.ProductCategory{}, fmt.Errorf("querystruct: %w", err)
	}

	return toBusProductCategory(ci), nil
}
