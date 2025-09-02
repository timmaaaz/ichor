package inventorylocationdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/warehouse/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for zone database access.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (inventorylocationbus.Storer, error) {
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

func (s *Store) Create(ctx context.Context, il inventorylocationbus.InventoryLocation) error {
	const q = `
	INSERT INTO inventory.inventory_locations (
		id, zone_id, warehouse_id, aisle, rack, shelf, bin, is_pick_location, 
		is_reserve_location, max_capacity, current_utilization, created_date, updated_date
	) VALUES (
		:id, :zone_id, :warehouse_id, :aisle, :rack, :shelf, :bin, :is_pick_location, 
        :is_reserve_location, :max_capacity, :current_utilization, :created_date, :updated_date
    )
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBInvLocation(il)); err != nil {
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", inventorylocationbus.ErrForeignKeyViolation)
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", inventorylocationbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil

}

func (s *Store) Update(ctx context.Context, il inventorylocationbus.InventoryLocation) error {
	const q = `
    UPDATE
        inventory.inventory_locations
    SET
		id = :id, 
		zone_id = :zone_id, 
		warehouse_id = :warehouse_id, 
		aisle = :aisle, 
		rack = :rack, 
		shelf = :shelf, 
		bin = :bin, 
		is_pick_location = :is_pick_location, 
		is_reserve_location = :is_reserve_location, 
		max_capacity = :max_capacity, 
		current_utilization = :current_utilization, 
		created_date = :created_date, 
		updated_date = :updated_date
	WHERE 
		id = :id
    `

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBInvLocation(il)); err != nil {
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", inventorylocationbus.ErrForeignKeyViolation)
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", inventorylocationbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

func (s *Store) Delete(ctx context.Context, il inventorylocationbus.InventoryLocation) error {
	const q = `
    DELETE FROM
        inventory.inventory_locations
    WHERE
        id = :id
    `

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBInvLocation(il)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

func (s *Store) Query(ctx context.Context, filter inventorylocationbus.QueryFilter, orderBy order.By, page page.Page) ([]inventorylocationbus.InventoryLocation, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
		id, zone_id, warehouse_id, aisle, rack, shelf, bin, is_pick_location, 
		is_reserve_location, max_capacity, current_utilization, created_date, updated_date
	FROM 
		inventory.inventory_locations
	`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbILS []inventoryLocation

	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbILS); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusInvLocations(dbILS)
}

func (s *Store) Count(ctx context.Context, filter inventorylocationbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
    SELECT
        COUNT(1) AS count
    FROM 
        inventory.inventory_locations
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

func (s *Store) QueryByID(ctx context.Context, id uuid.UUID) (inventorylocationbus.InventoryLocation, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: id.String(),
	}

	const q = `
    SELECT
        id, zone_id, warehouse_id, aisle, rack, shelf, bin, is_pick_location, 
        is_reserve_location, max_capacity, current_utilization, created_date, updated_date
    FROM 
        inventory.inventory_locations
    WHERE
        id = :id
    `

	var dbIL inventoryLocation

	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbIL); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return inventorylocationbus.InventoryLocation{}, fmt.Errorf("namedexeccontext: %w", inventorylocationbus.ErrNotFound)
		}
		return inventorylocationbus.InventoryLocation{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusInvLocation(dbIL)
}
