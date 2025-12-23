package timezonedb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/geography/timezonebus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for timezones database access.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (timezonebus.Storer, error) {
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

// Create inserts a new timezone into the database.
func (s *Store) Create(ctx context.Context, tz timezonebus.Timezone) error {
	const q = `
    INSERT INTO geography.timezones (
        id, name, display_name, utc_offset, is_active
    ) VALUES (
        :id, :name, :display_name, :utc_offset, :is_active
    )
    `
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBTimezone(tz)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", timezonebus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update replaces a timezone document in the database.
func (s *Store) Update(ctx context.Context, tz timezonebus.Timezone) error {
	const q = `
    UPDATE geography.timezones
    SET
        name = :name,
        display_name = :display_name,
        utc_offset = :utc_offset,
        is_active = :is_active
    WHERE
        id = :id
    `
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBTimezone(tz)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", timezonebus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes a timezone from the database.
func (s *Store) Delete(ctx context.Context, tz timezonebus.Timezone) error {
	const q = `
    DELETE FROM
        geography.timezones
    WHERE
        id = :id
    `
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBTimezone(tz)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of timezones from the database.
func (s *Store) Query(ctx context.Context, filter timezonebus.QueryFilter, orderBy order.By, pg page.Page) ([]timezonebus.Timezone, error) {
	data := map[string]any{
		"offset":        (pg.Number() - 1) * pg.RowsPerPage(),
		"rows_per_page": pg.RowsPerPage(),
	}

	const q = `
    SELECT
        id, name, display_name, utc_offset, is_active
    FROM
        geography.timezones`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var tzs []timezone

	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &tzs); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusTimezones(tzs), nil
}

// Count returns the total number of timezones.
func (s *Store) Count(ctx context.Context, filter timezonebus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
    SELECT
        COUNT(1) AS count
    FROM
        geography.timezones`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("namedquerysingle: %w", err)
	}

	return count.Count, nil
}

// QueryByID finds the timezone by the specified ID.
func (s *Store) QueryByID(ctx context.Context, timezoneID uuid.UUID) (timezonebus.Timezone, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: timezoneID.String(),
	}

	const q = `
    SELECT
        id, name, display_name, utc_offset, is_active
    FROM
        geography.timezones
    WHERE
        id = :id
    `

	var tz timezone
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &tz); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return timezonebus.Timezone{}, fmt.Errorf("db: %w", timezonebus.ErrNotFound)
		}
		return timezonebus.Timezone{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusTimezone(tz), nil
}

// QueryAll retrieves all timezones from the database without pagination.
func (s *Store) QueryAll(ctx context.Context) ([]timezonebus.Timezone, error) {
	const q = `
    SELECT
        id, name, display_name, utc_offset, is_active
    FROM
        geography.timezones
    ORDER BY
        name
    `

	var tzs []timezone
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, map[string]any{}, &tzs); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusTimezones(tzs), nil
}
