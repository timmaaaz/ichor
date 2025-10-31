package rolepagedb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/core/rolepagebus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for role page database access.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (rolepagebus.Storer, error) {
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

// Create adds a new role page mapping to the system
func (s *Store) Create(ctx context.Context, rp rolepagebus.RolePage) error {
	const q = `
	INSERT INTO core.role_pages (
		id, role_id, page_id, can_access, show_in_menu
	) VALUES (
		:id, :role_id, :page_id, :can_access, :show_in_menu
	)
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBRolePage(rp)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", rolepagebus.ErrUnique)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update modifies a role page mapping in the system
func (s *Store) Update(ctx context.Context, rp rolepagebus.RolePage) error {
	const q = `
	UPDATE
		core.role_pages
	SET
		can_access = :can_access,
		show_in_menu = :show_in_menu
	WHERE
		id = :id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBRolePage(rp)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", rolepagebus.ErrUnique)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes a role page mapping from the system
func (s *Store) Delete(ctx context.Context, rp rolepagebus.RolePage) error {
	const q = `
	DELETE FROM
		core.role_pages
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBRolePage(rp)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of role page mappings from the system
func (s *Store) Query(ctx context.Context, filter rolepagebus.QueryFilter, orderBy order.By, page page.Page) ([]rolepagebus.RolePage, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
		id, role_id, page_id, can_access
	FROM
		core.role_pages`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbRolePages []rolePage
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbRolePages); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusRolePages(dbRolePages), nil
}

// Count returns the total number of role page mappings in the DB.
func (s *Store) Count(ctx context.Context, filter rolepagebus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
	SELECT
		COUNT(1) AS count
	FROM
		core.role_pages`

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

// QueryByID retrieves a single role page mapping from the system by its ID.
func (s *Store) QueryByID(ctx context.Context, rolePageID uuid.UUID) (rolepagebus.RolePage, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: rolePageID.String(),
	}

	const q = `
	SELECT
		id, role_id, page_id, can_access
	FROM
		core.role_pages
	WHERE
		id = :id`

	var dbRolePage rolePage
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbRolePage); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return rolepagebus.RolePage{}, fmt.Errorf("db: %w", rolepagebus.ErrNotFound)
		}
		return rolepagebus.RolePage{}, fmt.Errorf("db: %w", err)
	}

	return toBusRolePage(dbRolePage), nil
}
