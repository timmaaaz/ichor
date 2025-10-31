package pagedb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/core/pagebus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for page database access.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (pagebus.Storer, error) {
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

// Create adds a new page to the system
func (s *Store) Create(ctx context.Context, p pagebus.Page) error {
	const q = `
	INSERT INTO core.pages (
		id, path, name, module, icon, sort_order, is_active, show_in_menu
	) VALUES (
		:id, :path, :name, :module, :icon, :sort_order, :is_active, :show_in_menu
	)
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBPage(p)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", pagebus.ErrUnique)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update modifies a page in the system
func (s *Store) Update(ctx context.Context, p pagebus.Page) error {
	const q = `
	UPDATE
		core.pages
	SET
		path = :path,
		name = :name,
		module = :module,
		icon = :icon,
		sort_order = :sort_order,
		is_active = :is_active,
		show_in_menu = :show_in_menu
	WHERE
		id = :id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBPage(p)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", pagebus.ErrUnique)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes a page from the system
func (s *Store) Delete(ctx context.Context, p pagebus.Page) error {
	const q = `
	DELETE FROM
		core.pages
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBPage(p)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of pages from the system
func (s *Store) Query(ctx context.Context, filter pagebus.QueryFilter, orderBy order.By, page page.Page) ([]pagebus.Page, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
		id, path, name, module, icon, sort_order, is_active, show_in_menu
	FROM
		core.pages`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbPages []dbPage
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbPages); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusPages(dbPages), nil
}

// Count returns the total number of pages in the DB.
func (s *Store) Count(ctx context.Context, filter pagebus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
	SELECT
		COUNT(1) AS count
	FROM
		core.pages`

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

// QueryByID retrieves a single page from the system by its ID.
func (s *Store) QueryByID(ctx context.Context, pageID uuid.UUID) (pagebus.Page, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: pageID.String(),
	}

	const q = `
	SELECT
		id, path, name, module, icon, sort_order, is_active, show_in_menu
	FROM
		core.pages
	WHERE
		id = :id`

	var dbPage dbPage
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbPage); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return pagebus.Page{}, fmt.Errorf("db: %w", pagebus.ErrNotFound)
		}
		return pagebus.Page{}, fmt.Errorf("db: %w", err)
	}

	return toBusPage(dbPage), nil
}

// QueryByUserID retrieves all pages accessible to a user based on their role assignments.
// This performs a join across user_roles and role_pages to find all pages the user can access.
func (s *Store) QueryByUserID(ctx context.Context, userID uuid.UUID) ([]pagebus.Page, error) {
	data := struct {
		UserID string `db:"user_id"`
	}{
		UserID: userID.String(),
	}

	const q = `
	SELECT DISTINCT
		p.id, p.path, p.name, p.module, p.icon, p.sort_order, p.is_active, p.show_in_menu
	FROM
		core.pages p
		INNER JOIN core.role_pages rp ON p.id = rp.page_id
		INNER JOIN core.user_roles ur ON rp.role_id = ur.role_id
	WHERE
		ur.user_id = :user_id
		AND rp.can_access = true
		AND p.is_active = true
	ORDER BY
		p.sort_order ASC, p.name ASC`

	var dbPages []dbPage
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbPages); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusPages(dbPages), nil
}
