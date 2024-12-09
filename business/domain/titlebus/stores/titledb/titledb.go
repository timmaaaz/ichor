package titledb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/titlebus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for title database access.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (titlebus.Storer, error) {
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

// Create inserts a new title into the database.
func (s *Store) Create(ctx context.Context, fs titlebus.Title) error {
	const q = `
    INSERT INTO titles (
        title_id, description, name
    ) VALUES (
        :title_id, :description, :name
    )
    `
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBTitle(fs)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", titlebus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update replaces an title document in the database.
func (s *Store) Update(ctx context.Context, fs titlebus.Title) error {
	const q = `
	UPDATE titles
	SET 
	    description = :description,
        name = :name
	WHERE 
		title_id = :title_id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBTitle(fs)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", titlebus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes an title from the database.
func (s *Store) Delete(ctx context.Context, as titlebus.Title) error {
	const q = `
	DELETE FROM
		titles
	WHERE
		title_id = :title_id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBTitle(as)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of titles from the database.
func (s *Store) Query(ctx context.Context, filter titlebus.QueryFilter, orderBy order.By, page page.Page) ([]titlebus.Title, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT 
		title_id, description, name
	FROM
		titles
	`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var titles []title

	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &titles); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusTitles(titles), nil
}

// Count returns the total number of titlees
func (s *Store) Count(ctx context.Context, filter titlebus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
    SELECT
        COUNT(1) AS count
    FROM
        titles`

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

// QueryByID finds the title by the specified ID.
func (s *Store) QueryByID(ctx context.Context, aprvlStatusID uuid.UUID) (titlebus.Title, error) {
	data := struct {
		ID string `db:"title_id"`
	}{
		ID: aprvlStatusID.String(),
	}

	const q = `
    SELECT
        title_id, description, name
    FROM
        titles
    WHERE
        title_id = :title_id
    `

	var titles title
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &titles); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return titlebus.Title{}, fmt.Errorf("db: %w", titlebus.ErrNotFound)
		}
		return titlebus.Title{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusTitle(titles), nil
}
