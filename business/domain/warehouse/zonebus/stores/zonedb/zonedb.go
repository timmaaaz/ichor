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

func (s *Store) Create(ctx context.Context, zone zonebus.Zone) error {
	const q = `
	INSERT INTO zones (
		zone_id, warehouse_id, name, description, created_date, updated_date
	) VALUES (
		:zone_id, :warehouse_id, :name, :description, :created_date, :updated_date
    )
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBZone(zone)); err != nil {
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", zonebus.ErrForeignKeyViolation)
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", zonebus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

func (s *Store) Update(ctx context.Context, zone zonebus.Zone) error {
	const q = `
    UPDATE
        zones
    SET
        zone_id = :zone_id, 
		warehouse_id = :warehouse_id, 
		name = :name, 
		description = :description, 
		updated_date = :updated_date
    WHERE
        zone_id = :zone_id
    `

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBZone(zone)); err != nil {
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", zonebus.ErrForeignKeyViolation)
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", zonebus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

func (s *Store) Delete(ctx context.Context, zone zonebus.Zone) error {
	const q = `
	DELETE FROM
		zones
	WHERE 
		zone_id = :zone_id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBZone(zone)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

func (s *Store) Query(ctx context.Context, filter zonebus.QueryFilter, orderBy order.By, page page.Page) ([]zonebus.Zone, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
	    zone_id, warehouse_id, name, description, created_date, updated_date
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

	var dbZones []zone
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbZones); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusZones(dbZones), nil
}

func (s *Store) Count(ctx context.Context, filter zonebus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
	SELECT 
		COUNT(1) AS count
	FROM
		zones
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

func (s *Store) QueryByID(ctx context.Context, zoneID uuid.UUID) (zonebus.Zone, error) {
	data := struct {
		ID string `db:"zone_id"`
	}{
		ID: zoneID.String(),
	}

	const q = `
	SELECT 
	    zone_id, warehouse_id, name, description, created_date, updated_date
	FROM
		zones
	WHERE
		zone_id = :zone_id
	`

	var dbZone zone

	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbZone); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return zonebus.Zone{}, fmt.Errorf("namedexeccontext: %w", zonebus.ErrNotFound)
		}
		return zonebus.Zone{}, fmt.Errorf("querystruct: %w", err)
	}

	return toBusZone(dbZone), nil
}
