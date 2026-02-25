package purchaseorderapp

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderbus"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

// QueryParams holds the query parameters for filtering purchase orders.
type QueryParams struct {
	Page                  string
	Rows                  string
	OrderBy               string
	ID                    string
	OrderNumber           string
	SupplierID            string
	PurchaseOrderStatusID string
	DeliveryWarehouseID   string
	RequestedBy           string
	ApprovedBy            string
	StartOrderDate          string
	EndOrderDate            string
	StartExpectedDelivery   string
	EndExpectedDelivery     string
	StartActualDeliveryDate string
	EndActualDeliveryDate   string
	IsUndelivered           string
}

// PurchaseOrder represents a purchase order response.
type PurchaseOrder struct {
	ID                      string `json:"id"`
	OrderNumber             string `json:"order_number"`
	SupplierID              string `json:"supplier_id"`
	PurchaseOrderStatusID   string `json:"purchase_order_status_id"`
	DeliveryWarehouseID     string `json:"delivery_warehouse_id"`
	DeliveryLocationID      string `json:"delivery_location_id"`
	DeliveryStreetID        string `json:"delivery_street_id"`
	OrderDate               string `json:"order_date"`
	ExpectedDeliveryDate    string `json:"expected_delivery_date"`
	ActualDeliveryDate      string `json:"actual_delivery_date"`
	Subtotal                string `json:"subtotal"`
	TaxAmount               string `json:"tax_amount"`
	ShippingCost            string `json:"shipping_cost"`
	TotalAmount             string `json:"total_amount"`
	CurrencyID              string `json:"currency_id"`
	RequestedBy             string `json:"requested_by"`
	ApprovedBy              string `json:"approved_by"`
	ApprovedDate            string `json:"approved_date"`
	Notes                   string `json:"notes"`
	SupplierReferenceNumber string `json:"supplier_reference_number"`
	CreatedBy               string `json:"created_by"`
	UpdatedBy               string `json:"updated_by"`
	CreatedDate             string `json:"created_date"`
	UpdatedDate             string `json:"updated_date"`
}

// Encode implements the Encoder interface.
func (app PurchaseOrder) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// ToAppPurchaseOrder converts a business purchase order to an app purchase order.
func ToAppPurchaseOrder(bus purchaseorderbus.PurchaseOrder) PurchaseOrder {
	actualDeliveryDate := ""
	if !bus.ActualDeliveryDate.IsZero() {
		actualDeliveryDate = bus.ActualDeliveryDate.Format(timeutil.FORMAT)
	}

	approvedDate := ""
	if !bus.ApprovedDate.IsZero() {
		approvedDate = bus.ApprovedDate.Format(timeutil.FORMAT)
	}

	approvedBy := ""
	if bus.ApprovedBy != uuid.Nil {
		approvedBy = bus.ApprovedBy.String()
	}

	deliveryLocationID := ""
	if bus.DeliveryLocationID != uuid.Nil {
		deliveryLocationID = bus.DeliveryLocationID.String()
	}

	deliveryStreetID := ""
	if bus.DeliveryStreetID != uuid.Nil {
		deliveryStreetID = bus.DeliveryStreetID.String()
	}

	return PurchaseOrder{
		ID:                      bus.ID.String(),
		OrderNumber:             bus.OrderNumber,
		SupplierID:              bus.SupplierID.String(),
		PurchaseOrderStatusID:   bus.PurchaseOrderStatusID.String(),
		DeliveryWarehouseID:     bus.DeliveryWarehouseID.String(),
		DeliveryLocationID:      deliveryLocationID,
		DeliveryStreetID:        deliveryStreetID,
		OrderDate:               bus.OrderDate.Format(timeutil.FORMAT),
		ExpectedDeliveryDate:    bus.ExpectedDeliveryDate.Format(timeutil.FORMAT),
		ActualDeliveryDate:      actualDeliveryDate,
		Subtotal:                fmt.Sprintf("%.2f", bus.Subtotal),
		TaxAmount:               fmt.Sprintf("%.2f", bus.TaxAmount),
		ShippingCost:            fmt.Sprintf("%.2f", bus.ShippingCost),
		TotalAmount:             fmt.Sprintf("%.2f", bus.TotalAmount),
		CurrencyID:              bus.CurrencyID.String(),
		RequestedBy:             bus.RequestedBy.String(),
		ApprovedBy:              approvedBy,
		ApprovedDate:            approvedDate,
		Notes:                   bus.Notes,
		SupplierReferenceNumber: bus.SupplierReferenceNumber,
		CreatedBy:               bus.CreatedBy.String(),
		UpdatedBy:               bus.UpdatedBy.String(),
		CreatedDate:             bus.CreatedDate.Format(timeutil.FORMAT),
		UpdatedDate:             bus.UpdatedDate.Format(timeutil.FORMAT),
	}
}

// ToAppPurchaseOrders converts a slice of business purchase orders to app purchase orders.
func ToAppPurchaseOrders(bus []purchaseorderbus.PurchaseOrder) []PurchaseOrder {
	app := make([]PurchaseOrder, len(bus))
	for i, v := range bus {
		app[i] = ToAppPurchaseOrder(v)
	}
	return app
}

// NewPurchaseOrder contains information needed to create a new purchase order.
type NewPurchaseOrder struct {
	OrderNumber             string  `json:"order_number" validate:"required"`
	SupplierID              string  `json:"supplier_id" validate:"required,uuid"`
	PurchaseOrderStatusID   string  `json:"purchase_order_status_id" validate:"required,uuid"`
	DeliveryWarehouseID     string  `json:"delivery_warehouse_id" validate:"required,uuid"`
	DeliveryLocationID      string  `json:"delivery_location_id" validate:"omitempty,uuid"`
	DeliveryStreetID        string  `json:"delivery_street_id" validate:"omitempty,uuid"`
	OrderDate               string  `json:"order_date" validate:"required"`
	ExpectedDeliveryDate    string  `json:"expected_delivery_date" validate:"required"`
	Subtotal                string  `json:"subtotal" validate:"required"`
	TaxAmount               string  `json:"tax_amount" validate:"required"`
	ShippingCost            string  `json:"shipping_cost" validate:"required"`
	TotalAmount             string  `json:"total_amount" validate:"required"`
	CurrencyID              string  `json:"currency_id" validate:"required,uuid"`
	RequestedBy             string  `json:"requested_by" validate:"required,uuid"`
	Notes                   string  `json:"notes"`
	SupplierReferenceNumber string  `json:"supplier_reference_number"`
	CreatedBy               string  `json:"created_by" validate:"required,uuid"`
	CreatedDate             *string `json:"created_date"` // Optional: for seeding/import
}

// Decode implements the Decoder interface.
func (app *NewPurchaseOrder) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate validates the NewPurchaseOrder fields.
func (app NewPurchaseOrder) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

// toBusNewPurchaseOrder converts an app NewPurchaseOrder to a business NewPurchaseOrder.
func toBusNewPurchaseOrder(app NewPurchaseOrder) (purchaseorderbus.NewPurchaseOrder, error) {
	supplierID, err := uuid.Parse(app.SupplierID)
	if err != nil {
		return purchaseorderbus.NewPurchaseOrder{}, errs.NewFieldsError("supplierId", err)
	}

	purchaseOrderStatusID, err := uuid.Parse(app.PurchaseOrderStatusID)
	if err != nil {
		return purchaseorderbus.NewPurchaseOrder{}, errs.NewFieldsError("purchaseOrderStatusId", err)
	}

	deliveryWarehouseID, err := uuid.Parse(app.DeliveryWarehouseID)
	if err != nil {
		return purchaseorderbus.NewPurchaseOrder{}, errs.NewFieldsError("deliveryWarehouseId", err)
	}

	var deliveryLocationID uuid.UUID
	if app.DeliveryLocationID != "" {
		deliveryLocationID, err = uuid.Parse(app.DeliveryLocationID)
		if err != nil {
			return purchaseorderbus.NewPurchaseOrder{}, errs.NewFieldsError("deliveryLocationId", err)
		}
	}

	var deliveryStreetID uuid.UUID
	if app.DeliveryStreetID != "" {
		deliveryStreetID, err = uuid.Parse(app.DeliveryStreetID)
		if err != nil {
			return purchaseorderbus.NewPurchaseOrder{}, errs.NewFieldsError("deliveryStreetId", err)
		}
	}

	orderDate, err := time.Parse(timeutil.FORMAT, app.OrderDate)
	if err != nil {
		return purchaseorderbus.NewPurchaseOrder{}, errs.NewFieldsError("orderDate", err)
	}

	expectedDeliveryDate, err := time.Parse(timeutil.FORMAT, app.ExpectedDeliveryDate)
	if err != nil {
		return purchaseorderbus.NewPurchaseOrder{}, errs.NewFieldsError("expectedDeliveryDate", err)
	}

	var subtotal float64
	if _, err := fmt.Sscanf(app.Subtotal, "%f", &subtotal); err != nil {
		return purchaseorderbus.NewPurchaseOrder{}, errs.NewFieldsError("subtotal", err)
	}

	var taxAmount float64
	if _, err := fmt.Sscanf(app.TaxAmount, "%f", &taxAmount); err != nil {
		return purchaseorderbus.NewPurchaseOrder{}, errs.NewFieldsError("taxAmount", err)
	}

	var shippingCost float64
	if _, err := fmt.Sscanf(app.ShippingCost, "%f", &shippingCost); err != nil {
		return purchaseorderbus.NewPurchaseOrder{}, errs.NewFieldsError("shippingCost", err)
	}

	var totalAmount float64
	if _, err := fmt.Sscanf(app.TotalAmount, "%f", &totalAmount); err != nil {
		return purchaseorderbus.NewPurchaseOrder{}, errs.NewFieldsError("totalAmount", err)
	}

	requestedBy, err := uuid.Parse(app.RequestedBy)
	if err != nil {
		return purchaseorderbus.NewPurchaseOrder{}, errs.NewFieldsError("requestedBy", err)
	}

	createdBy, err := uuid.Parse(app.CreatedBy)
	if err != nil {
		return purchaseorderbus.NewPurchaseOrder{}, errs.NewFieldsError("createdBy", err)
	}

	currencyID, err := uuid.Parse(app.CurrencyID)
	if err != nil {
		return purchaseorderbus.NewPurchaseOrder{}, errs.NewFieldsError("currencyId", err)
	}

	bus := purchaseorderbus.NewPurchaseOrder{
		OrderNumber:             app.OrderNumber,
		SupplierID:              supplierID,
		PurchaseOrderStatusID:   purchaseOrderStatusID,
		DeliveryWarehouseID:     deliveryWarehouseID,
		DeliveryLocationID:      deliveryLocationID,
		DeliveryStreetID:        deliveryStreetID,
		OrderDate:               orderDate,
		ExpectedDeliveryDate:    expectedDeliveryDate,
		Subtotal:                subtotal,
		TaxAmount:               taxAmount,
		ShippingCost:            shippingCost,
		TotalAmount:             totalAmount,
		CurrencyID:              currencyID,
		RequestedBy:             requestedBy,
		Notes:                   app.Notes,
		SupplierReferenceNumber: app.SupplierReferenceNumber,
		CreatedBy:               createdBy,
		// CreatedDate: nil by default - API always uses server time
	}

	// Handle optional CreatedDate (for imports/admin tools only)
	if app.CreatedDate != nil && *app.CreatedDate != "" {
		createdDate, err := time.Parse(time.RFC3339, *app.CreatedDate)
		if err != nil {
			return purchaseorderbus.NewPurchaseOrder{}, errs.NewFieldsError("createdDate", err)
		}
		bus.CreatedDate = &createdDate
	}

	return bus, nil
}

// UpdatePurchaseOrder contains information needed to update a purchase order.
type UpdatePurchaseOrder struct {
	OrderNumber             *string `json:"order_number" validate:"omitempty"`
	SupplierID              *string `json:"supplier_id" validate:"omitempty,uuid"`
	PurchaseOrderStatusID   *string `json:"purchase_order_status_id" validate:"omitempty,uuid"`
	DeliveryWarehouseID     *string `json:"delivery_warehouse_id" validate:"omitempty,uuid"`
	DeliveryLocationID      *string `json:"delivery_location_id" validate:"omitempty,uuid"`
	DeliveryStreetID        *string `json:"delivery_street_id" validate:"omitempty,uuid"`
	OrderDate               *string `json:"order_date" validate:"omitempty"`
	ExpectedDeliveryDate    *string `json:"expected_delivery_date" validate:"omitempty"`
	ActualDeliveryDate      *string `json:"actual_delivery_date" validate:"omitempty"`
	Subtotal                *string `json:"subtotal" validate:"omitempty"`
	TaxAmount               *string `json:"tax_amount" validate:"omitempty"`
	ShippingCost            *string `json:"shipping_cost" validate:"omitempty"`
	TotalAmount             *string `json:"total_amount" validate:"omitempty"`
	CurrencyID              *string `json:"currency_id" validate:"omitempty,uuid"`
	ApprovedBy              *string `json:"approved_by" validate:"omitempty,uuid"`
	ApprovedDate            *string `json:"approved_date" validate:"omitempty"`
	Notes                   *string `json:"notes" validate:"omitempty"`
	SupplierReferenceNumber *string `json:"supplier_reference_number" validate:"omitempty"`
	UpdatedBy               *string `json:"updated_by" validate:"omitempty,uuid"`
}

// Decode implements the Decoder interface.
func (app *UpdatePurchaseOrder) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate validates the UpdatePurchaseOrder fields.
func (app UpdatePurchaseOrder) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

// toBusUpdatePurchaseOrder converts an app UpdatePurchaseOrder to a business UpdatePurchaseOrder.
func toBusUpdatePurchaseOrder(app UpdatePurchaseOrder) (purchaseorderbus.UpdatePurchaseOrder, error) {
	dest := purchaseorderbus.UpdatePurchaseOrder{}

	if app.OrderNumber != nil {
		dest.OrderNumber = app.OrderNumber
	}

	if app.SupplierID != nil {
		supplierID, err := uuid.Parse(*app.SupplierID)
		if err != nil {
			return purchaseorderbus.UpdatePurchaseOrder{}, errs.NewFieldsError("supplierId", err)
		}
		dest.SupplierID = &supplierID
	}

	if app.PurchaseOrderStatusID != nil {
		purchaseOrderStatusID, err := uuid.Parse(*app.PurchaseOrderStatusID)
		if err != nil {
			return purchaseorderbus.UpdatePurchaseOrder{}, errs.NewFieldsError("purchaseOrderStatusId", err)
		}
		dest.PurchaseOrderStatusID = &purchaseOrderStatusID
	}

	if app.DeliveryWarehouseID != nil {
		deliveryWarehouseID, err := uuid.Parse(*app.DeliveryWarehouseID)
		if err != nil {
			return purchaseorderbus.UpdatePurchaseOrder{}, errs.NewFieldsError("deliveryWarehouseId", err)
		}
		dest.DeliveryWarehouseID = &deliveryWarehouseID
	}

	if app.DeliveryLocationID != nil {
		deliveryLocationID, err := uuid.Parse(*app.DeliveryLocationID)
		if err != nil {
			return purchaseorderbus.UpdatePurchaseOrder{}, errs.NewFieldsError("deliveryLocationId", err)
		}
		dest.DeliveryLocationID = &deliveryLocationID
	}

	if app.DeliveryStreetID != nil {
		deliveryStreetID, err := uuid.Parse(*app.DeliveryStreetID)
		if err != nil {
			return purchaseorderbus.UpdatePurchaseOrder{}, errs.NewFieldsError("deliveryStreetId", err)
		}
		dest.DeliveryStreetID = &deliveryStreetID
	}

	if app.OrderDate != nil {
		orderDate, err := time.Parse(timeutil.FORMAT, *app.OrderDate)
		if err != nil {
			return purchaseorderbus.UpdatePurchaseOrder{}, errs.NewFieldsError("orderDate", err)
		}
		dest.OrderDate = &orderDate
	}

	if app.ExpectedDeliveryDate != nil {
		expectedDeliveryDate, err := time.Parse(timeutil.FORMAT, *app.ExpectedDeliveryDate)
		if err != nil {
			return purchaseorderbus.UpdatePurchaseOrder{}, errs.NewFieldsError("expectedDeliveryDate", err)
		}
		dest.ExpectedDeliveryDate = &expectedDeliveryDate
	}

	if app.ActualDeliveryDate != nil {
		actualDeliveryDate, err := time.Parse(timeutil.FORMAT, *app.ActualDeliveryDate)
		if err != nil {
			return purchaseorderbus.UpdatePurchaseOrder{}, errs.NewFieldsError("actualDeliveryDate", err)
		}
		dest.ActualDeliveryDate = &actualDeliveryDate
	}

	if app.Subtotal != nil {
		var subtotal float64
		if _, err := fmt.Sscanf(*app.Subtotal, "%f", &subtotal); err != nil {
			return purchaseorderbus.UpdatePurchaseOrder{}, errs.NewFieldsError("subtotal", err)
		}
		dest.Subtotal = &subtotal
	}

	if app.TaxAmount != nil {
		var taxAmount float64
		if _, err := fmt.Sscanf(*app.TaxAmount, "%f", &taxAmount); err != nil {
			return purchaseorderbus.UpdatePurchaseOrder{}, errs.NewFieldsError("taxAmount", err)
		}
		dest.TaxAmount = &taxAmount
	}

	if app.ShippingCost != nil {
		var shippingCost float64
		if _, err := fmt.Sscanf(*app.ShippingCost, "%f", &shippingCost); err != nil {
			return purchaseorderbus.UpdatePurchaseOrder{}, errs.NewFieldsError("shippingCost", err)
		}
		dest.ShippingCost = &shippingCost
	}

	if app.TotalAmount != nil {
		var totalAmount float64
		if _, err := fmt.Sscanf(*app.TotalAmount, "%f", &totalAmount); err != nil {
			return purchaseorderbus.UpdatePurchaseOrder{}, errs.NewFieldsError("totalAmount", err)
		}
		dest.TotalAmount = &totalAmount
	}

	if app.CurrencyID != nil {
		currencyID, err := uuid.Parse(*app.CurrencyID)
		if err != nil {
			return purchaseorderbus.UpdatePurchaseOrder{}, errs.NewFieldsError("currencyId", err)
		}
		dest.CurrencyID = &currencyID
	}

	if app.ApprovedBy != nil {
		approvedBy, err := uuid.Parse(*app.ApprovedBy)
		if err != nil {
			return purchaseorderbus.UpdatePurchaseOrder{}, errs.NewFieldsError("approvedBy", err)
		}
		dest.ApprovedBy = &approvedBy
	}

	if app.ApprovedDate != nil {
		approvedDate, err := time.Parse(timeutil.FORMAT, *app.ApprovedDate)
		if err != nil {
			return purchaseorderbus.UpdatePurchaseOrder{}, errs.NewFieldsError("approvedDate", err)
		}
		dest.ApprovedDate = &approvedDate
	}

	if app.Notes != nil {
		dest.Notes = app.Notes
	}

	if app.SupplierReferenceNumber != nil {
		dest.SupplierReferenceNumber = app.SupplierReferenceNumber
	}

	if app.UpdatedBy != nil {
		updatedBy, err := uuid.Parse(*app.UpdatedBy)
		if err != nil {
			return purchaseorderbus.UpdatePurchaseOrder{}, errs.NewFieldsError("updatedBy", err)
		}
		dest.UpdatedBy = &updatedBy
	}

	return dest, nil
}

// PurchaseOrders is a collection wrapper that implements the Encoder interface.
type PurchaseOrders []PurchaseOrder

// Encode implements the Encoder interface.
func (app PurchaseOrders) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// QueryByIDsRequest represents a request to query multiple purchase orders by their IDs.
type QueryByIDsRequest struct {
	IDs []string `json:"ids" validate:"required,min=1"`
}

// Decode implements the Decoder interface.
func (app *QueryByIDsRequest) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate validates the QueryByIDsRequest fields.
func (app QueryByIDsRequest) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

// ApproveRequest represents a request to approve a purchase order.
type ApproveRequest struct {
	ApprovedBy string `json:"approved_by" validate:"required,uuid"`
}

// Decode implements the Decoder interface.
func (app *ApproveRequest) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate validates the ApproveRequest fields.
func (app ApproveRequest) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

// toBusIDs converts a slice of string IDs to a slice of UUIDs.
func toBusIDs(ids []string) ([]uuid.UUID, error) {
	uuids := make([]uuid.UUID, len(ids))
	for i, id := range ids {
		uid, err := uuid.Parse(id)
		if err != nil {
			return nil, fmt.Errorf("parse id[%d]: %w", i, err)
		}
		uuids[i] = uid
	}
	return uuids, nil
}
