package purchaseorderdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for purchase order database access.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (purchaseorderbus.Storer, error) {
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

// Create adds a new purchase order to the system.
func (s *Store) Create(ctx context.Context, po purchaseorderbus.PurchaseOrder) error {
	const q = `
	INSERT INTO procurement.purchase_orders (
		id, order_number, supplier_id, purchase_order_status_id,
		delivery_warehouse_id, delivery_location_id, delivery_street_id,
		order_date, expected_delivery_date, actual_delivery_date,
		subtotal, tax_amount, shipping_cost, total_amount, currency_id,
		requested_by, approved_by, approved_date,
		notes, supplier_reference_number,
		created_by, updated_by, created_date, updated_date
	) VALUES (
		:id, :order_number, :supplier_id, :purchase_order_status_id,
		:delivery_warehouse_id, :delivery_location_id, :delivery_street_id,
		:order_date, :expected_delivery_date, :actual_delivery_date,
		:subtotal, :tax_amount, :shipping_cost, :total_amount, :currency_id,
		:requested_by, :approved_by, :approved_date,
		:notes, :supplier_reference_number,
		:created_by, :updated_by, :created_date, :updated_date
	)
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBPurchaseOrder(po)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", purchaseorderbus.ErrUnique)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update modifies a purchase order in the system.
func (s *Store) Update(ctx context.Context, po purchaseorderbus.PurchaseOrder) error {
	const q = `
	UPDATE
		procurement.purchase_orders
	SET
		order_number = :order_number,
		supplier_id = :supplier_id,
		purchase_order_status_id = :purchase_order_status_id,
		delivery_warehouse_id = :delivery_warehouse_id,
		delivery_location_id = :delivery_location_id,
		delivery_street_id = :delivery_street_id,
		order_date = :order_date,
		expected_delivery_date = :expected_delivery_date,
		actual_delivery_date = :actual_delivery_date,
		subtotal = :subtotal,
		tax_amount = :tax_amount,
		shipping_cost = :shipping_cost,
		total_amount = :total_amount,
		currency_id = :currency_id,
		approved_by = :approved_by,
		approved_date = :approved_date,
		notes = :notes,
		supplier_reference_number = :supplier_reference_number,
		updated_by = :updated_by,
		updated_date = :updated_date
	WHERE
		id = :id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBPurchaseOrder(po)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", purchaseorderbus.ErrUnique)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes a purchase order from the system.
func (s *Store) Delete(ctx context.Context, po purchaseorderbus.PurchaseOrder) error {
	const q = `
	DELETE FROM
		procurement.purchase_orders
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBPurchaseOrder(po)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of purchase orders from the system.
func (s *Store) Query(ctx context.Context, filter purchaseorderbus.QueryFilter, orderBy order.By, page page.Page) ([]purchaseorderbus.PurchaseOrder, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
		id, order_number, supplier_id, purchase_order_status_id,
		delivery_warehouse_id, delivery_location_id, delivery_street_id,
		order_date, expected_delivery_date, actual_delivery_date,
		subtotal, tax_amount, shipping_cost, total_amount, currency_id,
		requested_by, approved_by, approved_date,
		notes, supplier_reference_number,
		created_by, updated_by, created_date, updated_date
	FROM
		procurement.purchase_orders`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbOrders []purchaseOrder
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbOrders); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusPurchaseOrders(dbOrders), nil
}

// Count returns the total number of purchase orders in the DB.
func (s *Store) Count(ctx context.Context, filter purchaseorderbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
	SELECT
		COUNT(1) AS count
	FROM
		procurement.purchase_orders`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("db: %w", err)
	}

	return count.Count, nil
}

// QueryByID retrieves a single purchase order from the system by its ID.
func (s *Store) QueryByID(ctx context.Context, poID uuid.UUID) (purchaseorderbus.PurchaseOrder, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: poID.String(),
	}

	const q = `
	SELECT
		id, order_number, supplier_id, purchase_order_status_id,
		delivery_warehouse_id, delivery_location_id, delivery_street_id,
		order_date, expected_delivery_date, actual_delivery_date,
		subtotal, tax_amount, shipping_cost, total_amount, currency_id,
		requested_by, approved_by, approved_date,
		notes, supplier_reference_number,
		created_by, updated_by, created_date, updated_date
	FROM
		procurement.purchase_orders
	WHERE
		id = :id`

	var dbOrder purchaseOrder
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbOrder); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return purchaseorderbus.PurchaseOrder{}, fmt.Errorf("db: %w", purchaseorderbus.ErrNotFound)
		}
		return purchaseorderbus.PurchaseOrder{}, fmt.Errorf("db: %w", err)
	}

	return toBusPurchaseOrder(dbOrder), nil
}

// QueryByIDs retrieves a list of purchase orders from the system by their IDs.
func (s *Store) QueryByIDs(ctx context.Context, poIDs []uuid.UUID) ([]purchaseorderbus.PurchaseOrder, error) {
	uuidStrings := make([]string, len(poIDs))
	for i, id := range poIDs {
		uuidStrings[i] = id.String()
	}

	data := struct {
		OrderIDs []string `db:"order_ids"`
	}{
		OrderIDs: uuidStrings,
	}

	const q = `
	SELECT
		id, order_number, supplier_id, purchase_order_status_id,
		delivery_warehouse_id, delivery_location_id, delivery_street_id,
		order_date, expected_delivery_date, actual_delivery_date,
		subtotal, tax_amount, shipping_cost, total_amount, currency_id,
		requested_by, approved_by, approved_date,
		notes, supplier_reference_number,
		created_by, updated_by, created_date, updated_date
	FROM
		procurement.purchase_orders
	WHERE
		id IN (:order_ids)`

	var dbOrders []purchaseOrder
	if err := sqldb.NamedQuerySliceUsingIn(ctx, s.log, s.db, q, data, &dbOrders); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusPurchaseOrders(dbOrders), nil
}