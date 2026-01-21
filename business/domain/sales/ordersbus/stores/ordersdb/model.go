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
	OrderDate           time.Time      `db:"order_date"`
	BillingAddressID    *uuid.UUID     `db:"billing_address_id"`
	ShippingAddressID   *uuid.UUID     `db:"shipping_address_id"`
	Subtotal            sql.NullString `db:"subtotal"`
	TaxRate             sql.NullString `db:"tax_rate"`
	TaxAmount           sql.NullString `db:"tax_amount"`
	ShippingCost        sql.NullString `db:"shipping_cost"`
	TotalAmount         sql.NullString `db:"total_amount"`
	CurrencyID          uuid.UUID      `db:"currency_id"`
	PaymentTermID       *uuid.UUID     `db:"payment_term_id"`
	Notes               string         `db:"notes"`
	CreatedBy           uuid.UUID      `db:"created_by"`
	UpdatedBy           uuid.UUID      `db:"updated_by"`
	CreatedDate         time.Time      `db:"created_date"`
	UpdatedDate         time.Time      `db:"updated_date"`
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

	return ordersbus.Order{
		ID:                  db.ID,
		Number:              db.Number,
		CustomerID:          db.CustomerID,
		DueDate:             db.DueDate.In(time.Local),
		FulfillmentStatusID: db.FulfillmentStatusID,
		OrderDate:           db.OrderDate.In(time.Local),
		BillingAddressID:    db.BillingAddressID,
		ShippingAddressID:   db.ShippingAddressID,
		Subtotal:            subtotal,
		TaxRate:             taxRate,
		TaxAmount:           taxAmount,
		ShippingCost:        shippingCost,
		TotalAmount:         totalAmount,
		CurrencyID:            db.CurrencyID,
		PaymentTermID:       db.PaymentTermID,
		Notes:               db.Notes,
		CreatedBy:           db.CreatedBy,
		UpdatedBy:           db.UpdatedBy,
		CreatedDate:         db.CreatedDate.In(time.Local),
		UpdatedDate:         db.UpdatedDate.In(time.Local),
	}, nil
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

func toDBOrder(bus ordersbus.Order) dbOrder {
	return dbOrder{
		ID:                  bus.ID,
		Number:              bus.Number,
		CustomerID:          bus.CustomerID,
		DueDate:             bus.DueDate.UTC(),
		FulfillmentStatusID: bus.FulfillmentStatusID,
		OrderDate:           bus.OrderDate.UTC(),
		BillingAddressID:    bus.BillingAddressID,
		ShippingAddressID:   bus.ShippingAddressID,
		Subtotal:            bus.Subtotal.DBValue(),
		TaxRate:             bus.TaxRate.DBValue(),
		TaxAmount:           bus.TaxAmount.DBValue(),
		ShippingCost:        bus.ShippingCost.DBValue(),
		TotalAmount:         bus.TotalAmount.DBValue(),
		CurrencyID:            bus.CurrencyID,
		PaymentTermID:       bus.PaymentTermID,
		Notes:               bus.Notes,
		CreatedBy:           bus.CreatedBy,
		UpdatedBy:           bus.UpdatedBy,
		CreatedDate:         bus.CreatedDate.UTC(),
		UpdatedDate:         bus.UpdatedDate.UTC(),
	}
}
