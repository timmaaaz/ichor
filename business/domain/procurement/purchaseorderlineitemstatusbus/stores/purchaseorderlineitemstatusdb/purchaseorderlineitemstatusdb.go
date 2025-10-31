package purchaseorderlineitemstatusdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitemstatusbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for purchase order line item status database access.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (purchaseorderlineitemstatusbus.Storer, error) {
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

// Create adds a new purchase order line item status to the system.
func (s *Store) Create(ctx context.Context, polis purchaseorderlineitemstatusbus.PurchaseOrderLineItemStatus) error {
	const q = `
	INSERT INTO procurement.purchase_order_line_item_statuses (
		id, name, description, sort_order
	) VALUES (
		:id, :name, :description, :sort_order
	)
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBPurchaseOrderLineItemStatus(polis)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", purchaseorderlineitemstatusbus.ErrUnique)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update modifies a purchase order line item status in the system.
func (s *Store) Update(ctx context.Context, polis purchaseorderlineitemstatusbus.PurchaseOrderLineItemStatus) error {
	const q = `
	UPDATE
		procurement.purchase_order_line_item_statuses
	SET
		name = :name,
		description = :description,
		sort_order = :sort_order
	WHERE
		id = :id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBPurchaseOrderLineItemStatus(polis)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", purchaseorderlineitemstatusbus.ErrUnique)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes a purchase order line item status from the system.
func (s *Store) Delete(ctx context.Context, polis purchaseorderlineitemstatusbus.PurchaseOrderLineItemStatus) error {
	const q = `
	DELETE FROM
		procurement.purchase_order_line_item_statuses
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBPurchaseOrderLineItemStatus(polis)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of purchase order line item statuses from the system.
func (s *Store) Query(ctx context.Context, filter purchaseorderlineitemstatusbus.QueryFilter, orderBy order.By, page page.Page) ([]purchaseorderlineitemstatusbus.PurchaseOrderLineItemStatus, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
		id, name, description, sort_order
	FROM
		procurement.purchase_order_line_item_statuses`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbStatuses []purchaseOrderLineItemStatus
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbStatuses); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusPurchaseOrderLineItemStatuses(dbStatuses), nil
}

// Count returns the total number of purchase order line item statuses in the DB.
func (s *Store) Count(ctx context.Context, filter purchaseorderlineitemstatusbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
	SELECT
		COUNT(1) AS count
	FROM
		procurement.purchase_order_line_item_statuses`

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

// QueryByID retrieves a single purchase order line item status from the system by its ID.
func (s *Store) QueryByID(ctx context.Context, polisID uuid.UUID) (purchaseorderlineitemstatusbus.PurchaseOrderLineItemStatus, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: polisID.String(),
	}

	const q = `
	SELECT
		id, name, description, sort_order
	FROM
		procurement.purchase_order_line_item_statuses
	WHERE
		id = :id`

	var dbStatus purchaseOrderLineItemStatus
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbStatus); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return purchaseorderlineitemstatusbus.PurchaseOrderLineItemStatus{}, fmt.Errorf("db: %w", purchaseorderlineitemstatusbus.ErrNotFound)
		}
		return purchaseorderlineitemstatusbus.PurchaseOrderLineItemStatus{}, fmt.Errorf("db: %w", err)
	}

	return toBusPurchaseOrderLineItemStatus(dbStatus), nil
}

// QueryByIDs retrieves a list of purchase order line item statuses from the system by their IDs.
func (s *Store) QueryByIDs(ctx context.Context, polisIDs []uuid.UUID) ([]purchaseorderlineitemstatusbus.PurchaseOrderLineItemStatus, error) {
	uuidStrings := make([]string, len(polisIDs))
	for i, id := range polisIDs {
		uuidStrings[i] = id.String()
	}

	data := struct {
		StatusIDs []string `db:"status_ids"`
	}{
		StatusIDs: uuidStrings,
	}

	const q = `
	SELECT
		id, name, description, sort_order
	FROM
		procurement.purchase_order_line_item_statuses
	WHERE
		id IN (:status_ids)`

	var dbStatuses []purchaseOrderLineItemStatus
	if err := sqldb.NamedQuerySliceUsingIn(ctx, s.log, s.db, q, data, &dbStatuses); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusPurchaseOrderLineItemStatuses(dbStatuses), nil
}

// QueryAll retrieves all purchase order line item statuses from the system.
func (s *Store) QueryAll(ctx context.Context) ([]purchaseorderlineitemstatusbus.PurchaseOrderLineItemStatus, error) {
	const q = `
	SELECT
		id, name, description, sort_order
	FROM
		procurement.purchase_order_line_item_statuses
	ORDER BY
		sort_order ASC, name ASC`

	var dbStatuses []purchaseOrderLineItemStatus
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, struct{}{}, &dbStatuses); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusPurchaseOrderLineItemStatuses(dbStatuses), nil
}