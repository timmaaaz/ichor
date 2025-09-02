package supplierproductdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/supplier/supplierproductbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for supplier-product database access.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (supplierproductbus.Storer, error) {
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

func (s *Store) Create(ctx context.Context, ch supplierproductbus.SupplierProduct) error {
	const q = `
    INSERT INTO procurement.supplier_products ( 
		id, supplier_id, product_id, supplier_part_number, min_order_quantity, max_order_quantity, 
		lead_time_days, unit_cost, is_primary_supplier, created_date, updated_date
    ) VALUES (
		:id, :supplier_id, :product_id, :supplier_part_number, :min_order_quantity, :max_order_quantity, 
		:lead_time_days, :unit_cost, :is_primary_supplier, :created_date, :updated_date
    )
    `

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBSupplierProduct(ch)); err != nil {
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", supplierproductbus.ErrForeignKeyViolation)
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", supplierproductbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

func (s *Store) Update(ctx context.Context, ch supplierproductbus.SupplierProduct) error {
	const q = `
    UPDATE
        procurement.supplier_products
    SET
		id = :id,
		supplier_id = :supplier_id,
		product_id = :product_id,
		supplier_part_number = :supplier_part_number,
		min_order_quantity = :min_order_quantity,
		max_order_quantity = :max_order_quantity,
		lead_time_days = :lead_time_days,
		unit_cost = :unit_cost,
		is_primary_supplier = :is_primary_supplier,
		created_date = :created_date,
		updated_date = :updated_date
    WHERE
        id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBSupplierProduct(ch)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", supplierproductbus.ErrUniqueEntry)
		}
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", supplierproductbus.ErrForeignKeyViolation)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

func (s *Store) Delete(ctx context.Context, ch supplierproductbus.SupplierProduct) error {
	const q = `
	DELETE FROM
		procurement.supplier_products
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBSupplierProduct(ch)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

func (s *Store) Query(ctx context.Context, filter supplierproductbus.QueryFilter, orderBy order.By, page page.Page) ([]supplierproductbus.SupplierProduct, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
		id, supplier_id, product_id, supplier_part_number, min_order_quantity, max_order_quantity, 
		lead_time_days, unit_cost, is_primary_supplier, created_date, updated_date
	FROM 
		procurement.supplier_products
	`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var sp []supplierProduct
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &sp); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusSupplierProducts(sp)
}

func (s *Store) Count(ctx context.Context, filter supplierproductbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
    SELECT
        COUNT(1) AS count
    FROM
        procurement.supplier_products`

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

func (s *Store) QueryByID(ctx context.Context, supplierProductID uuid.UUID) (supplierproductbus.SupplierProduct, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: supplierProductID.String(),
	}

	const q = `
	SELECT
		id, supplier_id, product_id, supplier_part_number, min_order_quantity, max_order_quantity, 
		lead_time_days, unit_cost, is_primary_supplier, created_date, updated_date
	FROM 
		procurement.supplier_products
	WHERE 
	    id = :id`

	var sp supplierProduct
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &sp); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return supplierproductbus.SupplierProduct{}, supplierproductbus.ErrNotFound
		}
		return supplierproductbus.SupplierProduct{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusSupplierProduct(sp)
}
