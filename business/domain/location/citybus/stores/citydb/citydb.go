package citydb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/location/citybus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for cities database access.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (citybus.Storer, error) {
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

// Create inserts a new city into the database.
func (s *Store) Create(ctx context.Context, cty citybus.City) error {
	const q = `
    INSERT INTO cities (
        id, region_id, name
    ) VALUES (
        :id, :region_id, :name
    )
    `
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBCity(cty)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", citybus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update replaces a city document in the database.
func (s *Store) Update(ctx context.Context, cty citybus.City) error {
	const q = `
    UPDATE cities
    SET
        region_id = :region_id,
        name = :name
    WHERE
        id = :id
    `
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBCity(cty)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", citybus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes a city from the database.
func (s *Store) Delete(ctx context.Context, cty citybus.City) error {
	const q = `
    DELETE FROM 
        cities
    WHERE
        id = :id
    `
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBCity(cty)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of cities from the database.
func (s *Store) Query(ctx context.Context, filter citybus.QueryFilter, orderBy order.By, page page.Page) ([]citybus.City, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
    SELECT
        id, region_id, name
    FROM
        cities`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var cities []city

	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &cities); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusCities(cities), nil
}

// Count returns the total number of cities.
func (s *Store) Count(ctx context.Context, filter citybus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
    SELECT
        COUNT(1) AS count
    FROM
        cities`

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

// QueryByID finds the city by the specified ID.
func (s *Store) QueryByID(ctx context.Context, cityID uuid.UUID) (citybus.City, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: cityID.String(),
	}

	const q = `
    SELECT
        id, region_id, name
    FROM
        cities
    WHERE
        id = :id
    `

	var city city
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &city); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return citybus.City{}, fmt.Errorf("db: %w", citybus.ErrNotFound)
		}
		return citybus.City{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusCity(city), nil
}
