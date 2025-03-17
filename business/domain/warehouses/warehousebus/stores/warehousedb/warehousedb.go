package warehousedb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/domain/warehouses/warehousebus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for warehouse database access.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (warehousebus.Storer, error) {
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

// Create inserts a new warehouse into the database.
func (s *Store) Create(ctx context.Context, bus warehousebus.Warehouse) error {
	const q = `
		INSERT INTO warehouses
			(warehouse_id, street_id, name, is_active, date_created, date_updated, created_by, updated_by)
		VALUES
			(:warehouse_id, :street_id, :name, :is_active, :date_created, :date_updated, :created_by, :updated_by)
		`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBWarehouse(bus)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", userbus.ErrUniqueEmail)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update replaces a warehouse document in the database.
func (s *Store) Update(ctx context.Context, bus warehousebus.Warehouse) error {
	const q = `
		UPDATE 
			warehouses
		SET
			street_id = :street_id,
			name = :name,
			is_active = :is_active,
			date_updated = :date_updated,
			updated_by = :updated_by
		WHERE
			warehouse_id = :warehouse_id
		`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBWarehouse(bus)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return warehousebus.ErrUniqueEntry
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes a warehouse from the database.
func (s *Store) Delete(ctx context.Context, bus warehousebus.Warehouse) error {
	const q = `
		DELETE FROM
			warehouses
		WHERE
			warehouse_id = :warehouse_id
		`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBWarehouse(bus)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// QueryByID gets the specified warehouses from the database.
func (s *Store) Query(ctx context.Context, filter warehousebus.QueryFilter, orderBy order.By, page page.Page) ([]warehousebus.Warehouse, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
		warehouse_id,
		street_id,
		name,
		is_active,
		date_created,
		date_updated,
		created_by,
		updated_by
	FROM
		warehouses`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbW []warehouse
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbW); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusWarehouses(dbW), nil
}

// Count returns the number of warehouses in the database.
func (s *Store) Count(ctx context.Context, filter warehousebus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
	SELECT
		COUNT(*) AS count
	FROM
		warehouses`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("namedquerystruct: %w", err)
	}

	return count.Count, nil
}

// QueryByID gets the specified warehouse from the database.
func (s *Store) QueryByID(ctx context.Context, wID uuid.UUID) (warehousebus.Warehouse, error) {
	data := struct {
		WarehouseID uuid.UUID `db:"warehouse_id"`
	}{
		WarehouseID: wID,
	}

	const q = `
	SELECT
		warehouse_id,
		street_id,
		name,
		is_active,
		date_created,
		date_updated,
		created_by,
		updated_by
	FROM
		warehouses
	WHERE
		warehouse_id = :warehouse_id`

	var dbW warehouse
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbW); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return warehousebus.Warehouse{}, fmt.Errorf("db: %w", warehousebus.ErrNotFound)
		}
		return warehousebus.Warehouse{}, fmt.Errorf("namedquerystruct: %w", err)
	}
	return toBusWarehouse(dbW), nil
}
