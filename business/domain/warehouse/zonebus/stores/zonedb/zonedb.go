package zonedb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/warehouse/zonebus"
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (zonebus.Storer, error) {
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

// Create inserts a new zone into the database.
func (s *Store) Create(ctx context.Context, bus zonebus.Zone) error {
	const q = `
		INSERT INTO zones
			(zone_id, warehouse_id, name, description, is_active, date_created, date_updated, created_by, updated_by)
		VALUES
			(:zone_id, :warehouse_id, :name, :description, :is_active, :date_created, :date_updated, :created_by, :updated_by)
		`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBZone(bus)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", zonebus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

// Update replaces a zone document in the database.
func (s *Store) Update(ctx context.Context, bus zonebus.Zone) error {
	const q = `
		UPDATE 
			zones
		SET
			warehouse_id = :warehouse_id,
			name = :name,
			description = :description,
			is_active = :is_active,
			date_updated = :date_updated,
			updated_by = :updated_by
		WHERE
			zone_id = :zone_id
		`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBZone(bus)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", zonebus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

// Delete removes a zone from the database.
func (s *Store) Delete(ctx context.Context, bus zonebus.Zone) error {
	const q = `
		DELETE FROM 
			zones
		WHERE
			zone_id = :zone_id
		`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBZone(bus)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

// Query retrieves a list of existing zones from the database.
func (s *Store) Query(ctx context.Context, filter zonebus.QueryFilter, orderBy order.By, page page.Page) ([]zonebus.Zone, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
		zone_id, 
		warehouse_id, 
		name, 
		description, 
		is_active, 
		date_created, 
		date_updated, 
		created_by, 
		updated_by
	FROM
		zones
	`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var db []zone
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &db); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusZones(db), nil
}

// Count returns the number of zones in the database.
func (s *Store) Count(ctx context.Context, filter zonebus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
	SELECT
		COUNT(*) AS count
	FROM
		zones`

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

// QueryByID retrieves a single zone from the database by its ID.
func (s *Store) QueryByID(ctx context.Context, id uuid.UUID) (zonebus.Zone, error) {
	data := struct {
		ZoneID uuid.UUID `db:"zone_id"`
	}{
		ZoneID: id,
	}

	const q = `
	SELECT
		zone_id,
		warehouse_id,
		name,
		description,
		is_active,
		date_created,
		date_updated,
		created_by,
		updated_by
	FROM
		zones
	WHERE
		zone_id = :zone_id`

	var db zone
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &db); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return zonebus.Zone{}, fmt.Errorf("namedquerystruct: %w", zonebus.ErrNotFound)
		}
		return zonebus.Zone{}, fmt.Errorf("namedquerystruct: %w", err)
	}
	return toBusZone(db), nil
}
