package productapp

import (
	"encoding/json"
	"fmt"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/products/productbus"
	"github.com/timmaaaz/ichor/business/sdk/convert"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

type QueryParams struct {
	Page                 string
	Rows                 string
	OrderBy              string
	ProductID            string
	SKU                  string
	BrandID              string
	ProductCategoryID    string
	Name                 string
	Description          string
	ModelNumber          string
	UpcCode              string
	Status               string
	IsActive             string
	IsPerishable         string
	HandlingInstructions string
	UnitsPerCase         string
	CreatedDate          string
	UpdatedDate          string
}

type Product struct {
	ProductID            string `json:"id"`
	SKU                  string `json:"sku"`
	BrandID              string `json:"brand_id"`
	ProductCategoryID    string `json:"product_category_id"`
	Name                 string `json:"name"`
	Description          string `json:"description"`
	ModelNumber          string `json:"model_number"`
	UpcCode              string `json:"upc_code"`
	Status               string `json:"status"`
	IsActive             string `json:"is_active"`
	IsPerishable         string `json:"is_perishable"`
	HandlingInstructions string `json:"handling_instructions"`
	UnitsPerCase         string `json:"units_per_case"`
	CreatedDate          string `json:"created_date"`
	UpdatedDate          string `json:"updated_date"`
}

func (app Product) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppProduct(bus productbus.Product) Product {
	return Product{
		ProductID:            bus.ProductID.String(),
		SKU:                  bus.SKU,
		BrandID:              bus.BrandID.String(),
		ProductCategoryID:    bus.ProductCategoryID.String(),
		Name:                 bus.Name,
		Description:          bus.Description,
		ModelNumber:          bus.ModelNumber,
		UpcCode:              bus.UpcCode,
		Status:               bus.Status,
		IsActive:             fmt.Sprintf("%v", bus.IsActive),
		IsPerishable:         fmt.Sprintf("%v", bus.IsPerishable),
		HandlingInstructions: bus.HandlingInstructions,
		UnitsPerCase:         fmt.Sprintf("%d", bus.UnitsPerCase),
		CreatedDate:          bus.CreatedDate.Format(timeutil.FORMAT),
		UpdatedDate:          bus.UpdatedDate.Format(timeutil.FORMAT),
	}
}

func ToAppProducts(bus []productbus.Product) []Product {
	app := make([]Product, len(bus))
	for i, v := range bus {
		app[i] = ToAppProduct(v)
	}
	return app
}

type NewProduct struct {
	SKU                  string `json:"sku" validate:"required"`
	BrandID              string `json:"brand_id" validate:"required,min=36,max=36"`
	ProductCategoryID    string `json:"product_category_id" validate:"required,min=36,max=36"`
	Name                 string `json:"name" validate:"required"`
	Description          string `json:"description" validate:"required"`
	ModelNumber          string `json:"model_number" validate:"required"`
	UpcCode              string `json:"upc_code" validate:"required"`
	Status               string `json:"status" validate:"required"`
	IsActive             string `json:"is_active" validate:"required"`
	IsPerishable         string `json:"is_perishable" validate:"required"`
	HandlingInstructions string `json:"handling_instructions"`
	UnitsPerCase         string `json:"units_per_case" validate:"required"`
}

func (app *NewProduct) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewProduct) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusNewProduct(app NewProduct) (productbus.NewProduct, error) {
	dest := productbus.NewProduct{}
	err := convert.PopulateTypesFromStrings(app, &dest)
	return dest, err
}

type UpdateProduct struct {
	SKU                  *string `json:"sku" validate:"omitempty"`
	BrandID              *string `json:"brand_id" validate:"omitempty,min=36,max=36"`
	ProductCategoryID    *string `json:"product_category_id" validate:"omitempty,min=36,max=36"`
	Name                 *string `json:"name" validate:"omitempty"`
	Description          *string `json:"description" validate:"omitempty"`
	ModelNumber          *string `json:"model_number" validate:"omitempty"`
	UpcCode              *string `json:"upc_code" validate:"omitempty"`
	Status               *string `json:"status" validate:"omitempty"`
	IsActive             *string `json:"is_active" validate:"omitempty"`
	IsPerishable         *string `json:"is_perishable" validate:"omitempty"`
	HandlingInstructions *string `json:"handling_instructions"`
	UnitsPerCase         *string `json:"units_per_case" validate:"omitempty"`
}

// Decode implements the decoder interface.
func (app *UpdateProduct) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks the data in the model is considered clean.
func (app UpdateProduct) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusUpdateProduct(app UpdateProduct) (productbus.UpdateProduct, error) {
	dest := productbus.UpdateProduct{}

	err := convert.PopulateTypesFromStrings(app, &dest)

	return dest, err
}
