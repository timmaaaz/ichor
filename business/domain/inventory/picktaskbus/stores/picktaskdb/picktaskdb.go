package picktaskdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/inventory/picktaskbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for pick task database access.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (picktaskbus.Storer, error) {
	ec, err := sqldb.GetExtContext(tx)
	if err != nil {
		return nil, err
	}

	return &Store{
		log: s.log,
		db:  ec,
	}, nil
}

// Create inserts a new pick task into the database.
func (s *Store) Create(ctx context.Context, task picktaskbus.PickTask) error {
	const q = `
	INSERT INTO inventory.pick_tasks
		(id, task_number, sales_order_id, sales_order_line_item_id, product_id, lot_id, serial_id,
		 location_id, quantity_to_pick, quantity_picked, status,
		 assigned_to, assigned_at, completed_by, completed_at,
		 short_pick_reason, created_by, created_date, updated_date, scenario_id)
	VALUES
		(:id, :task_number, :sales_order_id, :sales_order_line_item_id, :product_id, :lot_id, :serial_id,
		 :location_id, :quantity_to_pick, :quantity_picked, :status,
		 :assigned_to, :assigned_at, :completed_by, :completed_at,
		 :short_pick_reason, :created_by, :created_date, :updated_date, :scenario_id)
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBPickTask(task)); err != nil {
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", picktaskbus.ErrForeignKeyViolation)
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", picktaskbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update modifies an existing pick task in the database.
func (s *Store) Update(ctx context.Context, task picktaskbus.PickTask) error {
	const q = `
	UPDATE inventory.pick_tasks
	SET
		task_number               = :task_number,
		sales_order_id            = :sales_order_id,
		sales_order_line_item_id  = :sales_order_line_item_id,
		product_id                = :product_id,
		lot_id                    = :lot_id,
		serial_id                 = :serial_id,
		location_id               = :location_id,
		quantity_to_pick          = :quantity_to_pick,
		quantity_picked           = :quantity_picked,
		status                    = :status,
		assigned_to               = :assigned_to,
		assigned_at               = :assigned_at,
		completed_by              = :completed_by,
		completed_at              = :completed_at,
		short_pick_reason         = :short_pick_reason,
		updated_date              = :updated_date
	WHERE
		id = :id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBPickTask(task)); err != nil {
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", picktaskbus.ErrForeignKeyViolation)
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", picktaskbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes a pick task from the database.
func (s *Store) Delete(ctx context.Context, task picktaskbus.PickTask) error {
	const q = `
	DELETE FROM inventory.pick_tasks
	WHERE id = :id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBPickTask(task)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of pick tasks from the database.
func (s *Store) Query(ctx context.Context, filter picktaskbus.QueryFilter, orderBy order.By, page page.Page) ([]picktaskbus.PickTask, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
		id, task_number, sales_order_id, sales_order_line_item_id, product_id, lot_id, serial_id,
		location_id, quantity_to_pick, quantity_picked, status,
		assigned_to, assigned_at, completed_by, completed_at,
		short_pick_reason, created_by, created_date, updated_date, scenario_id
	FROM
		inventory.pick_tasks
	`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)
	sqldb.ApplyScenarioFilter(ctx, buf, data)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbTasks []pickTask
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbTasks); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	tasks, err := toBusPickTasks(dbTasks)
	if err != nil {
		return nil, fmt.Errorf("tobuspicktasks: %w", err)
	}

	return tasks, nil
}

// Count returns the total number of pick tasks matching the filter.
func (s *Store) Count(ctx context.Context, filter picktaskbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
	SELECT
		COUNT(1) AS count
	FROM
		inventory.pick_tasks
	`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)
	sqldb.ApplyScenarioFilter(ctx, buf, data)

	var count struct {
		Count int `db:"count"`
	}
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("namedqueryrow: %w", err)
	}

	return count.Count, nil
}

// QueryByID retrieves a single pick task by its ID.
func (s *Store) QueryByID(ctx context.Context, taskID uuid.UUID) (picktaskbus.PickTask, error) {
	data := map[string]any{
		"id": taskID.String(),
	}

	const q = `
	SELECT
		id, task_number, sales_order_id, sales_order_line_item_id, product_id, lot_id, serial_id,
		location_id, quantity_to_pick, quantity_picked, status,
		assigned_to, assigned_at, completed_by, completed_at,
		short_pick_reason, created_by, created_date, updated_date, scenario_id
	FROM
		inventory.pick_tasks
	WHERE
		id = :id
	`

	buf := bytes.NewBufferString(q)
	sqldb.ApplyScenarioFilter(ctx, buf, data)

	var dbTask pickTask
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &dbTask); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return picktaskbus.PickTask{}, picktaskbus.ErrNotFound
		}
		return picktaskbus.PickTask{}, fmt.Errorf("namedqueryrow: %w", err)
	}

	task, err := toBusPickTask(dbTask)
	if err != nil {
		return picktaskbus.PickTask{}, fmt.Errorf("tobuspicktask: %w", err)
	}

	return task, nil
}
