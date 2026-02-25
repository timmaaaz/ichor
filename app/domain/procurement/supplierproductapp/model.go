package supplierproductapp

import (
	"encoding/json"
	"strconv"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierproductbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierproductbus/types"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

type QueryParams struct {
	Page    string
	Rows    string
	OrderBy string

	SupplierProductID  string
	SupplierID         string
	ProductID          string
	SupplierPartNumber string
	MinOrderQuantity   string
	MaxOrderQuantity   string
	LeadTimeDays       string
	UnitCost           string
	IsPrimarySupplier  string
	CreatedDate        string
	UpdatedDate        string
}

type SupplierProduct struct {
	SupplierProductID  string `json:"supplier_product_id"`
	SupplierID         string `json:"supplier_id"`
	ProductID          string `json:"product_id"`
	SupplierPartNumber string `json:"supplier_part_number"`
	MinOrderQuantity   string `json:"min_order_quantity"`
	MaxOrderQuantity   string `json:"max_order_quantity"`
	LeadTimeDays       string `json:"lead_time_days"`
	UnitCost           string `json:"unit_cost"`
	IsPrimarySupplier  string `json:"is_primary_supplier"`
	CreatedDate        string `json:"created_date"`
	UpdatedDate        string `json:"updated_date"`
}

func (app SupplierProduct) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppSupplierProduct(bus supplierproductbus.SupplierProduct) SupplierProduct {
	return SupplierProduct{
		SupplierProductID:  bus.SupplierProductID.String(),
		SupplierID:         bus.SupplierID.String(),
		ProductID:          bus.ProductID.String(),
		SupplierPartNumber: bus.SupplierPartNumber,
		MinOrderQuantity:   strconv.Itoa(bus.MinOrderQuantity),
		MaxOrderQuantity:   strconv.Itoa(bus.MaxOrderQuantity),
		LeadTimeDays:       strconv.Itoa(bus.LeadTimeDays),
		UnitCost:           bus.UnitCost.Value(),
		IsPrimarySupplier:  strconv.FormatBool(bus.IsPrimarySupplier),
		CreatedDate:        bus.CreatedDate.Format(timeutil.FORMAT),
		UpdatedDate:        bus.UpdatedDate.Format(timeutil.FORMAT),
	}
}

func ToAppSupplierProducts(bus []supplierproductbus.SupplierProduct) []SupplierProduct {
	app := make([]SupplierProduct, len(bus))
	for i, v := range bus {
		app[i] = ToAppSupplierProduct(v)
	}
	return app
}

// =========================================================================

type NewSupplierProduct struct {
	SupplierID         string `json:"supplier_id" validate:"required,min=36,max=36"`
	ProductID          string `json:"product_id" validate:"required,min=36,max=36"`
	SupplierPartNumber string `json:"supplier_part_number" validate:"required"`
	MinOrderQuantity   string `json:"min_order_quantity" validate:"required"`
	MaxOrderQuantity   string `json:"max_order_quantity" validate:"required"`
	LeadTimeDays       string `json:"lead_time_days" validate:"required"`
	UnitCost           string `json:"unit_cost" validate:"required"`
	IsPrimarySupplier  string `json:"is_primary_supplier" validate:"required"`
}

func (app *NewSupplierProduct) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewSupplierProduct) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusNewSupplierProduct(app NewSupplierProduct) (supplierproductbus.NewSupplierProduct, error) {
	supplierID, err := uuid.Parse(app.SupplierID)
	if err != nil {
		return supplierproductbus.NewSupplierProduct{}, errs.Newf(errs.InvalidArgument, "parse supplier_id: %s", err)
	}

	productID, err := uuid.Parse(app.ProductID)
	if err != nil {
		return supplierproductbus.NewSupplierProduct{}, errs.Newf(errs.InvalidArgument, "parse product_id: %s", err)
	}

	minOrderQuantity, err := strconv.Atoi(app.MinOrderQuantity)
	if err != nil {
		return supplierproductbus.NewSupplierProduct{}, errs.Newf(errs.InvalidArgument, "parse min_order_quantity: %s", err)
	}

	maxOrderQuantity, err := strconv.Atoi(app.MaxOrderQuantity)
	if err != nil {
		return supplierproductbus.NewSupplierProduct{}, errs.Newf(errs.InvalidArgument, "parse max_order_quantity: %s", err)
	}

	leadTimeDays, err := strconv.Atoi(app.LeadTimeDays)
	if err != nil {
		return supplierproductbus.NewSupplierProduct{}, errs.Newf(errs.InvalidArgument, "parse lead_time_days: %s", err)
	}

	unitCost, err := types.ParseMoney(app.UnitCost)
	if err != nil {
		return supplierproductbus.NewSupplierProduct{}, errs.Newf(errs.InvalidArgument, "parse unit_cost: %s", err)
	}

	isPrimarySupplier, err := strconv.ParseBool(app.IsPrimarySupplier)
	if err != nil {
		return supplierproductbus.NewSupplierProduct{}, errs.Newf(errs.InvalidArgument, "parse is_primary_supplier: %s", err)
	}

	bus := supplierproductbus.NewSupplierProduct{
		SupplierID:         supplierID,
		ProductID:          productID,
		SupplierPartNumber: app.SupplierPartNumber,
		MinOrderQuantity:   minOrderQuantity,
		MaxOrderQuantity:   maxOrderQuantity,
		LeadTimeDays:       leadTimeDays,
		UnitCost:           unitCost,
		IsPrimarySupplier:  isPrimarySupplier,
	}

	return bus, nil
}

// =========================================================================

type UpdateSupplierProduct struct {
	SupplierID         *string `json:"supplier_id" validate:"omitempty,min=36,max=36"`
	ProductID          *string `json:"product_id" validate:"omitempty,min=36,max=36"`
	SupplierPartNumber *string `json:"supplier_part_number"`
	MinOrderQuantity   *string `json:"min_order_quantity"`
	MaxOrderQuantity   *string `json:"max_order_quantity"`
	LeadTimeDays       *string `json:"lead_time_days"`
	UnitCost           *string `json:"unit_cost"`
	IsPrimarySupplier  *string `json:"is_primary_supplier"`
}

func (app *UpdateSupplierProduct) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app UpdateSupplierProduct) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusUpdateSupplierProduct(app UpdateSupplierProduct) (supplierproductbus.UpdateSupplierProduct, error) {
	bus := supplierproductbus.UpdateSupplierProduct{}

	if app.SupplierID != nil {
		id, err := uuid.Parse(*app.SupplierID)
		if err != nil {
			return supplierproductbus.UpdateSupplierProduct{}, errs.Newf(errs.InvalidArgument, "parse supplier_id: %s", err)
		}
		bus.SupplierID = &id
	}

	if app.ProductID != nil {
		id, err := uuid.Parse(*app.ProductID)
		if err != nil {
			return supplierproductbus.UpdateSupplierProduct{}, errs.Newf(errs.InvalidArgument, "parse product_id: %s", err)
		}
		bus.ProductID = &id
	}

	if app.MinOrderQuantity != nil {
		qty, err := strconv.Atoi(*app.MinOrderQuantity)
		if err != nil {
			return supplierproductbus.UpdateSupplierProduct{}, errs.Newf(errs.InvalidArgument, "parse min_order_quantity: %s", err)
		}
		bus.MinOrderQuantity = &qty
	}

	if app.MaxOrderQuantity != nil {
		qty, err := strconv.Atoi(*app.MaxOrderQuantity)
		if err != nil {
			return supplierproductbus.UpdateSupplierProduct{}, errs.Newf(errs.InvalidArgument, "parse max_order_quantity: %s", err)
		}
		bus.MaxOrderQuantity = &qty
	}

	if app.LeadTimeDays != nil {
		days, err := strconv.Atoi(*app.LeadTimeDays)
		if err != nil {
			return supplierproductbus.UpdateSupplierProduct{}, errs.Newf(errs.InvalidArgument, "parse lead_time_days: %s", err)
		}
		bus.LeadTimeDays = &days
	}

	if app.UnitCost != nil {
		cost, err := types.ParseMoney(*app.UnitCost)
		if err != nil {
			return supplierproductbus.UpdateSupplierProduct{}, errs.Newf(errs.InvalidArgument, "parse unit_cost: %s", err)
		}
		bus.UnitCost = &cost
	}

	if app.IsPrimarySupplier != nil {
		isPrimary, err := strconv.ParseBool(*app.IsPrimarySupplier)
		if err != nil {
			return supplierproductbus.UpdateSupplierProduct{}, errs.Newf(errs.InvalidArgument, "parse is_primary_supplier: %s", err)
		}
		bus.IsPrimarySupplier = &isPrimary
	}

	bus.SupplierPartNumber = app.SupplierPartNumber

	return bus, nil
}

// SupplierProducts is a collection wrapper that implements the Encoder interface.
type SupplierProducts []SupplierProduct

// Encode implements the Encoder interface.
func (app SupplierProducts) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// QueryByIDsRequest represents a request to query multiple supplier products by their IDs.
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


