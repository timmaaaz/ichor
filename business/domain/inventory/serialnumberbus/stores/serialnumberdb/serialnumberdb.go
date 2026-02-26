package serialnumberdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/inventory/serialnumberbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for metrics database access.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (serialnumberbus.Storer, error) {
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

func (s *Store) Create(ctx context.Context, sn serialnumberbus.SerialNumber) error {
	const q = `
    INSERT INTO inventory.serial_numbers (
        id, lot_id, product_id, location_id, serial_number, status, created_date, updated_date
    ) VALUES (
        :id, :lot_id, :product_id, :location_id, :serial_number, :status, :created_date, :updated_date
    )
    `

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBSerialNumber(sn)); err != nil {
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", serialnumberbus.ErrForeignKeyViolation)
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", serialnumberbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil

}

func (s *Store) Update(ctx context.Context, sn serialnumberbus.SerialNumber) error {
	const q = `
    UPDATE
        inventory.serial_numbers
    SET
        id = :id,
        lot_id = :lot_id,
        product_id = :product_id,
        location_id = :location_id,
        serial_number = :serial_number,
        status = :status,
        updated_date = :updated_date
    WHERE
        id = :id
    `

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBSerialNumber(sn)); err != nil {
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", serialnumberbus.ErrForeignKeyViolation)
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", serialnumberbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

func (s *Store) Delete(ctx context.Context, sn serialnumberbus.SerialNumber) error {
	const q = `
    DELETE FROM
        inventory.serial_numbers
    WHERE
        id = :id
    `

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBSerialNumber(sn)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

func (s *Store) Query(ctx context.Context, filter serialnumberbus.QueryFilter, orderBy order.By, page page.Page) ([]serialnumberbus.SerialNumber, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
		id, lot_id, product_id, location_id, serial_number, status, created_date, updated_date
	FROM 
		inventory.serial_numbers
	`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var sns []serialNumber
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &sns); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusSerialNumbers(sns), nil

}

func (s *Store) Count(ctx context.Context, filter serialnumberbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
	SELECT
		COUNT(1) AS count
	FROM 
		inventory.serial_numbers
		`

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

// QueryLocationBySerialID retrieves the location of the specified serial number.
func (s *Store) QueryLocationBySerialID(ctx context.Context, serialID uuid.UUID) (serialnumberbus.SerialLocation, error) {
	data := struct {
		SerialID string `db:"serial_id"`
	}{
		SerialID: serialID.String(),
	}

	const q = `
	SELECT
	    il.id AS location_id,
	    COALESCE(il.location_code, CONCAT(il.aisle, '-', il.rack, '-', il.shelf, '-', il.bin)) AS location_code,
	    il.aisle,
	    il.rack,
	    il.shelf,
	    il.bin,
	    w.name AS warehouse_name,
	    z.name AS zone_name
	FROM inventory.serial_numbers sn
	JOIN inventory.inventory_locations il ON il.id = sn.location_id
	JOIN inventory.zones z ON z.id = il.zone_id
	JOIN inventory.warehouses w ON w.id = z.warehouse_id
	WHERE sn.id = :serial_id
	`

	var loc serialLocation
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &loc); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return serialnumberbus.SerialLocation{}, fmt.Errorf("namedquerystruct: %w", serialnumberbus.ErrNotFound)
		}
		return serialnumberbus.SerialLocation{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusSerialLocation(loc), nil
}

// QueryByID gets the specified serial number from the database.
func (s *Store) QueryByID(ctx context.Context, serialID uuid.UUID) (serialnumberbus.SerialNumber, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: serialID.String(),
	}

	const q = `
    SELECT
        id, lot_id, product_id, location_id, serial_number, status, created_date, updated_date
    FROM 
        inventory.serial_numbers
    WHERE
        id = :id
    `

	var sn serialNumber
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &sn); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return serialnumberbus.SerialNumber{}, fmt.Errorf("namedquerystruct: %w", serialnumberbus.ErrNotFound)
		}
		return serialnumberbus.SerialNumber{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusSerialNumber(sn), nil
}
