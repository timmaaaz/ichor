package orderlineitemsdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus"
	"github.com/timmaaaz/ichor/business/domain/sales/orderlineitemsbus"
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

func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (orderlineitemsbus.Storer, error) {
	ec, err := sqldb.GetExtContext(tx)
	if err != nil {
		return nil, err
	}

	return &Store{
		log: s.log,
		db:  ec,
	}, nil
}

func (s *Store) Create(ctx context.Context, status orderlineitemsbus.OrderLineItem) error {
	const q = `
	INSERT INTO sales.order_line_items (
	  id, order_id, product_id, description, quantity, unit_price, discount, discount_type, line_total, line_item_fulfillment_statuses_id, created_by, created_date, updated_by, updated_date
    ) VALUES (
        :id, :order_id, :product_id, :description, :quantity, :unit_price, :discount, :discount_type, :line_total, :line_item_fulfillment_statuses_id, :created_by, :created_date, :updated_by, :updated_date
    )
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBOrderLineItem(status)); err != nil {
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext %w", orderlineitemsbus.ErrForeignKeyViolation)
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext %w", orderlineitemsbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext %w", err)
	}

	return nil

}

func (s *Store) Update(ctx context.Context, status orderlineitemsbus.OrderLineItem) error {
	const q = `
    UPDATE
        sales.order_line_items
    SET
       order_id = :order_id,
       product_id = :product_id,
       description = :description,
       quantity = :quantity,
       unit_price = :unit_price,
       discount = :discount,
       discount_type = :discount_type,
       line_total = :line_total,
       line_item_fulfillment_statuses_id = :line_item_fulfillment_statuses_id,
       created_by = :created_by,
       created_date = :created_date,
       updated_by = :updated_by,
       updated_date = :updated_date
    WHERE
        id = :id
    `

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBOrderLineItem(status)); err != nil {
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext %w", orderlineitemsbus.ErrForeignKeyViolation)
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext %w", transferorderbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext %w", err)
	}

	return nil
}

func (s *Store) Delete(ctx context.Context, status orderlineitemsbus.OrderLineItem) error {
	const q = `
    DELETE FROM sales.order_line_items
    WHERE id = :id
    `

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBOrderLineItem(status)); err != nil {
		return fmt.Errorf("namedexeccontext %w", err)
	}

	return nil
}

func (s *Store) Query(ctx context.Context, filter orderlineitemsbus.QueryFilter, orderBy order.By, page page.Page) ([]orderlineitemsbus.OrderLineItem, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
		id, order_id, product_id, description, quantity, unit_price, discount, discount_type, line_total, line_item_fulfillment_statuses_id, created_by, created_date, updated_by, updated_date
    FROM
	    sales.order_line_items
		`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbStatuses []orderLineItem
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbStatuses); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	items, err := toBusOrderLineItems(dbStatuses)
	if err != nil {
		return nil, fmt.Errorf("tobusorderlineitems: %w", err)
	}

	return items, nil
}

func (s *Store) Count(ctx context.Context, filter orderlineitemsbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
    SELECT
        COUNT(1) AS count
    FROM
        sales.order_line_items
    `

	var count struct {
		Count int `db:"count"`
	}
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &count); err != nil {
		return 0, fmt.Errorf("namedqueryrow: %w", err)
	}

	return count.Count, nil
}

func (s *Store) QueryByID(ctx context.Context, statusID uuid.UUID) (orderlineitemsbus.OrderLineItem, error) {
	data := struct {
		StatusID string `db:"id"`
	}{
		StatusID: statusID.String(),
	}

	const q = `
    SELECT
        id, order_id, product_id, description, quantity, unit_price, discount, discount_type, line_total, line_item_fulfillment_statuses_id, created_by, created_date, updated_by, updated_date
    FROM
        sales.order_line_items
    WHERE
        id = :id
    `

	var dbStatus orderLineItem
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbStatus); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return orderlineitemsbus.OrderLineItem{}, orderlineitemsbus.ErrNotFound
		}
		return orderlineitemsbus.OrderLineItem{}, fmt.Errorf("namedqueryrow: %w", err)
	}

	item, err := toBusOrderLineItem(dbStatus)
	if err != nil {
		return orderlineitemsbus.OrderLineItem{}, fmt.Errorf("tobusorderlineitem: %w", err)
	}

	return item, nil
}
