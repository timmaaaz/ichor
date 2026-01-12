package orderlineitemsapp

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/sales/orderlineitemsbus"
	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus/types"
)

type QueryParams struct {
	Page    string
	Rows    string
	OrderBy string

	ID                            string
	OrderID                       string
	ProductID                     string
	Description                   string
	Quantity                      string
	UnitPrice                     string
	Discount                      string
	DiscountType                  string
	LineTotal                     string
	LineItemFulfillmentStatusesID string
	CreatedBy                     string
	StartCreatedDate              string
	EndCreatedDate                string
	UpdatedBy                     string
	StartUpdatedDate              string
	EndUpdatedDate                string
}

type OrderLineItem struct {
	ID                            string `json:"id"`
	OrderID                       string `json:"order_id"`
	ProductID                     string `json:"product_id"`
	Description                   string `json:"description"`
	Quantity                      string `json:"quantity"`
	UnitPrice                     string `json:"unit_price"`
	Discount                      string `json:"discount"`
	DiscountType                  string `json:"discount_type"`
	LineTotal                     string `json:"line_total"`
	LineItemFulfillmentStatusesID string `json:"line_item_fulfillment_statuses_id"`
	CreatedBy                     string `json:"created_by"`
	CreatedDate                   string `json:"created_date"`
	UpdatedBy                     string `json:"updated_by"`
	UpdatedDate                   string `json:"updated_date"`
}

func (app OrderLineItem) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppOrderLineItem(bus orderlineitemsbus.OrderLineItem) OrderLineItem {
	return OrderLineItem{
		ID:                            bus.ID.String(),
		OrderID:                       bus.OrderID.String(),
		ProductID:                     bus.ProductID.String(),
		Description:                   bus.Description,
		Quantity:                      strconv.Itoa(bus.Quantity),
		UnitPrice:                     bus.UnitPrice.Value(),
		Discount:                      bus.Discount.Value(),
		DiscountType:                  bus.DiscountType,
		LineTotal:                     bus.LineTotal.Value(),
		LineItemFulfillmentStatusesID: bus.LineItemFulfillmentStatusesID.String(),
		CreatedBy:                     bus.CreatedBy.String(),
		CreatedDate:                   bus.CreatedDate.Format(time.RFC3339),
		UpdatedBy:                     bus.UpdatedBy.String(),
		UpdatedDate:                   bus.UpdatedDate.Format(time.RFC3339),
	}
}

func ToAppOrderLineItems(bus []orderlineitemsbus.OrderLineItem) []OrderLineItem {
	appStatuses := make([]OrderLineItem, len(bus))
	for i, status := range bus {
		appStatuses[i] = ToAppOrderLineItem(status)
	}
	return appStatuses
}

type NewOrderLineItem struct {
	OrderID                       string  `json:"order_id" validate:"required,uuid4"`
	ProductID                     string  `json:"product_id" validate:"required,uuid4"`
	Description                   string  `json:"description" validate:"omitempty"`
	Quantity                      string  `json:"quantity" validate:"required,numeric"`
	UnitPrice                     string  `json:"unit_price" validate:"omitempty"`
	Discount                      string  `json:"discount" validate:"omitempty"`
	DiscountType                  string  `json:"discount_type" validate:"omitempty,oneof=flat percent"`
	LineTotal                     string  `json:"line_total" validate:"omitempty"`
	LineItemFulfillmentStatusesID string  `json:"line_item_fulfillment_statuses_id" validate:"required,uuid4"`
	CreatedBy                     string  `json:"created_by" validate:"required,uuid4"`
	CreatedDate                   *string `json:"created_date"` // Optional: for seeding/import
}

func (app *NewOrderLineItem) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewOrderLineItem) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusNewOrderLineItem(app NewOrderLineItem) (orderlineitemsbus.NewOrderLineItem, error) {
	orderID, err := uuid.Parse(app.OrderID)
	if err != nil {
		return orderlineitemsbus.NewOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse orderID: %s", err)
	}

	productID, err := uuid.Parse(app.ProductID)
	if err != nil {
		return orderlineitemsbus.NewOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse productID: %s", err)
	}

	quantity, err := strconv.Atoi(app.Quantity)
	if err != nil {
		return orderlineitemsbus.NewOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse quantity: %s", err)
	}

	unitPrice, err := types.ParseMoney(app.UnitPrice)
	if err != nil {
		return orderlineitemsbus.NewOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse unit_price: %s", err)
	}

	discount, err := types.ParseMoney(app.Discount)
	if err != nil {
		return orderlineitemsbus.NewOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse discount: %s", err)
	}

	// Validate discount_type if provided
	discountType := app.DiscountType
	if discountType != "" && discountType != "flat" && discountType != "percent" {
		return orderlineitemsbus.NewOrderLineItem{}, errs.Newf(errs.InvalidArgument, "discount_type must be 'flat' or 'percent'")
	}

	lineTotal, err := types.ParseMoney(app.LineTotal)
	if err != nil {
		return orderlineitemsbus.NewOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse line_total: %s", err)
	}

	lineItemFulfillmentStatusesID, err := uuid.Parse(app.LineItemFulfillmentStatusesID)
	if err != nil {
		return orderlineitemsbus.NewOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse lineItemFulfillmentStatusesID: %s", err)
	}

	createdBy, err := uuid.Parse(app.CreatedBy)
	if err != nil {
		return orderlineitemsbus.NewOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse createdBy: %s", err)
	}

	bus := orderlineitemsbus.NewOrderLineItem{
		OrderID:                       orderID,
		ProductID:                     productID,
		Description:                   app.Description,
		Quantity:                      quantity,
		UnitPrice:                     unitPrice,
		Discount:                      discount,
		DiscountType:                  discountType,
		LineTotal:                     lineTotal,
		LineItemFulfillmentStatusesID: lineItemFulfillmentStatusesID,
		CreatedBy:                     createdBy,
		// CreatedDate: nil by default - API always uses server time
	}

	// Handle optional CreatedDate (for imports/admin tools only)
	if app.CreatedDate != nil && *app.CreatedDate != "" {
		createdDate, err := time.Parse(time.RFC3339, *app.CreatedDate)
		if err != nil {
			return orderlineitemsbus.NewOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse createdDate: %s", err)
		}
		bus.CreatedDate = &createdDate
	}

	return bus, nil
}

type UpdateOrderLineItem struct {
	OrderID                       *string `json:"order_id" validate:"omitempty,uuid4"`
	ProductID                     *string `json:"product_id" validate:"omitempty,uuid4"`
	Description                   *string `json:"description" validate:"omitempty"`
	Quantity                      *string `json:"quantity" validate:"omitempty,numeric"`
	UnitPrice                     *string `json:"unit_price" validate:"omitempty"`
	Discount                      *string `json:"discount" validate:"omitempty"`
	DiscountType                  *string `json:"discount_type" validate:"omitempty,oneof=flat percent"`
	LineTotal                     *string `json:"line_total" validate:"omitempty"`
	LineItemFulfillmentStatusesID *string `json:"line_item_fulfillment_statuses_id" validate:"omitempty,uuid4"`
	UpdatedBy                     *string `json:"updated_by" validate:"omitempty,uuid4"`
}

func (app *UpdateOrderLineItem) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app UpdateOrderLineItem) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusUpdateOrderLineItem(app UpdateOrderLineItem) (orderlineitemsbus.UpdateOrderLineItem, error) {
	var orderID *uuid.UUID
	if app.OrderID != nil {
		id, err := uuid.Parse(*app.OrderID)
		if err != nil {
			return orderlineitemsbus.UpdateOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse orderID: %s", err)
		}
		orderID = &id
	}

	var productID *uuid.UUID
	if app.ProductID != nil {
		id, err := uuid.Parse(*app.ProductID)
		if err != nil {
			return orderlineitemsbus.UpdateOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse productID: %s", err)
		}
		productID = &id
	}

	var quantity *int
	if app.Quantity != nil {
		q, err := strconv.Atoi(*app.Quantity)
		if err != nil {
			return orderlineitemsbus.UpdateOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse quantity: %s", err)
		}
		quantity = &q
	}

	var unitPrice *types.Money
	if app.UnitPrice != nil {
		m, err := types.ParseMoneyPtr(*app.UnitPrice)
		if err != nil {
			return orderlineitemsbus.UpdateOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse unit_price: %s", err)
		}
		unitPrice = m
	}

	var discount *types.Money
	if app.Discount != nil {
		m, err := types.ParseMoneyPtr(*app.Discount)
		if err != nil {
			return orderlineitemsbus.UpdateOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse discount: %s", err)
		}
		discount = m
	}

	var discountType *string
	if app.DiscountType != nil {
		if *app.DiscountType != "flat" && *app.DiscountType != "percent" {
			return orderlineitemsbus.UpdateOrderLineItem{}, errs.Newf(errs.InvalidArgument, "discount_type must be 'flat' or 'percent'")
		}
		discountType = app.DiscountType
	}

	var lineTotal *types.Money
	if app.LineTotal != nil {
		m, err := types.ParseMoneyPtr(*app.LineTotal)
		if err != nil {
			return orderlineitemsbus.UpdateOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse line_total: %s", err)
		}
		lineTotal = m
	}

	var lineItemFulfillmentStatusesID *uuid.UUID
	if app.LineItemFulfillmentStatusesID != nil {
		id, err := uuid.Parse(*app.LineItemFulfillmentStatusesID)
		if err != nil {
			return orderlineitemsbus.UpdateOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse lineItemFulfillmentStatusesID: %s", err)
		}
		lineItemFulfillmentStatusesID = &id
	}

	var updatedBy *uuid.UUID
	if app.UpdatedBy != nil {
		id, err := uuid.Parse(*app.UpdatedBy)
		if err != nil {
			return orderlineitemsbus.UpdateOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse updatedBy: %s", err)
		}
		updatedBy = &id
	}

	bus := orderlineitemsbus.UpdateOrderLineItem{
		OrderID:                       orderID,
		ProductID:                     productID,
		Description:                   app.Description,
		Quantity:                      quantity,
		UnitPrice:                     unitPrice,
		Discount:                      discount,
		DiscountType:                  discountType,
		LineTotal:                     lineTotal,
		LineItemFulfillmentStatusesID: lineItemFulfillmentStatusesID,
		UpdatedBy:                     updatedBy,
	}
	return bus, nil
}
