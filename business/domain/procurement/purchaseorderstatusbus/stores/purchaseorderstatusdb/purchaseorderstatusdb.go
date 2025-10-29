package purchaseorderstatusdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderstatusbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for purchase order status database access.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (purchaseorderstatusbus.Storer, error) {
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

// Create adds a new purchase order status to the system.
func (s *Store) Create(ctx context.Context, pos purchaseorderstatusbus.PurchaseOrderStatus) error {
	const q = `
	INSERT INTO procurement.purchase_order_statuses (
		id, name, description, sort_order
	) VALUES (
		:id, :name, :description, :sort_order
	)
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBPurchaseOrderStatus(pos)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", purchaseorderstatusbus.ErrUnique)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update modifies a purchase order status in the system.
func (s *Store) Update(ctx context.Context, pos purchaseorderstatusbus.PurchaseOrderStatus) error {
	const q = `
	UPDATE
		procurement.purchase_order_statuses
	SET
		name = :name,
		description = :description,
		sort_order = :sort_order
	WHERE
		id = :id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBPurchaseOrderStatus(pos)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", purchaseorderstatusbus.ErrUnique)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes a purchase order status from the system.
func (s *Store) Delete(ctx context.Context, pos purchaseorderstatusbus.PurchaseOrderStatus) error {
	const q = `
	DELETE FROM
		procurement.purchase_order_statuses
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBPurchaseOrderStatus(pos)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of purchase order statuses from the system.
func (s *Store) Query(ctx context.Context, filter purchaseorderstatusbus.QueryFilter, orderBy order.By, page page.Page) ([]purchaseorderstatusbus.PurchaseOrderStatus, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
		id, name, description, sort_order
	FROM
		procurement.purchase_order_statuses`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbStatuses []purchaseOrderStatus
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbStatuses); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusPurchaseOrderStatuses(dbStatuses), nil
}

// Count returns the total number of purchase order statuses in the DB.
func (s *Store) Count(ctx context.Context, filter purchaseorderstatusbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
	SELECT
		COUNT(1) AS count
	FROM
		procurement.purchase_order_statuses`

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

// QueryByID retrieves a single purchase order status from the system by its ID.
func (s *Store) QueryByID(ctx context.Context, posID uuid.UUID) (purchaseorderstatusbus.PurchaseOrderStatus, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: posID.String(),
	}

	const q = `
	SELECT
		id, name, description, sort_order
	FROM
		procurement.purchase_order_statuses
	WHERE
		id = :id`

	var dbStatus purchaseOrderStatus
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbStatus); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return purchaseorderstatusbus.PurchaseOrderStatus{}, fmt.Errorf("db: %w", purchaseorderstatusbus.ErrNotFound)
		}
		return purchaseorderstatusbus.PurchaseOrderStatus{}, fmt.Errorf("db: %w", err)
	}

	return toBusPurchaseOrderStatus(dbStatus), nil
}

// QueryByIDs retrieves a list of purchase order statuses from the system by their IDs.
func (s *Store) QueryByIDs(ctx context.Context, posIDs []uuid.UUID) ([]purchaseorderstatusbus.PurchaseOrderStatus, error) {
	uuidStrings := make([]string, len(posIDs))
	for i, id := range posIDs {
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
		procurement.purchase_order_statuses
	WHERE
		id IN (:status_ids)`

	var dbStatuses []purchaseOrderStatus
	if err := sqldb.NamedQuerySliceUsingIn(ctx, s.log, s.db, q, data, &dbStatuses); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusPurchaseOrderStatuses(dbStatuses), nil
}

// QueryAll retrieves all purchase order statuses from the system.
func (s *Store) QueryAll(ctx context.Context) ([]purchaseorderstatusbus.PurchaseOrderStatus, error) {
	const q = `
	SELECT
		id, name, description, sort_order
	FROM
		procurement.purchase_order_statuses
	ORDER BY
		sort_order ASC, name ASC`

	var dbStatuses []purchaseOrderStatus
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, struct{}{}, &dbStatuses); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusPurchaseOrderStatuses(dbStatuses), nil
}