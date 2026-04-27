package ordersdb

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus"
	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus/types"
)

type dbOrder struct {
	ID                  uuid.UUID      `db:"id"`
	Number              string         `db:"number"`
	CustomerID          uuid.UUID      `db:"customer_id"`
	DueDate             time.Time      `db:"due_date"`
	FulfillmentStatusID uuid.UUID      `db:"order_fulfillment_status_id"`
	OrderDate           sql.NullTime   `db:"order_date"`
	BillingAddressID    *uuid.UUID     `db:"billing_address_id"`
	ShippingAddressID   *uuid.UUID     `db:"shipping_address_id"`
	AssignedTo          *uuid.UUID     `db:"assigned_to"`
	Subtotal            sql.NullString `db:"subtotal"`
	TaxRate             sql.NullString `db:"tax_rate"`
	TaxAmount           sql.NullString `db:"tax_amount"`
	ShippingCost        sql.NullString `db:"shipping_cost"`
	TotalAmount         sql.NullString `db:"total_amount"`
	CurrencyID          uuid.UUID      `db:"currency_id"`
	PaymentTermID       *uuid.UUID     `db:"payment_term_id"`
	Notes               sql.NullString `db:"notes"`
	Priority            string         `db:"priority"`
	CreatedBy           uuid.UUID      `db:"created_by"`
	UpdatedBy           uuid.UUID      `db:"updated_by"`
	CreatedDate         time.Time      `db:"created_date"`
	UpdatedDate         time.Time      `db:"updated_date"`
	ScenarioID          *uuid.UUID     `db:"scenario_id"`
}

func toBusOrder(db dbOrder) (ordersbus.Order, error) {
	subtotal, err := types.ParseMoney(db.Subtotal.String)
	if err != nil {
		return ordersbus.Order{}, fmt.Errorf("tobusorder: subtotal: %w", err)
	}

	taxRate, err := types.ParsePercentage(db.TaxRate.String)
	if err != nil {
		return ordersbus.Order{}, fmt.Errorf("tobusorder: tax_rate: %w", err)
	}

	taxAmount, err := types.ParseMoney(db.TaxAmount.String)
	if err != nil {
		return ordersbus.Order{}, fmt.Errorf("tobusorder: tax_amount: %w", err)
	}

	shippingCost, err := types.ParseMoney(db.ShippingCost.String)
	if err != nil {
		return ordersbus.Order{}, fmt.Errorf("tobusorder: shipping_cost: %w", err)
	}

	totalAmount, err := types.ParseMoney(db.TotalAmount.String)
	if err != nil {
		return ordersbus.Order{}, fmt.Errorf("tobusorder: total_amount: %w", err)
	}

	bus := ordersbus.Order{
		ID:                  db.ID,
		Number:              db.Number,
		CustomerID:          db.CustomerID,
		DueDate:             db.DueDate.In(time.Local),
		FulfillmentStatusID: db.FulfillmentStatusID,
		BillingAddressID:    db.BillingAddressID,
		ShippingAddressID:   db.ShippingAddressID,
		AssignedTo:          db.AssignedTo,
		Subtotal:            subtotal,
		TaxRate:             taxRate,
		TaxAmount:           taxAmount,
		ShippingCost:        shippingCost,
		TotalAmount:         totalAmount,
		CurrencyID:          db.CurrencyID,
		PaymentTermID:       db.PaymentTermID,
		Priority:            db.Priority,
		CreatedBy:           db.CreatedBy,
		UpdatedBy:           db.UpdatedBy,
		CreatedDate:         db.CreatedDate.In(time.Local),
		UpdatedDate:         db.UpdatedDate.In(time.Local),
		ScenarioID:          db.ScenarioID,
	}

	if db.OrderDate.Valid {
		bus.OrderDate = db.OrderDate.Time.In(time.Local)
	}

	if db.Notes.Valid {
		bus.Notes = db.Notes.String
	}

	return bus, nil
}

func toBusOrders(dbs []dbOrder) ([]ordersbus.Order, error) {
	orders := make([]ordersbus.Order, len(dbs))
	for i, db := range dbs {
		order, err := toBusOrder(db)
		if err != nil {
			return nil, err
		}
		orders[i] = order
	}
	return orders, nil
}

type dbOrderContainerBinding struct {
	ID               uuid.UUID  `db:"id"`
	OrderID          uuid.UUID  `db:"order_id"`
	ContainerLabelID uuid.UUID  `db:"container_label_id"`
	BoundAt          time.Time  `db:"bound_at"`
	UnboundAt        *time.Time `db:"unbound_at"`
	ScenarioID       *uuid.UUID `db:"scenario_id"`
}

func toBusBinding(r dbOrderContainerBinding) ordersbus.OrderContainerBinding {
	return ordersbus.OrderContainerBinding{
		ID:               r.ID,
		OrderID:          r.OrderID,
		ContainerLabelID: r.ContainerLabelID,
		BoundAt:          r.BoundAt.In(time.Local),
		UnboundAt:        r.UnboundAt,
		ScenarioID:       r.ScenarioID,
	}
}

func toBusBindings(rs []dbOrderContainerBinding) []ordersbus.OrderContainerBinding {
	bindings := make([]ordersbus.OrderContainerBinding, len(rs))
	for i, r := range rs {
		bindings[i] = toBusBinding(r)
	}
	return bindings
}

func toDBOrder(bus ordersbus.Order) dbOrder {
	db := dbOrder{
		ID:                  bus.ID,
		Number:              bus.Number,
		CustomerID:          bus.CustomerID,
		DueDate:             bus.DueDate.UTC(),
		FulfillmentStatusID: bus.FulfillmentStatusID,
		BillingAddressID:    bus.BillingAddressID,
		ShippingAddressID:   bus.ShippingAddressID,
		AssignedTo:          bus.AssignedTo,
		Subtotal:            bus.Subtotal.DBValue(),
		TaxRate:             bus.TaxRate.DBValue(),
		TaxAmount:           bus.TaxAmount.DBValue(),
		ShippingCost:        bus.ShippingCost.DBValue(),
		TotalAmount:         bus.TotalAmount.DBValue(),
		CurrencyID:          bus.CurrencyID,
		PaymentTermID:       bus.PaymentTermID,
		Priority:            bus.Priority,
		CreatedBy:           bus.CreatedBy,
		UpdatedBy:           bus.UpdatedBy,
		CreatedDate:         bus.CreatedDate.UTC(),
		UpdatedDate:         bus.UpdatedDate.UTC(),
		ScenarioID:          bus.ScenarioID,
	}

	if !bus.OrderDate.IsZero() {
		db.OrderDate = sql.NullTime{Time: bus.OrderDate.UTC(), Valid: true}
	}

	if bus.Notes != "" {
		db.Notes = sql.NullString{String: bus.Notes, Valid: true}
	}

	return db
}
