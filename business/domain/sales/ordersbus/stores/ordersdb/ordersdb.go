package ordersdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus"
	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus"
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

func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (ordersbus.Storer, error) {
	ec, err := sqldb.GetExtContext(tx)
	if err != nil {
		return nil, err
	}

	return &Store{
		log: s.log,
		db:  ec,
	}, nil
}

func (s *Store) Create(ctx context.Context, status ordersbus.Order) error {
	const q = `
	INSERT INTO sales.orders (
	  id, number, customer_id, due_date, order_fulfillment_status_id,
	  order_date, billing_address_id, shipping_address_id,
	  subtotal, tax_rate, tax_amount, shipping_cost, total_amount,
	  currency, payment_term_id, notes,
	  created_by, updated_by, created_date, updated_date
    ) VALUES (
        :id, :number, :customer_id, :due_date, :order_fulfillment_status_id,
        :order_date, :billing_address_id, :shipping_address_id,
        :subtotal, :tax_rate, :tax_amount, :shipping_cost, :total_amount,
        :currency, :payment_term_id, :notes,
        :created_by, :updated_by, :created_date, :updated_date
    )
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBOrder(status)); err != nil {
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext %w", ordersbus.ErrForeignKeyViolation)
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext %w", ordersbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext %w", err)
	}

	return nil

}

func (s *Store) Update(ctx context.Context, status ordersbus.Order) error {
	const q = `
    UPDATE
        sales.orders
    SET
        number = :number,
        customer_id = :customer_id,
        due_date = :due_date,
        order_fulfillment_status_id = :order_fulfillment_status_id,
        order_date = :order_date,
        billing_address_id = :billing_address_id,
        shipping_address_id = :shipping_address_id,
        subtotal = :subtotal,
        tax_rate = :tax_rate,
        tax_amount = :tax_amount,
        shipping_cost = :shipping_cost,
        total_amount = :total_amount,
        currency = :currency,
        payment_term_id = :payment_term_id,
        notes = :notes,
        created_by = :created_by,
        updated_by = :updated_by,
        created_date = :created_date,
        updated_date = :updated_date
    WHERE
        id = :id
    `

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBOrder(status)); err != nil {
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext %w", ordersbus.ErrForeignKeyViolation)
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext %w", transferorderbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext %w", err)
	}

	return nil
}

func (s *Store) Delete(ctx context.Context, status ordersbus.Order) error {
	const q = `
    DELETE FROM sales.orders
    WHERE id = :id
    `

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBOrder(status)); err != nil {
		return fmt.Errorf("namedexeccontext %w", err)
	}

	return nil
}

func (s *Store) Query(ctx context.Context, filter ordersbus.QueryFilter, orderBy order.By, page page.Page) ([]ordersbus.Order, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
		id, number, customer_id, due_date, order_fulfillment_status_id,
		order_date, billing_address_id, shipping_address_id,
		subtotal, tax_rate, tax_amount, shipping_cost, total_amount,
		currency, payment_term_id, notes,
		created_by, updated_by, created_date, updated_date
    FROM
	    sales.orders
		`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbStatuses []dbOrder
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbStatuses); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	orders, err := toBusOrders(dbStatuses)
	if err != nil {
		return nil, fmt.Errorf("tobusorders: %w", err)
	}

	return orders, nil
}

func (s *Store) Count(ctx context.Context, filter ordersbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
    SELECT
        COUNT(1) AS count
    FROM
        sales.orders
    `

	var count struct {
		Count int `db:"count"`
	}
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &count); err != nil {
		return 0, fmt.Errorf("namedqueryrow: %w", err)
	}

	return count.Count, nil
}

func (s *Store) QueryByID(ctx context.Context, statusID uuid.UUID) (ordersbus.Order, error) {
	data := struct {
		StatusID string `db:"id"`
	}{
		StatusID: statusID.String(),
	}

	const q = `
    SELECT
        id, number, customer_id, due_date, order_fulfillment_status_id,
        order_date, billing_address_id, shipping_address_id,
        subtotal, tax_rate, tax_amount, shipping_cost, total_amount,
        currency, payment_term_id, notes,
        created_by, updated_by, created_date, updated_date
    FROM
        sales.orders
    WHERE
        id = :id
    `

	var dbOrd dbOrder
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbOrd); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return ordersbus.Order{}, ordersbus.ErrNotFound
		}
		return ordersbus.Order{}, fmt.Errorf("namedqueryrow: %w", err)
	}

	order, err := toBusOrder(dbOrd)
	if err != nil {
		return ordersbus.Order{}, fmt.Errorf("tobusorder: %w", err)
	}

	return order, nil
}
