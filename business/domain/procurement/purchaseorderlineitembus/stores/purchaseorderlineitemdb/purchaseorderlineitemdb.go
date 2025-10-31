package purchaseorderlineitemdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitembus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for purchase order line item database access.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (purchaseorderlineitembus.Storer, error) {
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

// Create adds a new purchase order line item to the system.
func (s *Store) Create(ctx context.Context, poli purchaseorderlineitembus.PurchaseOrderLineItem) error {
	const q = `
	INSERT INTO procurement.purchase_order_line_items (
		id, purchase_order_id, supplier_product_id,
		quantity_ordered, quantity_received, quantity_cancelled,
		unit_cost, discount, line_total, line_item_status_id,
		expected_delivery_date, actual_delivery_date, notes,
		created_by, updated_by, created_date, updated_date
	) VALUES (
		:id, :purchase_order_id, :supplier_product_id,
		:quantity_ordered, :quantity_received, :quantity_cancelled,
		:unit_cost, :discount, :line_total, :line_item_status_id,
		:expected_delivery_date, :actual_delivery_date, :notes,
		:created_by, :updated_by, :created_date, :updated_date
	)
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBPurchaseOrderLineItem(poli)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", purchaseorderlineitembus.ErrUnique)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update modifies a purchase order line item in the system.
func (s *Store) Update(ctx context.Context, poli purchaseorderlineitembus.PurchaseOrderLineItem) error {
	const q = `
	UPDATE
		procurement.purchase_order_line_items
	SET
		supplier_product_id = :supplier_product_id,
		quantity_ordered = :quantity_ordered,
		quantity_received = :quantity_received,
		quantity_cancelled = :quantity_cancelled,
		unit_cost = :unit_cost,
		discount = :discount,
		line_total = :line_total,
		line_item_status_id = :line_item_status_id,
		expected_delivery_date = :expected_delivery_date,
		actual_delivery_date = :actual_delivery_date,
		notes = :notes,
		updated_by = :updated_by,
		updated_date = :updated_date
	WHERE
		id = :id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBPurchaseOrderLineItem(poli)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", purchaseorderlineitembus.ErrUnique)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes a purchase order line item from the system.
func (s *Store) Delete(ctx context.Context, poli purchaseorderlineitembus.PurchaseOrderLineItem) error {
	const q = `
	DELETE FROM
		procurement.purchase_order_line_items
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBPurchaseOrderLineItem(poli)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of purchase order line items from the system.
func (s *Store) Query(ctx context.Context, filter purchaseorderlineitembus.QueryFilter, orderBy order.By, page page.Page) ([]purchaseorderlineitembus.PurchaseOrderLineItem, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
		id, purchase_order_id, supplier_product_id,
		quantity_ordered, quantity_received, quantity_cancelled,
		unit_cost, discount, line_total, line_item_status_id,
		expected_delivery_date, actual_delivery_date, notes,
		created_by, updated_by, created_date, updated_date
	FROM
		procurement.purchase_order_line_items`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbItems []purchaseOrderLineItem
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbItems); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusPurchaseOrderLineItems(dbItems), nil
}

// Count returns the total number of purchase order line items in the DB.
func (s *Store) Count(ctx context.Context, filter purchaseorderlineitembus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
	SELECT
		COUNT(1) AS count
	FROM
		procurement.purchase_order_line_items`

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

// QueryByID retrieves a single purchase order line item from the system by its ID.
func (s *Store) QueryByID(ctx context.Context, poliID uuid.UUID) (purchaseorderlineitembus.PurchaseOrderLineItem, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: poliID.String(),
	}

	const q = `
	SELECT
		id, purchase_order_id, supplier_product_id,
		quantity_ordered, quantity_received, quantity_cancelled,
		unit_cost, discount, line_total, line_item_status_id,
		expected_delivery_date, actual_delivery_date, notes,
		created_by, updated_by, created_date, updated_date
	FROM
		procurement.purchase_order_line_items
	WHERE
		id = :id`

	var dbItem purchaseOrderLineItem
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbItem); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return purchaseorderlineitembus.PurchaseOrderLineItem{}, fmt.Errorf("db: %w", purchaseorderlineitembus.ErrNotFound)
		}
		return purchaseorderlineitembus.PurchaseOrderLineItem{}, fmt.Errorf("db: %w", err)
	}

	return toBusPurchaseOrderLineItem(dbItem), nil
}

// QueryByIDs retrieves a list of purchase order line items from the system by their IDs.
func (s *Store) QueryByIDs(ctx context.Context, poliIDs []uuid.UUID) ([]purchaseorderlineitembus.PurchaseOrderLineItem, error) {
	uuidStrings := make([]string, len(poliIDs))
	for i, id := range poliIDs {
		uuidStrings[i] = id.String()
	}

	data := struct {
		LineItemIDs []string `db:"line_item_ids"`
	}{
		LineItemIDs: uuidStrings,
	}

	const q = `
	SELECT
		id, purchase_order_id, supplier_product_id,
		quantity_ordered, quantity_received, quantity_cancelled,
		unit_cost, discount, line_total, line_item_status_id,
		expected_delivery_date, actual_delivery_date, notes,
		created_by, updated_by, created_date, updated_date
	FROM
		procurement.purchase_order_line_items
	WHERE
		id IN (:line_item_ids)`

	var dbItems []purchaseOrderLineItem
	if err := sqldb.NamedQuerySliceUsingIn(ctx, s.log, s.db, q, data, &dbItems); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusPurchaseOrderLineItems(dbItems), nil
}

// QueryByPurchaseOrderID retrieves all line items for a specific purchase order.
func (s *Store) QueryByPurchaseOrderID(ctx context.Context, poID uuid.UUID) ([]purchaseorderlineitembus.PurchaseOrderLineItem, error) {
	data := struct {
		PurchaseOrderID string `db:"purchase_order_id"`
	}{
		PurchaseOrderID: poID.String(),
	}

	const q = `
	SELECT
		id, purchase_order_id, supplier_product_id,
		quantity_ordered, quantity_received, quantity_cancelled,
		unit_cost, discount, line_total, line_item_status_id,
		expected_delivery_date, actual_delivery_date, notes,
		created_by, updated_by, created_date, updated_date
	FROM
		procurement.purchase_order_line_items
	WHERE
		purchase_order_id = :purchase_order_id
	ORDER BY
		created_date ASC`

	var dbItems []purchaseOrderLineItem
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbItems); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusPurchaseOrderLineItems(dbItems), nil
}