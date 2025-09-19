package productcostdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/products/productcostbus"
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (productcostbus.Storer, error) {
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

// Create inserts a new product cost into the database.
func (s *Store) Create(ctx context.Context, productcost productcostbus.ProductCost) error {
	const q = `
    INSERT INTO products.product_costs (
        id, product_id, purchase_cost, selling_price, currency, msrp, markup_percentage, landed_cost, carrying_cost,
		abc_classification, depreciation_value, insurance_value, effective_date, created_date, updated_date
    ) VALUES (
		:id, :product_id, :purchase_cost, :selling_price, :currency, :msrp, :markup_percentage, :landed_cost, :carrying_cost,
		:abc_classification, :depreciation_value, :insurance_value, :effective_date, :created_date, :updated_date
	)
    `

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBProductCost(productcost)); err != nil {
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", productcostbus.ErrForeignKeyViolation)
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", productcostbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

// Update replaces a product cost document in the database.
func (s *Store) Update(ctx context.Context, pc productcostbus.ProductCost) error {
	const q = `
	UPDATE
		products.product_costs
	SET
		id = :id,
		product_id = :product_id,
		purchase_cost = :purchase_cost,
		selling_price = :selling_price,
		currency = :currency,
		msrp = :msrp,
		markup_percentage = :markup_percentage,
		landed_cost = :landed_cost,
		carrying_cost = :carrying_cost,
		abc_classification = :abc_classification,
		depreciation_value = :depreciation_value,
		insurance_value = :insurance_value,
		effective_date = :effective_date,
		created_date = :created_date,
		updated_date = :updated_date
	WHERE
		id = :id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBProductCost(pc)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", productcostbus.ErrUniqueEntry)
		}
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", productcostbus.ErrForeignKeyViolation)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

// Delete removes an product cost from the database.
func (s *Store) Delete(ctx context.Context, productCost productcostbus.ProductCost) error {
	const q = `
	DELETE FROM
		products.product_costs
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBProductCost(productCost)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of product costs from the database.
func (s *Store) Query(ctx context.Context, filter productcostbus.QueryFilter, orderBy order.By, page page.Page) ([]productcostbus.ProductCost, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
    SELECT
		id, product_id, purchase_cost, selling_price, currency, msrp, markup_percentage, landed_cost, carrying_cost,
		abc_classification, depreciation_value, insurance_value, effective_date, created_date, updated_date
    FROM
        products.product_costs`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var ci []productCost
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &ci); err != nil {
		return nil, fmt.Errorf("namedselectcontext: %w", err)
	}

	return toBusProductCosts(ci)
}

// Count returns the number of productcosts in the database.
func (s *Store) Count(ctx context.Context, filter productcostbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
    SELECT
        COUNT(1) AS count
    FROM
        products.product_costs`

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
func (s *Store) QueryByID(ctx context.Context, productID uuid.UUID) (productcostbus.ProductCost, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: productID.String(),
	}

	const q = `
    SELECT
        id, product_id, purchase_cost, selling_price, currency, msrp, markup_percentage, landed_cost, carrying_cost,
		abc_classification, depreciation_value, insurance_value, effective_date, created_date, updated_date
    FROM
        products.product_costs
    WHERE
        id = :id
    `
	var ci productCost

	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &ci); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return productcostbus.ProductCost{}, productcostbus.ErrNotFound
		}
		return productcostbus.ProductCost{}, fmt.Errorf("querystruct: %w", err)
	}

	return toBusProductCost(ci)
}
