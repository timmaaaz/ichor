package ordersapp

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus"
	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus/types"
)

const dateFormat = "2006-01-02"

type QueryParams struct {
	Page    string
	Rows    string
	OrderBy string

	ID                  string
	Number              string
	CustomerID          string
	FulfillmentStatusID string
	BillingAddressID    string
	ShippingAddressID   string
	Currency            string
	PaymentTermID       string
	CreatedBy           string
	UpdatedBy           string
	StartDueDate        string
	EndDueDate          string
	StartOrderDate      string
	EndOrderDate        string
	StartCreatedDate    string
	EndCreatedDate      string
	StartUpdatedDate    string
	EndUpdatedDate      string
	// Monetary range filters
	MinSubtotal     string
	MaxSubtotal     string
	MinTaxAmount    string
	MaxTaxAmount    string
	MinShippingCost string
	MaxShippingCost string
	MinTotalAmount  string
	MaxTotalAmount  string
}

type Order struct {
	ID                  string `json:"id"`
	Number              string `json:"number"`
	CustomerID          string `json:"customer_id"`
	DueDate             string `json:"due_date"`
	FulfillmentStatusID string `json:"fulfillment_status_id"`
	OrderDate           string `json:"order_date"`
	BillingAddressID    string `json:"billing_address_id,omitempty"`
	ShippingAddressID   string `json:"shipping_address_id,omitempty"`
	Subtotal            string `json:"subtotal"`
	TaxRate             string `json:"tax_rate"`
	TaxAmount           string `json:"tax_amount"`
	ShippingCost        string `json:"shipping_cost"`
	TotalAmount         string `json:"total_amount"`
	Currency            string `json:"currency"`
	PaymentTermID       string `json:"payment_term_id,omitempty"`
	Notes               string `json:"notes"`
	CreatedBy           string `json:"created_by"`
	UpdatedBy           string `json:"updated_by"`
	CreatedDate         string `json:"created_date"`
	UpdatedDate         string `json:"updated_date"`
}

func (app Order) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppOrder(bus ordersbus.Order) Order {
	app := Order{
		ID:                  bus.ID.String(),
		Number:              bus.Number,
		CustomerID:          bus.CustomerID.String(),
		DueDate:             bus.DueDate.Format(dateFormat),
		FulfillmentStatusID: bus.FulfillmentStatusID.String(),
		OrderDate:           bus.OrderDate.Format(dateFormat),
		Subtotal:            bus.Subtotal.Value(),
		TaxRate:             bus.TaxRate.Value(),
		TaxAmount:           bus.TaxAmount.Value(),
		ShippingCost:        bus.ShippingCost.Value(),
		TotalAmount:         bus.TotalAmount.Value(),
		Currency:            bus.Currency,
		Notes:               bus.Notes,
		CreatedBy:           bus.CreatedBy.String(),
		UpdatedBy:           bus.UpdatedBy.String(),
		CreatedDate:         bus.CreatedDate.Format(time.RFC3339),
		UpdatedDate:         bus.UpdatedDate.Format(time.RFC3339),
	}

	// Handle nullable UUIDs
	if bus.BillingAddressID != nil {
		app.BillingAddressID = bus.BillingAddressID.String()
	}
	if bus.ShippingAddressID != nil {
		app.ShippingAddressID = bus.ShippingAddressID.String()
	}
	if bus.PaymentTermID != nil {
		app.PaymentTermID = bus.PaymentTermID.String()
	}

	return app
}

func ToAppOrders(bus []ordersbus.Order) []Order {
	appStatuses := make([]Order, len(bus))
	for i, status := range bus {
		appStatuses[i] = ToAppOrder(status)
	}
	return appStatuses
}

type NewOrder struct {
	Number              string  `json:"number" validate:"required"`
	CustomerID          string  `json:"customer_id" validate:"required,uuid4"`
	DueDate             string  `json:"due_date" validate:"required"`
	FulfillmentStatusID string  `json:"fulfillment_status_id" validate:"required,uuid4"`
	OrderDate           string  `json:"order_date" validate:"required"`
	BillingAddressID    string  `json:"billing_address_id" validate:"omitempty,uuid4"`
	ShippingAddressID   string  `json:"shipping_address_id" validate:"omitempty,uuid4"`
	Subtotal            string  `json:"subtotal"`
	TaxRate             string  `json:"tax_rate"`
	TaxAmount           string  `json:"tax_amount"`
	ShippingCost        string  `json:"shipping_cost"`
	TotalAmount         string  `json:"total_amount"`
	Currency            string  `json:"currency"`
	PaymentTermID       string  `json:"payment_term_id" validate:"omitempty,uuid4"`
	Notes               string  `json:"notes"`
	CreatedBy           string  `json:"created_by" validate:"required,uuid4"`
	CreatedDate         *string `json:"created_date"` // Optional: for seeding/import
}

func (app *NewOrder) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewOrder) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusNewOrder(app NewOrder) (ordersbus.NewOrder, error) {
	customerID, err := uuid.Parse(app.CustomerID)
	if err != nil {
		return ordersbus.NewOrder{}, errs.Newf(errs.InvalidArgument, "parse customerID: %s", err)
	}

	dueDate, err := time.Parse(dateFormat, app.DueDate)
	if err != nil {
		return ordersbus.NewOrder{}, errs.Newf(errs.InvalidArgument, "parse dueDate: %s", err)
	}

	fulfillmentStatusID, err := uuid.Parse(app.FulfillmentStatusID)
	if err != nil {
		return ordersbus.NewOrder{}, errs.Newf(errs.InvalidArgument, "parse fulfillmentStatusID: %s", err)
	}

	orderDate, err := time.Parse(dateFormat, app.OrderDate)
	if err != nil {
		return ordersbus.NewOrder{}, errs.Newf(errs.InvalidArgument, "parse orderDate: %s", err)
	}

	createdBy, err := uuid.Parse(app.CreatedBy)
	if err != nil {
		return ordersbus.NewOrder{}, errs.Newf(errs.InvalidArgument, "parse createdBy: %s", err)
	}

	// Parse Money fields
	subtotal, err := types.ParseMoney(app.Subtotal)
	if err != nil {
		return ordersbus.NewOrder{}, errs.Newf(errs.InvalidArgument, "parse subtotal: %s", err)
	}

	taxRate, err := types.ParseMoney(app.TaxRate)
	if err != nil {
		return ordersbus.NewOrder{}, errs.Newf(errs.InvalidArgument, "parse taxRate: %s", err)
	}

	taxAmount, err := types.ParseMoney(app.TaxAmount)
	if err != nil {
		return ordersbus.NewOrder{}, errs.Newf(errs.InvalidArgument, "parse taxAmount: %s", err)
	}

	shippingCost, err := types.ParseMoney(app.ShippingCost)
	if err != nil {
		return ordersbus.NewOrder{}, errs.Newf(errs.InvalidArgument, "parse shippingCost: %s", err)
	}

	totalAmount, err := types.ParseMoney(app.TotalAmount)
	if err != nil {
		return ordersbus.NewOrder{}, errs.Newf(errs.InvalidArgument, "parse totalAmount: %s", err)
	}

	// Parse nullable UUIDs
	var billingAddrID *uuid.UUID
	if app.BillingAddressID != "" {
		id, err := uuid.Parse(app.BillingAddressID)
		if err != nil {
			return ordersbus.NewOrder{}, errs.Newf(errs.InvalidArgument, "parse billingAddressID: %s", err)
		}
		billingAddrID = &id
	}

	var shippingAddrID *uuid.UUID
	if app.ShippingAddressID != "" {
		id, err := uuid.Parse(app.ShippingAddressID)
		if err != nil {
			return ordersbus.NewOrder{}, errs.Newf(errs.InvalidArgument, "parse shippingAddressID: %s", err)
		}
		shippingAddrID = &id
	}

	bus := ordersbus.NewOrder{
		Number:              app.Number,
		CustomerID:          customerID,
		DueDate:             dueDate,
		FulfillmentStatusID: fulfillmentStatusID,
		OrderDate:           orderDate,
		BillingAddressID:    billingAddrID,
		ShippingAddressID:   shippingAddrID,
		Subtotal:            subtotal,
		TaxRate:             taxRate,
		TaxAmount:           taxAmount,
		ShippingCost:        shippingCost,
		TotalAmount:         totalAmount,
		Currency:            app.Currency,
		Notes:               app.Notes,
		CreatedBy:           createdBy,
		// CreatedDate: nil by default - API always uses server time
	}

	// Parse optional PaymentTermID
	if app.PaymentTermID != "" {
		id, err := uuid.Parse(app.PaymentTermID)
		if err != nil {
			return ordersbus.NewOrder{}, errs.Newf(errs.InvalidArgument, "parse paymentTermID: %s", err)
		}
		bus.PaymentTermID = &id
	}

	// Handle optional CreatedDate (for imports/admin tools only)
	if app.CreatedDate != nil && *app.CreatedDate != "" {
		createdDate, err := time.Parse(time.RFC3339, *app.CreatedDate)
		if err != nil {
			return ordersbus.NewOrder{}, errs.Newf(errs.InvalidArgument, "parse createdDate: %s", err)
		}
		bus.CreatedDate = &createdDate
	}

	return bus, nil
}

type UpdateOrder struct {
	Number              *string `json:"number" validate:"omitempty"`
	CustomerID          *string `json:"customer_id" validate:"omitempty,uuid4"`
	DueDate             *string `json:"due_date" validate:"omitempty"`
	FulfillmentStatusID *string `json:"fulfillment_status_id" validate:"omitempty,uuid4"`
	OrderDate           *string `json:"order_date" validate:"omitempty"`
	BillingAddressID    *string `json:"billing_address_id" validate:"omitempty,uuid4"`
	ShippingAddressID   *string `json:"shipping_address_id" validate:"omitempty,uuid4"`
	Subtotal            *string `json:"subtotal" validate:"omitempty"`
	TaxRate             *string `json:"tax_rate" validate:"omitempty"`
	TaxAmount           *string `json:"tax_amount" validate:"omitempty"`
	ShippingCost        *string `json:"shipping_cost" validate:"omitempty"`
	TotalAmount         *string `json:"total_amount" validate:"omitempty"`
	Currency            *string `json:"currency" validate:"omitempty"`
	PaymentTermID       *string `json:"payment_term_id" validate:"omitempty,uuid4"`
	Notes               *string `json:"notes" validate:"omitempty"`
	UpdatedBy           *string `json:"updated_by" validate:"omitempty,uuid4"`
}

func (app *UpdateOrder) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app UpdateOrder) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusUpdateOrder(app UpdateOrder) (ordersbus.UpdateOrder, error) {
	var customerID *uuid.UUID
	if app.CustomerID != nil {
		id, err := uuid.Parse(*app.CustomerID)
		if err != nil {
			return ordersbus.UpdateOrder{}, errs.Newf(errs.InvalidArgument, "parse customerID: %s", err)
		}
		customerID = &id
	}

	var dueDate *time.Time
	if app.DueDate != nil {
		t, err := time.Parse(dateFormat, *app.DueDate)
		if err != nil {
			return ordersbus.UpdateOrder{}, errs.Newf(errs.InvalidArgument, "parse dueDate: %s", err)
		}
		dueDate = &t
	}

	var fulfillmentStatusID *uuid.UUID
	if app.FulfillmentStatusID != nil {
		id, err := uuid.Parse(*app.FulfillmentStatusID)
		if err != nil {
			return ordersbus.UpdateOrder{}, errs.Newf(errs.InvalidArgument, "parse fulfillmentStatusID: %s", err)
		}
		fulfillmentStatusID = &id
	}

	var orderDate *time.Time
	if app.OrderDate != nil {
		t, err := time.Parse(dateFormat, *app.OrderDate)
		if err != nil {
			return ordersbus.UpdateOrder{}, errs.Newf(errs.InvalidArgument, "parse orderDate: %s", err)
		}
		orderDate = &t
	}

	var billingAddrID *uuid.UUID
	if app.BillingAddressID != nil && *app.BillingAddressID != "" {
		id, err := uuid.Parse(*app.BillingAddressID)
		if err != nil {
			return ordersbus.UpdateOrder{}, errs.Newf(errs.InvalidArgument, "parse billingAddressID: %s", err)
		}
		billingAddrID = &id
	}

	var shippingAddrID *uuid.UUID
	if app.ShippingAddressID != nil && *app.ShippingAddressID != "" {
		id, err := uuid.Parse(*app.ShippingAddressID)
		if err != nil {
			return ordersbus.UpdateOrder{}, errs.Newf(errs.InvalidArgument, "parse shippingAddressID: %s", err)
		}
		shippingAddrID = &id
	}

	// Parse optional Money fields
	var subtotal *types.Money
	if app.Subtotal != nil {
		m, err := types.ParseMoneyPtr(*app.Subtotal)
		if err != nil {
			return ordersbus.UpdateOrder{}, errs.Newf(errs.InvalidArgument, "parse subtotal: %s", err)
		}
		subtotal = m
	}

	var taxRate *types.Money
	if app.TaxRate != nil {
		m, err := types.ParseMoneyPtr(*app.TaxRate)
		if err != nil {
			return ordersbus.UpdateOrder{}, errs.Newf(errs.InvalidArgument, "parse taxRate: %s", err)
		}
		taxRate = m
	}

	var taxAmount *types.Money
	if app.TaxAmount != nil {
		m, err := types.ParseMoneyPtr(*app.TaxAmount)
		if err != nil {
			return ordersbus.UpdateOrder{}, errs.Newf(errs.InvalidArgument, "parse taxAmount: %s", err)
		}
		taxAmount = m
	}

	var shippingCost *types.Money
	if app.ShippingCost != nil {
		m, err := types.ParseMoneyPtr(*app.ShippingCost)
		if err != nil {
			return ordersbus.UpdateOrder{}, errs.Newf(errs.InvalidArgument, "parse shippingCost: %s", err)
		}
		shippingCost = m
	}

	var totalAmount *types.Money
	if app.TotalAmount != nil {
		m, err := types.ParseMoneyPtr(*app.TotalAmount)
		if err != nil {
			return ordersbus.UpdateOrder{}, errs.Newf(errs.InvalidArgument, "parse totalAmount: %s", err)
		}
		totalAmount = m
	}

	var updatedBy *uuid.UUID
	if app.UpdatedBy != nil {
		id, err := uuid.Parse(*app.UpdatedBy)
		if err != nil {
			return ordersbus.UpdateOrder{}, errs.Newf(errs.InvalidArgument, "parse updatedBy: %s", err)
		}
		updatedBy = &id
	}

	var paymentTermID *uuid.UUID
	if app.PaymentTermID != nil && *app.PaymentTermID != "" {
		id, err := uuid.Parse(*app.PaymentTermID)
		if err != nil {
			return ordersbus.UpdateOrder{}, errs.Newf(errs.InvalidArgument, "parse paymentTermID: %s", err)
		}
		paymentTermID = &id
	}

	bus := ordersbus.UpdateOrder{
		Number:              app.Number,
		CustomerID:          customerID,
		DueDate:             dueDate,
		FulfillmentStatusID: fulfillmentStatusID,
		OrderDate:           orderDate,
		BillingAddressID:    billingAddrID,
		ShippingAddressID:   shippingAddrID,
		Subtotal:            subtotal,
		TaxRate:             taxRate,
		TaxAmount:           taxAmount,
		ShippingCost:        shippingCost,
		TotalAmount:         totalAmount,
		Currency:            app.Currency,
		PaymentTermID:       paymentTermID,
		Notes:               app.Notes,
		UpdatedBy:           updatedBy,
	}
	return bus, nil
}
