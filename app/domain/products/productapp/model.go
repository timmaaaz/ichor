package productapp

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/products/productbus"
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
	TrackingType         string
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
	TrackingType         string `json:"tracking_type"`
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
		TrackingType:         bus.TrackingType,
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
	SKU                  string  `json:"sku" validate:"required"`
	BrandID              string  `json:"brand_id" validate:"required,min=36,max=36"`
	ProductCategoryID    string  `json:"product_category_id" validate:"required,min=36,max=36"`
	Name                 string  `json:"name" validate:"required"`
	Description          string  `json:"description" validate:"required"`
	ModelNumber          string  `json:"model_number" validate:"omitempty"`
	UpcCode              string  `json:"upc_code" validate:"required"`
	Status               string  `json:"status" validate:"required"`
	IsActive             string  `json:"is_active" validate:"required"`
	IsPerishable         string  `json:"is_perishable" validate:"required"`
	HandlingInstructions string  `json:"handling_instructions"`
	UnitsPerCase         string  `json:"units_per_case" validate:"required"`
	TrackingType         string  `json:"tracking_type" validate:"omitempty,oneof=none lot serial"`
	CreatedDate          *string `json:"created_date"` // Optional: for seeding/import
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
	brandID, err := uuid.Parse(app.BrandID)
	if err != nil {
		return productbus.NewProduct{}, errs.Newf(errs.InvalidArgument, "parse brandID: %s", err)
	}

	productCategoryID, err := uuid.Parse(app.ProductCategoryID)
	if err != nil {
		return productbus.NewProduct{}, errs.Newf(errs.InvalidArgument, "parse productCategoryID: %s", err)
	}

	isActive, err := strconv.ParseBool(app.IsActive)
	if err != nil {
		return productbus.NewProduct{}, errs.Newf(errs.InvalidArgument, "parse isActive: %s", err)
	}

	isPerishable, err := strconv.ParseBool(app.IsPerishable)
	if err != nil {
		return productbus.NewProduct{}, errs.Newf(errs.InvalidArgument, "parse isPerishable: %s", err)
	}

	unitsPerCase, err := strconv.Atoi(app.UnitsPerCase)
	if err != nil {
		return productbus.NewProduct{}, errs.Newf(errs.InvalidArgument, "parse unitsPerCase: %s", err)
	}

	trackingType := app.TrackingType
	if trackingType == "" {
		trackingType = "none"
	}

	bus := productbus.NewProduct{
		SKU:                  app.SKU,
		BrandID:              brandID,
		ProductCategoryID:    productCategoryID,
		Name:                 app.Name,
		Description:          app.Description,
		ModelNumber:          app.ModelNumber,
		UpcCode:              app.UpcCode,
		Status:               app.Status,
		IsActive:             isActive,
		IsPerishable:         isPerishable,
		HandlingInstructions: app.HandlingInstructions,
		UnitsPerCase:         unitsPerCase,
		TrackingType:         trackingType,
		// CreatedDate: nil by default - API always uses server time
	}

	// Handle optional CreatedDate (for imports/admin tools only)
	if app.CreatedDate != nil && *app.CreatedDate != "" {
		createdDate, err := time.Parse(time.RFC3339, *app.CreatedDate)
		if err != nil {
			return productbus.NewProduct{}, errs.Newf(errs.InvalidArgument, "parse createdDate: %s", err)
		}
		bus.CreatedDate = &createdDate
	}

	return bus, nil
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
	TrackingType         *string `json:"tracking_type" validate:"omitempty,oneof=none lot serial"`
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
	var brandID *uuid.UUID
	if app.BrandID != nil {
		id, err := uuid.Parse(*app.BrandID)
		if err != nil {
			return productbus.UpdateProduct{}, errs.Newf(errs.InvalidArgument, "parse brandID: %s", err)
		}
		brandID = &id
	}

	var productCategoryID *uuid.UUID
	if app.ProductCategoryID != nil {
		id, err := uuid.Parse(*app.ProductCategoryID)
		if err != nil {
			return productbus.UpdateProduct{}, errs.Newf(errs.InvalidArgument, "parse productCategoryID: %s", err)
		}
		productCategoryID = &id
	}

	var isActive *bool
	if app.IsActive != nil {
		b, err := strconv.ParseBool(*app.IsActive)
		if err != nil {
			return productbus.UpdateProduct{}, errs.Newf(errs.InvalidArgument, "parse isActive: %s", err)
		}
		isActive = &b
	}

	var isPerishable *bool
	if app.IsPerishable != nil {
		b, err := strconv.ParseBool(*app.IsPerishable)
		if err != nil {
			return productbus.UpdateProduct{}, errs.Newf(errs.InvalidArgument, "parse isPerishable: %s", err)
		}
		isPerishable = &b
	}

	var unitsPerCase *int
	if app.UnitsPerCase != nil {
		i, err := strconv.Atoi(*app.UnitsPerCase)
		if err != nil {
			return productbus.UpdateProduct{}, errs.Newf(errs.InvalidArgument, "parse unitsPerCase: %s", err)
		}
		unitsPerCase = &i
	}

	bus := productbus.UpdateProduct{
		SKU:                  app.SKU,
		BrandID:              brandID,
		ProductCategoryID:    productCategoryID,
		Name:                 app.Name,
		Description:          app.Description,
		ModelNumber:          app.ModelNumber,
		UpcCode:              app.UpcCode,
		Status:               app.Status,
		IsActive:             isActive,
		IsPerishable:         isPerishable,
		HandlingInstructions: app.HandlingInstructions,
		UnitsPerCase:         unitsPerCase,
		TrackingType:         app.TrackingType,
	}
	return bus, nil
}

// Products is a collection wrapper that implements the Encoder interface.
type Products []Product

// Encode implements the Encoder interface.
func (app Products) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// QueryByIDsRequest represents a request to query multiple products by their IDs.
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
