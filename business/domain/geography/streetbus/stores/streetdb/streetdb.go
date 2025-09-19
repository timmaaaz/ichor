package streetdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/geography/streetbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for streets database access.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (streetbus.Storer, error) {
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

// Create inserts a new street into the database.
func (s *Store) Create(ctx context.Context, str streetbus.Street) error {
	const q = `
    INSERT INTO geography.streets (
        id, city_id, line_1, line_2, postal_code
    ) VALUES (
        :id, :city_id, :line_1, :line_2, :postal_code
    )
    `
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBStreet(str)); err != nil {
		// No duplicate entry check because there is no unique constraint.
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update modifies a street in the database.
func (s *Store) Update(ctx context.Context, str streetbus.Street) error {
	const q = `
    UPDATE geography.streets
    SET
        city_id = :city_id,
        line_1 = :line_1,
        line_2 = :line_2,
        postal_code = :postal_code
    WHERE
        id = :id
    `
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBStreet(str)); err != nil {
		// No duplicate entry check because there is no unique constraint.
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes a street from the database.
func (s *Store) Delete(ctx context.Context, str streetbus.Street) error {
	const q = `
    DELETE FROM 
        geography.streets
    WHERE
        id = :id
    `
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBStreet(str)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of streets from the database.
func (s *Store) Query(ctx context.Context, filter streetbus.QueryFilter, orderBy order.By, page page.Page) ([]streetbus.Street, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
    SELECT
        id, city_id, line_1, line_2, postal_code
    FROM
        geography.streets`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var streets []street
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &streets); err != nil {
		return nil, fmt.Errorf("namedselectcontext: %w", err)
	}

	return toBusStreets(streets), nil
}

// Count returns the number of streets in the database.
func (s *Store) Count(ctx context.Context, filter streetbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
    SELECT
        COUNT(1) AS count
    FROM
        geography.streets`

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

// QueryByID retrieves a single street from the database by its ID.
func (s *Store) QueryByID(ctx context.Context, streetID uuid.UUID) (streetbus.Street, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: streetID.String(),
	}

	const q = `
    SELECT
        id, city_id, line_1, line_2, postal_code
    FROM
        geography.streets
    WHERE
        id = :id
    `
	var str street

	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &str); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return streetbus.Street{}, streetbus.ErrNotFound
		}
		return streetbus.Street{}, fmt.Errorf("querystruct: %w", err)
	}

	return toBusStreet(str), nil
}
