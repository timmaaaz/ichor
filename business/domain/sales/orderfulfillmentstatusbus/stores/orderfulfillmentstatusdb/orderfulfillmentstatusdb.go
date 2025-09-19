package orderfulfillmentstatusdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus"
	"github.com/timmaaaz/ichor/business/domain/sales/orderfulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

type Store struct {
	log *logger.Logger
	db  sqlx.ExtContext
}

func NewStore(log *logger.Logger, db *sqlx.DB) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (orderfulfillmentstatusbus.Storer, error) {
	ec, err := sqldb.GetExtContext(tx)
	if err != nil {
		return nil, err
	}

	return &Store{
		log: s.log,
		db:  ec,
	}, nil
}

func (s *Store) Create(ctx context.Context, status orderfulfillmentstatusbus.OrderFulfillmentStatus) error {
	const q = `
	INSERT INTO sales.order_fulfillment_statuses (
	  id, name, description
    ) VALUES (
        :id, :name, :description
    )
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBOrderFulfillmentStatus(status)); err != nil {
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext %w", orderfulfillmentstatusbus.ErrForeignKeyViolation)
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext %w", orderfulfillmentstatusbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext %w", err)
	}

	return nil

}

func (s *Store) Update(ctx context.Context, status orderfulfillmentstatusbus.OrderFulfillmentStatus) error {
	const q = `
    UPDATE
        sales.order_fulfillment_statuses
    SET
        name = :name,
        description = :description
    WHERE
        id = :id
    `

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBOrderFulfillmentStatus(status)); err != nil {
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext %w", orderfulfillmentstatusbus.ErrForeignKeyViolation)
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext %w", transferorderbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext %w", err)
	}

	return nil
}

func (s *Store) Delete(ctx context.Context, status orderfulfillmentstatusbus.OrderFulfillmentStatus) error {
	const q = `
    DELETE FROM sales.order_fulfillment_statuses
    WHERE id = :id
    `

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBOrderFulfillmentStatus(status)); err != nil {
		return fmt.Errorf("namedexeccontext %w", err)
	}

	return nil
}

func (s *Store) Query(ctx context.Context, filter orderfulfillmentstatusbus.QueryFilter, orderBy order.By, page page.Page) ([]orderfulfillmentstatusbus.OrderFulfillmentStatus, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
		id, name, description
    FROM
	    sales.order_fulfillment_statuses
		`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbStatuses []orderFulfillmentStatus
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbStatuses); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusOrderFulfillmentStatuses(dbStatuses), nil
}

func (s *Store) Count(ctx context.Context, filter orderfulfillmentstatusbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
    SELECT
        COUNT(1) AS count
    FROM
        sales.order_fulfillment_statuses
    `

	var count struct {
		Count int `db:"count"`
	}
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &count); err != nil {
		return 0, fmt.Errorf("namedqueryrow: %w", err)
	}

	return count.Count, nil
}

func (s *Store) QueryByID(ctx context.Context, statusID uuid.UUID) (orderfulfillmentstatusbus.OrderFulfillmentStatus, error) {
	data := struct {
		StatusID string `db:"id"`
	}{
		StatusID: statusID.String(),
	}

	const q = `
    SELECT
        id, name, description
    FROM
        sales.order_fulfillment_statuses
    WHERE
        id = :id
    `

	var dbStatus orderFulfillmentStatus
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbStatus); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return orderfulfillmentstatusbus.OrderFulfillmentStatus{}, orderfulfillmentstatusbus.ErrNotFound
		}
		return orderfulfillmentstatusbus.OrderFulfillmentStatus{}, fmt.Errorf("namedqueryrow: %w", err)
	}

	return toBusOrderFulfillmentStatus(dbStatus), nil
}
