package metricsdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/quality/metricsbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for metrics database access.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (metricsbus.Storer, error) {
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

func (s *Store) Create(ctx context.Context, metric metricsbus.Metric) error {
	const q = `
	INSERT INTO quality_metrics (
		quality_metric_id, product_id, return_rate, defect_rate, measurement_period, created_date, updated_date
	) VALUES (
		:quality_metric_id, :product_id, :return_rate, :defect_rate, :measurement_period, :created_date, :updated_date
	)
	`
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBMetric(metric)); err != nil {
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", metricsbus.ErrForeignKeyViolation)
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", metricsbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

func (s *Store) Update(ctx context.Context, metric metricsbus.Metric) error {
	const q = `
	UPDATE
	    quality_metrics
	SET
		quality_metric_id = :quality_metric_id,
		product_id = :product_id,
		return_rate = :return_rate,
		defect_rate = :defect_rate,
		measurement_period = :measurement_period,
		updated_date = :updated_date
	WHERE 
	    quality_metric_id = :quality_metric_id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBMetric(metric)); err != nil {
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", metricsbus.ErrForeignKeyViolation)
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", metricsbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

func (s *Store) Delete(ctx context.Context, metric metricsbus.Metric) error {
	const q = `
    DELETE FROM
        quality_metrics
    WHERE
        quality_metric_id = :quality_metric_id
    `

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBMetric(metric)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

func (s *Store) Query(ctx context.Context, filter metricsbus.QueryFilter, orderBy order.By, page page.Page) ([]metricsbus.Metric, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
    SELECT
        quality_metric_id, product_id, return_rate, defect_rate, measurement_period, created_date, updated_date
    FROM
        quality_metrics
	`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbMetric []metric

	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbMetric); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusMetrics(dbMetric)
}

func (s *Store) Count(ctx context.Context, filter metricsbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
    SELECT
        COUNT(1) AS count
    FROM
        quality_metrics
    `

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}

	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("namedQueryStruct: %w", err)
	}

	return count.Count, nil
}

func (s *Store) QueryByID(ctx context.Context, id uuid.UUID) (metricsbus.Metric, error) {
	data := struct {
		ID string `db:"quality_metric_id"`
	}{
		ID: id.String(),
	}

	const q = `
    SELECT
        quality_metric_id, product_id, return_rate, defect_rate, measurement_period, created_date, updated_date
    FROM
        quality_metrics
	WHERE
		quality_metric_id = :quality_metric_id
	`

	var dbMetric metric

	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbMetric); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return metricsbus.Metric{}, metricsbus.ErrNotFound
		}
		return metricsbus.Metric{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusMetric(dbMetric)
}
