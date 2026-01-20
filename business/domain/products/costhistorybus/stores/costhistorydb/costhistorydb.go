package costhistorydb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/products/costhistorybus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for assets database access.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (costhistorybus.Storer, error) {
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

func (s *Store) Create(ctx context.Context, ch costhistorybus.CostHistory) error {
	const q = `
    INSERT INTO products.cost_history (
        id, product_id, cost_type, amount, currency_id,  effective_date, end_date, created_date, updated_date
    ) VALUES (
        :id, :product_id, :cost_type, :amount, :currency_id, :effective_date, :end_date, :created_date, :updated_date
    )
    `

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBCostHistory(ch)); err != nil {
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", costhistorybus.ErrForeignKeyViolation)
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", costhistorybus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

func (s *Store) Update(ctx context.Context, ch costhistorybus.CostHistory) error {
	const q = `
    UPDATE
        products.cost_history
    SET
        id = :id,
        product_id = :product_id,
        cost_type = :cost_type,
        amount = :amount,
        currency_id = :currency_id,
        effective_date = :effective_date,
        end_date = :end_date,
        updated_date = :updated_date
    WHERE
        id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBCostHistory(ch)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", costhistorybus.ErrUniqueEntry)
		}
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", costhistorybus.ErrForeignKeyViolation)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

// Delete removes an product cost from the database.
func (s *Store) Delete(ctx context.Context, ch costhistorybus.CostHistory) error {
	const q = `
	DELETE FROM
		products.cost_history
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBCostHistory(ch)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of cost histories from the database
func (s *Store) Query(ctx context.Context, filter costhistorybus.QueryFilter, orderBy order.By, page page.Page) ([]costhistorybus.CostHistory, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT 
		id, product_id, cost_type, amount, currency_id, effective_date, end_date, updated_date, created_date
	FROM
		products.cost_history
	`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var ch []costHistory
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &ch); err != nil {
		return nil, fmt.Errorf("namedselectcontext: %w", err)
	}

	return toBusCostHistories(ch)
}

// Count returns the number of productcosts in the database.
func (s *Store) Count(ctx context.Context, filter costhistorybus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
    SELECT
        COUNT(1) AS count
    FROM
        products.cost_history`

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

func (s *Store) QueryByID(ctx context.Context, costHistoryID uuid.UUID) (costhistorybus.CostHistory, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: costHistoryID.String(),
	}

	const q = `
	SELECT 
		id, product_id, cost_type, amount, currency_id, effective_date, end_date, updated_date, created_date
	FROM
		products.cost_history
	WHERE 
	    id = :id`

	var ch costHistory
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &ch); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return costhistorybus.CostHistory{}, costhistorybus.ErrNotFound
		}
		return costhistorybus.CostHistory{}, fmt.Errorf("querystruct: %w", err)
	}

	return toBusCostHistory(ch)
}
