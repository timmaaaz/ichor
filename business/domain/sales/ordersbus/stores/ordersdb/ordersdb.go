package ordersdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
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
	  order_date, billing_address_id, shipping_address_id, assigned_to,
	  subtotal, tax_rate, tax_amount, shipping_cost, total_amount,
	  currency_id, payment_term_id, notes, priority,
	  created_by, updated_by, created_date, updated_date, scenario_id
    ) VALUES (
        :id, :number, :customer_id, :due_date, :order_fulfillment_status_id,
        :order_date, :billing_address_id, :shipping_address_id, :assigned_to,
        :subtotal, :tax_rate, :tax_amount, :shipping_cost, :total_amount,
        :currency_id, :payment_term_id, :notes, :priority,
        :created_by, :updated_by, :created_date, :updated_date, :scenario_id
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
        assigned_to = :assigned_to,
        subtotal = :subtotal,
        tax_rate = :tax_rate,
        tax_amount = :tax_amount,
        shipping_cost = :shipping_cost,
        total_amount = :total_amount,
        currency_id = :currency_id,
        payment_term_id = :payment_term_id,
        notes = :notes,
        priority = :priority,
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
			return fmt.Errorf("namedexeccontext %w", ordersbus.ErrUniqueEntry)
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
		order_date, billing_address_id, shipping_address_id, assigned_to,
		subtotal, tax_rate, tax_amount, shipping_cost, total_amount,
		currency_id, payment_term_id, notes, priority,
		created_by, updated_by, created_date, updated_date, scenario_id
    FROM
	    sales.orders
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

func (s *Store) QueryByID(ctx context.Context, statusID uuid.UUID) (ordersbus.Order, error) {
	data := map[string]any{
		"id": statusID.String(),
	}

	const q = `
    SELECT
        id, number, customer_id, due_date, order_fulfillment_status_id,
        order_date, billing_address_id, shipping_address_id, assigned_to,
        subtotal, tax_rate, tax_amount, shipping_cost, total_amount,
        currency_id, payment_term_id, notes, priority,
        created_by, updated_by, created_date, updated_date, scenario_id
    FROM
        sales.orders
    WHERE
        id = :id
    `

	buf := bytes.NewBufferString(q)
	sqldb.ApplyScenarioFilter(ctx, buf, data)

	var dbOrd dbOrder
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &dbOrd); err != nil {
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

// QueryByIDs retrieves a list of orders by their IDs.
func (s *Store) QueryByIDs(ctx context.Context, orderIDs []uuid.UUID) ([]ordersbus.Order, error) {
	if len(orderIDs) == 0 {
		return []ordersbus.Order{}, nil
	}

	uuidStrings := make([]string, len(orderIDs))
	for i, id := range orderIDs {
		uuidStrings[i] = id.String()
	}

	data := map[string]any{
		"order_ids": uuidStrings,
	}

	const q = `
	SELECT
		id, number, customer_id, due_date, order_fulfillment_status_id,
		order_date, billing_address_id, shipping_address_id, assigned_to,
		subtotal, tax_rate, tax_amount, shipping_cost, total_amount,
		currency_id, payment_term_id, notes, priority,
		created_by, updated_by, created_date, updated_date, scenario_id
	FROM
		sales.orders
	WHERE
		id IN (:order_ids)`

	buf := bytes.NewBufferString(q)
	sqldb.ApplyScenarioFilter(ctx, buf, data)

	var dbOrders []dbOrder
	if err := sqldb.NamedQuerySliceUsingIn(ctx, s.log, s.db, buf.String(), data, &dbOrders); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	busOrders, err := toBusOrders(dbOrders)
	if err != nil {
		return nil, fmt.Errorf("tobusorders: %w", err)
	}

	return busOrders, nil
}

// BindContainer creates a new active binding from order to container label.
// New active bindings leave unbound_at NULL; the EXCLUDE constraint
// one_active_binding_per_container enforces uniqueness over rows where
// unbound_at IS NULL. UnbindContainer is the only path that sets unbound_at.
// EXCLUDE-constraint violations come back as raw pq errors (NamedQueryStruct
// does not translate SQLSTATE codes); callers that need to distinguish must
// string-match err.Error().
func (s *Store) BindContainer(ctx context.Context, nb ordersbus.NewOrderContainerBinding) (ordersbus.OrderContainerBinding, error) {
	const q = `
	INSERT INTO inventory.order_container_bindings
		(id, order_id, container_label_id, scenario_id)
	VALUES
		(:id, :order_id, :container_label_id, :scenario_id)
	RETURNING id, order_id, container_label_id, bound_at, unbound_at, scenario_id`

	row := dbOrderContainerBinding{
		ID:               uuid.New(),
		OrderID:          nb.OrderID,
		ContainerLabelID: nb.ContainerLabelID,
		ScenarioID:       nb.ScenarioID,
	}

	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, row, &row); err != nil {
		// NamedQueryStruct does not translate PG SQLSTATE codes; EXCLUDE
		// violations on one_active_binding_per_container surface as raw pq
		// errors. Tests string-match err.Error() to assert the constraint.
		return ordersbus.OrderContainerBinding{}, fmt.Errorf("namedqueryrow: %w", err)
	}

	return toBusBinding(row), nil
}

// UnbindContainer marks an active binding as released by setting unbound_at = NOW().
// The WHERE clause includes "AND unbound_at IS NULL" so calling Unbind on a
// binding that's already unbound is INTENTIONALLY a silent no-op (idempotent
// resource release). This contract diverges from approvalrequestdb.Resolve's
// "translate zero-rows to ErrAlreadyResolved" pattern because container
// release is replay-safe by design — re-issuing "free this container" should
// not error.
func (s *Store) UnbindContainer(ctx context.Context, bindingID uuid.UUID) error {
	const q = `
	UPDATE inventory.order_container_bindings
	SET unbound_at = NOW()
	WHERE id = :id AND unbound_at IS NULL`

	args := struct {
		ID uuid.UUID `db:"id"`
	}{ID: bindingID}

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, args); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// QueryActiveBindingsByOrder returns all bindings for the given order whose
// unbound_at IS NULL, ordered by bound_at ascending.
func (s *Store) QueryActiveBindingsByOrder(ctx context.Context, orderID uuid.UUID) ([]ordersbus.OrderContainerBinding, error) {
	const q = `
	SELECT
		id, order_id, container_label_id, bound_at, unbound_at, scenario_id
	FROM
		inventory.order_container_bindings
	WHERE
		order_id = :order_id AND unbound_at IS NULL
	ORDER BY
		bound_at`

	args := struct {
		OrderID uuid.UUID `db:"order_id"`
	}{OrderID: orderID}

	var rows []dbOrderContainerBinding
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, args, &rows); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusBindings(rows), nil
}
