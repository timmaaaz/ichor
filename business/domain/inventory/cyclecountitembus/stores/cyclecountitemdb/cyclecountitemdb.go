package cyclecountitemdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountitembus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for cycle count item database access.
type Store struct {
	log *logger.Logger
	db  sqlx.ExtContext
}

// NewStore constructs the API for data access.
func NewStore(log *logger.Logger, db *sqlx.DB) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

// NewWithTx constructs a new Store value replacing the sqlx.DB
// value with a sqlx.Tx value that is currently inside a transaction.
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (cyclecountitembus.Storer, error) {
	ec, err := sqldb.GetExtContext(tx)
	if err != nil {
		return nil, err
	}

	return &Store{
		log: s.log,
		db:  ec,
	}, nil
}

// Create inserts a new cycle count item into the database.
func (s *Store) Create(ctx context.Context, item cyclecountitembus.CycleCountItem) error {
	const q = `
	INSERT INTO inventory.cycle_count_items
		(id, item_code, session_id, product_id, location_id, system_quantity, counted_quantity, variance, status, counted_by, counted_date, created_date, updated_date)
	VALUES
		(:id, :item_code, :session_id, :product_id, :location_id, :system_quantity, :counted_quantity, :variance, :status, :counted_by, :counted_date, :created_date, :updated_date)
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBCycleCountItem(item)); err != nil {
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", cyclecountitembus.ErrForeignKeyViolation)
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", cyclecountitembus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update modifies an existing cycle count item in the database.
func (s *Store) Update(ctx context.Context, item cyclecountitembus.CycleCountItem) error {
	const q = `
	UPDATE inventory.cycle_count_items
	SET
		item_code        = :item_code,
		counted_quantity = :counted_quantity,
		variance         = :variance,
		status           = :status,
		counted_by       = :counted_by,
		counted_date     = :counted_date,
		updated_date     = :updated_date
	WHERE
		id = :id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBCycleCountItem(item)); err != nil {
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", cyclecountitembus.ErrForeignKeyViolation)
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", cyclecountitembus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes a cycle count item from the database.
func (s *Store) Delete(ctx context.Context, item cyclecountitembus.CycleCountItem) error {
	const q = `
	DELETE FROM inventory.cycle_count_items
	WHERE id = :id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBCycleCountItem(item)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of cycle count items from the database.
func (s *Store) Query(ctx context.Context, filter cyclecountitembus.QueryFilter, orderBy order.By, page page.Page) ([]cyclecountitembus.CycleCountItem, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
		id, item_code, session_id, product_id, location_id, system_quantity, counted_quantity, variance, status, counted_by, counted_date, created_date, updated_date
	FROM
		inventory.cycle_count_items
	`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbItems []cycleCountItem
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbItems); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	items, err := toBusCycleCountItems(dbItems)
	if err != nil {
		return nil, fmt.Errorf("tobuscyclecountitems: %w", err)
	}

	return items, nil
}

// Count returns the total number of cycle count items matching the filter.
func (s *Store) Count(ctx context.Context, filter cyclecountitembus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
	SELECT
		COUNT(1) AS count
	FROM
		inventory.cycle_count_items
	`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("namedqueryrow: %w", err)
	}

	return count.Count, nil
}

// QueryByID retrieves a single cycle count item by its ID.
func (s *Store) QueryByID(ctx context.Context, itemID uuid.UUID) (cyclecountitembus.CycleCountItem, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: itemID.String(),
	}

	const q = `
	SELECT
		id, item_code, session_id, product_id, location_id, system_quantity, counted_quantity, variance, status, counted_by, counted_date, created_date, updated_date
	FROM
		inventory.cycle_count_items
	WHERE
		id = :id
	`

	var dbItem cycleCountItem
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbItem); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return cyclecountitembus.CycleCountItem{}, cyclecountitembus.ErrNotFound
		}
		return cyclecountitembus.CycleCountItem{}, fmt.Errorf("namedqueryrow: %w", err)
	}

	item, err := toBusCycleCountItem(dbItem)
	if err != nil {
		return cyclecountitembus.CycleCountItem{}, fmt.Errorf("tobuscyclecountitem: %w", err)
	}

	return item, nil
}
