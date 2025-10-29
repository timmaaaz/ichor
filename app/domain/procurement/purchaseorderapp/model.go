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
	StartOrderDate        string
	EndOrderDate          string
	StartExpectedDelivery string
	EndExpectedDelivery   string
}

// PurchaseOrder represents a purchase order response.
type PurchaseOrder struct {
	ID                      string `json:"id"`
	OrderNumber             string `json:"orderNumber"`
	SupplierID              string `json:"supplierId"`
	PurchaseOrderStatusID   string `json:"purchaseOrderStatusId"`
	DeliveryWarehouseID     string `json:"deliveryWarehouseId"`
	DeliveryLocationID      string `json:"deliveryLocationId"`
	DeliveryStreetID        string `json:"deliveryStreetId"`
	OrderDate               string `json:"orderDate"`
	ExpectedDeliveryDate    string `json:"expectedDeliveryDate"`
	ActualDeliveryDate      string `json:"actualDeliveryDate"`
	Subtotal                string `json:"subtotal"`
	TaxAmount               string `json:"taxAmount"`
	ShippingCost            string `json:"shippingCost"`
	TotalAmount             string `json:"totalAmount"`
	Currency                string `json:"currency"`
	RequestedBy             string `json:"requestedBy"`
	ApprovedBy              string `json:"approvedBy"`
	ApprovedDate            string `json:"approvedDate"`
	Notes                   string `json:"notes"`
	SupplierReferenceNumber string `json:"supplierReferenceNumber"`
	CreatedBy               string `json:"createdBy"`
	UpdatedBy               string `json:"updatedBy"`
	CreatedDate             string `json:"createdDate"`
	UpdatedDate             string `json:"updatedDate"`
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

	return PurchaseOrder{
		ID:                      bus.ID.String(),
		OrderNumber:             bus.OrderNumber,
		SupplierID:              bus.SupplierID.String(),
		PurchaseOrderStatusID:   bus.PurchaseOrderStatusID.String(),
		DeliveryWarehouseID:     bus.DeliveryWarehouseID.String(),
		DeliveryLocationID:      bus.DeliveryLocationID.String(),
		DeliveryStreetID:        bus.DeliveryStreetID.String(),
		OrderDate:               bus.OrderDate.Format(timeutil.FORMAT),
		ExpectedDeliveryDate:    bus.ExpectedDeliveryDate.Format(timeutil.FORMAT),
		ActualDeliveryDate:      actualDeliveryDate,
		Subtotal:                fmt.Sprintf("%.2f", bus.Subtotal),
		TaxAmount:               fmt.Sprintf("%.2f", bus.TaxAmount),
		ShippingCost:            fmt.Sprintf("%.2f", bus.ShippingCost),
		TotalAmount:             fmt.Sprintf("%.2f", bus.TotalAmount),
		Currency:                bus.Currency,
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
	OrderNumber             string `json:"orderNumber" validate:"required"`
	SupplierID              string `json:"supplierId" validate:"required,uuid"`
	PurchaseOrderStatusID   string `json:"purchaseOrderStatusId" validate:"required,uuid"`
	DeliveryWarehouseID     string `json:"deliveryWarehouseId" validate:"required,uuid"`
	DeliveryLocationID      string `json:"deliveryLocationId" validate:"required,uuid"`
	DeliveryStreetID        string `json:"deliveryStreetId" validate:"required,uuid"`
	OrderDate               string `json:"orderDate" validate:"required"`
	ExpectedDeliveryDate    string `json:"expectedDeliveryDate" validate:"required"`
	Subtotal                string `json:"subtotal" validate:"required"`
	TaxAmount               string `json:"taxAmount" validate:"required"`
	ShippingCost            string `json:"shippingCost" validate:"required"`
	TotalAmount             string `json:"totalAmount" validate:"required"`
	Currency                string `json:"currency" validate:"required"`
	RequestedBy             string `json:"requestedBy" validate:"required,uuid"`
	Notes                   string `json:"notes"`
	SupplierReferenceNumber string `json:"supplierReferenceNumber"`
	CreatedBy               string `json:"createdBy" validate:"required,uuid"`
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

	deliveryLocationID, err := uuid.Parse(app.DeliveryLocationID)
	if err != nil {
		return purchaseorderbus.NewPurchaseOrder{}, errs.NewFieldsError("deliveryLocationId", err)
	}

	deliveryStreetID, err := uuid.Parse(app.DeliveryStreetID)
	if err != nil {
		return purchaseorderbus.NewPurchaseOrder{}, errs.NewFieldsError("deliveryStreetId", err)
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

	return purchaseorderbus.NewPurchaseOrder{
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
		Currency:                app.Currency,
		RequestedBy:             requestedBy,
		Notes:                   app.Notes,
		SupplierReferenceNumber: app.SupplierReferenceNumber,
		CreatedBy:               createdBy,
	}, nil
}

// UpdatePurchaseOrder contains information needed to update a purchase order.
type UpdatePurchaseOrder struct {
	OrderNumber             *string `json:"orderNumber" validate:"omitempty"`
	SupplierID              *string `json:"supplierId" validate:"omitempty,uuid"`
	PurchaseOrderStatusID   *string `json:"purchaseOrderStatusId" validate:"omitempty,uuid"`
	DeliveryWarehouseID     *string `json:"deliveryWarehouseId" validate:"omitempty,uuid"`
	DeliveryLocationID      *string `json:"deliveryLocationId" validate:"omitempty,uuid"`
	DeliveryStreetID        *string `json:"deliveryStreetId" validate:"omitempty,uuid"`
	OrderDate               *string `json:"orderDate" validate:"omitempty"`
	ExpectedDeliveryDate    *string `json:"expectedDeliveryDate" validate:"omitempty"`
	ActualDeliveryDate      *string `json:"actualDeliveryDate" validate:"omitempty"`
	Subtotal                *string `json:"subtotal" validate:"omitempty"`
	TaxAmount               *string `json:"taxAmount" validate:"omitempty"`
	ShippingCost            *string `json:"shippingCost" validate:"omitempty"`
	TotalAmount             *string `json:"totalAmount" validate:"omitempty"`
	Currency                *string `json:"currency" validate:"omitempty"`
	ApprovedBy              *string `json:"approvedBy" validate:"omitempty,uuid"`
	ApprovedDate            *string `json:"approvedDate" validate:"omitempty"`
	Notes                   *string `json:"notes" validate:"omitempty"`
	SupplierReferenceNumber *string `json:"supplierReferenceNumber" validate:"omitempty"`
	UpdatedBy               *string `json:"updatedBy" validate:"omitempty,uuid"`
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

	if app.Currency != nil {
		dest.Currency = app.Currency
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
	ApprovedBy string `json:"approvedBy" validate:"required,uuid"`
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
