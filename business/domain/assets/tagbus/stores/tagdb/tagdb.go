package tagdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/assets/tagbus"
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (tagbus.Storer, error) {
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

// Create inserts a new tag into the database.
func (s *Store) Create(ctx context.Context, t tagbus.Tag) error {
	const q = `
    INSERT INTO tags (
        id, name, description
    ) VALUES (
        :id, :name, :description
    )
    `
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBTag(t)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", tagbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update modifies data about an tag in the database.
func (s *Store) Update(ctx context.Context, t tagbus.Tag) error {
	const q = `
    UPDATE 
        tags
    SET
        name = :name,
        description = :description
    WHERE
        id = :id
    `
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBTag(t)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", tagbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes an tag from the database.
func (s *Store) Delete(ctx context.Context, at tagbus.Tag) error {
	const q = `
    DELETE FROM
        tags
    WHERE
        id = :id
    `
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBTag(at)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of existing tags from the database.
func (s *Store) Query(ctx context.Context, filter tagbus.QueryFilter, orderBy order.By, page page.Page) ([]tagbus.Tag, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
    SELECT
        id, name, description
    FROM
        tags`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbTs []tag
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbTs); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusTags(dbTs), nil
}

// Count returns the total number of tags in the DB.
func (s *Store) Count(ctx context.Context, filter tagbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
    SELECT
        COUNT(1) AS count
    FROM
        tags`

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

// QueryByID retrieves a single tag by its id.
func (s *Store) QueryByID(ctx context.Context, id uuid.UUID) (tagbus.Tag, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: id.String(),
	}

	const q = `
    SELECT
        id, name, description
    FROM
        tags
    WHERE
        id = :id
    `

	var dbT tag
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbT); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return tagbus.Tag{}, fmt.Errorf("db: %w", tagbus.ErrNotFound)
		}
		return tagbus.Tag{}, fmt.Errorf("querystruct: %w", err)
	}

	return toBusTag(dbT), nil
}
