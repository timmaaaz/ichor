package supplierproductapp

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierproductbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierproductbus/types"
	"github.com/timmaaaz/ichor/business/sdk/convert"
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
	dest := supplierproductbus.NewSupplierProduct{}

	err := convert.PopulateTypesFromStrings(app, &dest)
	if err != nil {
		return supplierproductbus.NewSupplierProduct{}, fmt.Errorf("error populating supplier product: %s", err)
	}

	dest.UnitCost, err = types.ParseMoney(app.UnitCost)
	if err != nil {
		return dest, err
	}

	return dest, err
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
	dest := supplierproductbus.UpdateSupplierProduct{}

	err := convert.PopulateTypesFromStrings(app, &dest)
	if err != nil {
		return supplierproductbus.UpdateSupplierProduct{}, fmt.Errorf("error populating supplier product: %s", err)
	}

	if app.UnitCost != nil {
		m, err := types.ParseMoney(*app.UnitCost)
		if err != nil {
			return dest, err
		}

		dest.UnitCost = &m
	}

	return dest, err
}
