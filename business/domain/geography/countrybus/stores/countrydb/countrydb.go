package countrydb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/geography/countrybus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (countrybus.Storer, error) {
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

// Query retrieves a list of existing countries from the database.
func (s *Store) Query(ctx context.Context, filter countrybus.QueryFilter, orderBy order.By, page page.Page) ([]countrybus.Country, error) {
	data := map[string]interface{}{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
		id, number, name, alpha_2, alpha_3
	FROM
		geography.countries`
	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbCtrys []country
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbCtrys); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusCountries(dbCtrys), nil
}

// Count returns the total number of countries in the DB.
func (s *Store) Count(ctx context.Context, filter countrybus.QueryFilter) (int, error) {
	data := map[string]interface{}{}

	const q = `
	SELECT
		count(1)
	FROM
		geography.countries`

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

// QueryByID gets the specified country from the database.
func (s *Store) QueryByID(ctx context.Context, countryID uuid.UUID) (countrybus.Country, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: countryID.String(),
	}

	const q = `
	SELECT
		id, number, name, alpha_2, alpha_3
	FROM
		geography.countries
	WHERE 
		id = :id`

	var dbCtry country
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbCtry); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return countrybus.Country{}, fmt.Errorf("db: %w", countrybus.ErrNotFound)
		}
		return countrybus.Country{}, fmt.Errorf("db: %w", err)
	}

	return toBusCountry(dbCtry), nil
}
