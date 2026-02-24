package putawaytaskdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/inventory/putawaytaskbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for put-away task database access.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (putawaytaskbus.Storer, error) {
	ec, err := sqldb.GetExtContext(tx)
	if err != nil {
		return nil, err
	}

	return &Store{
		log: s.log,
		db:  ec,
	}, nil
}

// Create inserts a new put-away task into the database.
func (s *Store) Create(ctx context.Context, task putawaytaskbus.PutAwayTask) error {
	const q = `
	INSERT INTO inventory.put_away_tasks
		(id, product_id, location_id, quantity, reference_number, status,
		 assigned_to, assigned_at, completed_by, completed_at,
		 created_by, created_date, updated_date)
	VALUES
		(:id, :product_id, :location_id, :quantity, :reference_number, :status,
		 :assigned_to, :assigned_at, :completed_by, :completed_at,
		 :created_by, :created_date, :updated_date)
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBPutAwayTask(task)); err != nil {
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", putawaytaskbus.ErrForeignKeyViolation)
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", putawaytaskbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update modifies an existing put-away task in the database.
func (s *Store) Update(ctx context.Context, task putawaytaskbus.PutAwayTask) error {
	const q = `
	UPDATE inventory.put_away_tasks
	SET
		product_id       = :product_id,
		location_id      = :location_id,
		quantity         = :quantity,
		reference_number = :reference_number,
		status           = :status,
		assigned_to      = :assigned_to,
		assigned_at      = :assigned_at,
		completed_by     = :completed_by,
		completed_at     = :completed_at,
		updated_date     = :updated_date
	WHERE
		id = :id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBPutAwayTask(task)); err != nil {
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", putawaytaskbus.ErrForeignKeyViolation)
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", putawaytaskbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes a put-away task from the database.
func (s *Store) Delete(ctx context.Context, task putawaytaskbus.PutAwayTask) error {
	const q = `
	DELETE FROM inventory.put_away_tasks
	WHERE id = :id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBPutAwayTask(task)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of put-away tasks from the database.
func (s *Store) Query(ctx context.Context, filter putawaytaskbus.QueryFilter, orderBy order.By, page page.Page) ([]putawaytaskbus.PutAwayTask, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
		id, product_id, location_id, quantity, reference_number, status,
		assigned_to, assigned_at, completed_by, completed_at,
		created_by, created_date, updated_date
	FROM
		inventory.put_away_tasks
	`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbTasks []putAwayTask
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbTasks); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	tasks, err := toBusPutAwayTasks(dbTasks)
	if err != nil {
		return nil, fmt.Errorf("tobusputawaytasks: %w", err)
	}

	return tasks, nil
}

// Count returns the total number of put-away tasks matching the filter.
func (s *Store) Count(ctx context.Context, filter putawaytaskbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
	SELECT
		COUNT(1) AS count
	FROM
		inventory.put_away_tasks
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

// QueryByID retrieves a single put-away task by its ID.
func (s *Store) QueryByID(ctx context.Context, taskID uuid.UUID) (putawaytaskbus.PutAwayTask, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: taskID.String(),
	}

	const q = `
	SELECT
		id, product_id, location_id, quantity, reference_number, status,
		assigned_to, assigned_at, completed_by, completed_at,
		created_by, created_date, updated_date
	FROM
		inventory.put_away_tasks
	WHERE
		id = :id
	`

	var dbTask putAwayTask
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbTask); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return putawaytaskbus.PutAwayTask{}, putawaytaskbus.ErrNotFound
		}
		return putawaytaskbus.PutAwayTask{}, fmt.Errorf("namedqueryrow: %w", err)
	}

	task, err := toBusPutAwayTask(dbTask)
	if err != nil {
		return putawaytaskbus.PutAwayTask{}, fmt.Errorf("tobusputawaytask: %w", err)
	}

	return task, nil
}
