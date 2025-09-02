package inventoryitemdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/inventoryitembus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for inventory products database access.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (inventoryitembus.Storer, error) {
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

// Create inserts a new inventory product into the database.
func (s *Store) Create(ctx context.Context, ip inventoryitembus.InventoryItem) error {
	const q = `
	INSERT INTO inventory.inventory_items (
		id, product_id, location_id, quantity, reserved_quantity, allocated_quantity, 
		minimum_stock, maximum_stock, reorder_point, economic_order_quantity, safety_stock, 
		avg_daily_usage, created_date, updated_date
	) VALUES (
        :id, :product_id, :location_id, :quantity, :reserved_quantity, :allocated_quantity, 
        :minimum_stock, :maximum_stock, :reorder_point, :economic_order_quantity, :safety_stock, 
        :avg_daily_usage, :created_date, :updated_date
    )
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBInventoryItem(ip)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", inventoryitembus.ErrUniqueEntry)
		}
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", inventoryitembus.ErrForeignKeyViolation)
		}
		return err
	}

	return nil
}

// Update updates an existing inventory product in the database.
func (s *Store) Update(ctx context.Context, ip inventoryitembus.InventoryItem) error {
	const q = `
    UPDATE
        inventory.inventory_items
    SET
		id = :id,
		product_id = :product_id,
        location_id = :location_id,
        quantity = :quantity,
        reserved_quantity = :reserved_quantity,
        allocated_quantity = :allocated_quantity,
        minimum_stock = :minimum_stock,
        maximum_stock = :maximum_stock,
        reorder_point = :reorder_point,
        economic_order_quantity = :economic_order_quantity,
        safety_stock = :safety_stock,
        avg_daily_usage = :avg_daily_usage,
        updated_date = :updated_date
    WHERE
        id = :id
    `

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBInventoryItem(ip)); err != nil {
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", inventoryitembus.ErrForeignKeyViolation)
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", inventoryitembus.ErrUniqueEntry)
		}
		return err
	}

	return nil
}

// Delete removes an existing inventory product from the database.
func (s *Store) Delete(ctx context.Context, ip inventoryitembus.InventoryItem) error {
	const q = `
	DELETE FROM
	    inventory.inventory_items
	WHERE
		id = :id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBInventoryItem(ip)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves inventory products from the database.
func (s *Store) Query(ctx context.Context, filter inventoryitembus.QueryFilter, orderBy order.By, page page.Page) ([]inventoryitembus.InventoryItem, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
    SELECT
        id, product_id, location_id, quantity, reserved_quantity, allocated_quantity, 
        minimum_stock, maximum_stock, reorder_point, economic_order_quantity, safety_stock, 
        avg_daily_usage, created_date, updated_date
    FROM
        inventory.inventory_items
    `

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var ips []inventoryItem
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &ips); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusInventoryItems(ips), nil
}

// Count retrieves the count of inventory products from the database.
func (s *Store) Count(ctx context.Context, filter inventoryitembus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
    SELECT
        COUNT(1) as count
    FROM
        inventory.inventory_items
    `

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

func (s *Store) QueryByID(ctx context.Context, itemID uuid.UUID) (inventoryitembus.InventoryItem, error) {
	data := struct {
		ItemID string `db:"id"`
	}{
		ItemID: itemID.String(),
	}

	const q = `
    SELECT
        id, product_id, location_id, quantity, reserved_quantity, allocated_quantity, 
        minimum_stock, maximum_stock, reorder_point, economic_order_quantity, safety_stock, 
        avg_daily_usage, created_date, updated_date
    FROM
        inventory.inventory_items
    WHERE
        id = :id

    `

	var ip inventoryItem
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &ip); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return inventoryitembus.InventoryItem{}, inventoryitembus.ErrNotFound
		}
		return inventoryitembus.InventoryItem{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusInventoryItem(ip), nil
}
