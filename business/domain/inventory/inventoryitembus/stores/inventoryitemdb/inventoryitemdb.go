package inventoryitemdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
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

// QueryAvailableForAllocation queries inventory items that have available quantity for allocation
// with row-level locking for transaction safety
func (s *Store) QueryAvailableForAllocation(ctx context.Context, productID uuid.UUID, locationID *uuid.UUID, warehouseID *uuid.UUID, strategy string, limit int) ([]inventoryitembus.InventoryItem, error) {
	args := map[string]interface{}{
		"product_id": productID.String(),
		"limit":      limit,
	}

	q := `
    SELECT
        ii.id, ii.product_id, ii.location_id, ii.quantity, ii.reserved_quantity, 
        ii.allocated_quantity, ii.minimum_stock, ii.maximum_stock, ii.reorder_point, 
        ii.economic_order_quantity, ii.safety_stock, ii.avg_daily_usage, 
        ii.created_date, ii.updated_date
    FROM
        inventory.inventory_items ii
    WHERE
        ii.product_id = :product_id
        AND (ii.quantity - ii.reserved_quantity - ii.allocated_quantity) > 0`

	// Add location filtering if specified
	if locationID != nil {
		q += ` AND ii.location_id = :location_id`
		args["location_id"] = locationID.String()
	}

	// Add warehouse filtering if specified
	if warehouseID != nil {
		q = `
        SELECT
            ii.id, ii.product_id, ii.location_id, ii.quantity, ii.reserved_quantity, 
            ii.allocated_quantity, ii.minimum_stock, ii.maximum_stock, ii.reorder_point, 
            ii.economic_order_quantity, ii.safety_stock, ii.avg_daily_usage, 
            ii.created_date, ii.updated_date
        FROM
            inventory.inventory_items ii
            INNER JOIN inventory.inventory_locations il ON ii.location_id = il.id
        WHERE
            ii.product_id = :product_id
            AND il.warehouse_id = :warehouse_id
            AND (ii.quantity - ii.reserved_quantity - ii.allocated_quantity) > 0`
		args["warehouse_id"] = warehouseID.String()
	}

	// Apply ordering based on strategy
	switch strategy {
	case "fifo":
		q += " ORDER BY ii.created_date ASC"
	case "lifo":
		q += " ORDER BY ii.created_date DESC"
	default:
		q += " ORDER BY ii.created_date ASC"
	}

	/* TODO: Advanced Allocation Strategies
	 *
	 * nearest_expiry: Requires joining with lot_trackings table
	 *   - Need to link inventory_items to lot_trackings (currently no FK relationship)
	 *   - ORDER BY lt.expiration_date ASC
	 *   - Consider adding lot_id to inventory_items table
	 *
	 * lowest_cost: Requires location-specific cost data
	 *   - product_costs table exists but is product-level, not location-specific
	 *   - Consider: carrying costs vary by warehouse, landed costs differ by location
	 *   - May need inventory_location_costs table with warehouse-specific costs
	 *
	 * nearest_location: Requires customer shipping address
	 *   - Need to calculate distance from warehouse to customer
	 *   - Could use geography tables (countries, regions, cities, streets)
	 *   - Consider caching distance calculations or using geospatial queries
	 *
	 * load_balancing: Requires warehouse utilization metrics
	 *   - inventory_locations has current_utilization field
	 *   - Need to aggregate by warehouse and factor into allocation decision
	 *   - Consider warehouse capacity and current workload
	 *
	 * priority_zone: Requires zone prioritization logic
	 *   - Use is_pick_location vs is_reserve_location flags
	 *   - May need zone_priorities table for customer-specific rules
	 *   - VIP customers get allocation from premium/faster zones
	 */

	q += " LIMIT :limit FOR UPDATE" // Row-level locking for transaction safety

	var items []inventoryItem
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, args, &items); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusInventoryItems(items), nil
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
		ID string `db:"id"`
	}{
		ID: itemID.String(),
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

// UpsertQuantity atomically creates or updates the inventory item for the given
// (product_id, location_id) pair, adding quantityDelta to the existing quantity.
// On insert (no existing record), all stock threshold fields default to 0.
// Relies on the unique_product_location constraint added in migration v1.998.
func (s *Store) UpsertQuantity(ctx context.Context, productID, locationID uuid.UUID, quantityDelta int) error {
	data := struct {
		ID          uuid.UUID `db:"id"`
		ProductID   uuid.UUID `db:"product_id"`
		LocationID  uuid.UUID `db:"location_id"`
		Quantity    int       `db:"quantity"`
	}{
		ID:         uuid.New(),
		ProductID:  productID,
		LocationID: locationID,
		Quantity:   quantityDelta,
	}

	const q = `
	INSERT INTO inventory.inventory_items
		(id, product_id, location_id, quantity,
		 reserved_quantity, allocated_quantity,
		 minimum_stock, maximum_stock, reorder_point,
		 economic_order_quantity, safety_stock, avg_daily_usage,
		 created_date, updated_date)
	VALUES
		(:id, :product_id, :location_id, :quantity,
		 0, 0, 0, 0, 0, 0, 0, 0,
		 NOW(), NOW())
	ON CONFLICT (product_id, location_id)
	DO UPDATE SET
		quantity     = inventory.inventory_items.quantity + EXCLUDED.quantity,
		updated_date = NOW()
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}
