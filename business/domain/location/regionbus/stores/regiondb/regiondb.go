package regiondb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"bitbucket.org/superiortechnologies/ichor/business/domain/location/regionbus"
	"bitbucket.org/superiortechnologies/ichor/business/sdk/order"
	"bitbucket.org/superiortechnologies/ichor/business/sdk/page"
	"bitbucket.org/superiortechnologies/ichor/business/sdk/sqldb"
	"bitbucket.org/superiortechnologies/ichor/foundation/logger"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Store manages the set of APIs for country database access.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (regionbus.Storer, error) {
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

func (s *Store) Query(ctx context.Context, filter regionbus.QueryFilter, orderBy order.By, page page.Page) ([]regionbus.Region, error) {
	data := map[string]interface{}{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
    SELECT
        region_id, country_id, name, code
    FROM
        regions`
	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbRgns []region
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbRgns); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusRegions(dbRgns), nil
}

// Count returns the total number of regions in the DB.
func (s *Store) Count(ctx context.Context, filter regionbus.QueryFilter) (int, error) {
	data := map[string]interface{}{}

	const q = `
	SELECT
		count(1)
	FROM
		regions`

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

// QueryByID retrieves a single region by its ID.
func (s *Store) QueryByID(ctx context.Context, regionID uuid.UUID) (regionbus.Region, error) {
	data := map[string]interface{}{
		"region_id": regionID,
	}

	const q = `
	SELECT
		region_id, country_id, name, code
	FROM
		regions
	WHERE
		region_id = :region_id`

	var dbRgn region
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbRgn); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return regionbus.Region{}, regionbus.ErrNotFound
		}
		return regionbus.Region{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusRegion(dbRgn), nil
}
