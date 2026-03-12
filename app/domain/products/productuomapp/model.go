package productuomapp

import (
	"encoding/json"
	"strconv"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/products/productuombus"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

// QueryParams holds the raw query string values for filtering.
type QueryParams struct {
	Page      string
	Rows      string
	OrderBy   string
	ID        string
	ProductID string
	IsBase    string
	Name      string
}

// ProductUOM is the app-layer representation of a product UOM.
type ProductUOM struct {
	ID               string `json:"id"`
	ProductID        string `json:"product_id"`
	Name             string `json:"name"`
	Abbreviation     string `json:"abbreviation"`
	ConversionFactor string `json:"conversion_factor"`
	IsBase           bool   `json:"is_base"`
	IsApproximate    bool   `json:"is_approximate"`
	Notes            string `json:"notes"`
	CreatedDate      string `json:"created_date"`
	UpdatedDate      string `json:"updated_date"`
}

// Encode implements web.Encoder.
func (app ProductUOM) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// ToAppProductUOM converts a bus ProductUOM to an app ProductUOM.
func ToAppProductUOM(bus productuombus.ProductUOM) ProductUOM {
	return ProductUOM{
		ID:               bus.ID.String(),
		ProductID:        bus.ProductID.String(),
		Name:             bus.Name,
		Abbreviation:     bus.Abbreviation,
		ConversionFactor: strconv.FormatFloat(bus.ConversionFactor, 'f', -1, 64),
		IsBase:           bus.IsBase,
		IsApproximate:    bus.IsApproximate,
		Notes:            bus.Notes,
		CreatedDate:      bus.CreatedDate.Format(timeutil.FORMAT),
		UpdatedDate:      bus.UpdatedDate.Format(timeutil.FORMAT),
	}
}

// ToAppProductUOMs converts a slice of bus ProductUOMs.
func ToAppProductUOMs(uoms []productuombus.ProductUOM) []ProductUOM {
	app := make([]ProductUOM, len(uoms))
	for i, u := range uoms {
		app[i] = ToAppProductUOM(u)
	}
	return app
}

// =============================================================================
// Create

// NewProductUOM is the app-layer create request.
type NewProductUOM struct {
	ProductID        string  `json:"product_id" validate:"required,min=36,max=36"`
	Name             string  `json:"name" validate:"required"`
	Abbreviation     string  `json:"abbreviation"`
	ConversionFactor float64 `json:"conversion_factor" validate:"required"`
	IsBase           bool    `json:"is_base"`
	IsApproximate    bool    `json:"is_approximate"`
	Notes            string  `json:"notes"`
}

// Decode implements web.Decoder.
func (app *NewProductUOM) Decode(data []byte) error {
	return json.Unmarshal(data, app)
}

// Validate checks the data in the model is considered clean.
func (app NewProductUOM) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusNewProductUOM(app NewProductUOM) (productuombus.NewProductUOM, error) {
	productID, err := uuid.Parse(app.ProductID)
	if err != nil {
		return productuombus.NewProductUOM{}, errs.Newf(errs.InvalidArgument, "invalid product_id: %s", err)
	}

	return productuombus.NewProductUOM{
		ProductID:        productID,
		Name:             app.Name,
		Abbreviation:     app.Abbreviation,
		ConversionFactor: app.ConversionFactor,
		IsBase:           app.IsBase,
		IsApproximate:    app.IsApproximate,
		Notes:            app.Notes,
	}, nil
}

// =============================================================================
// Update

// UpdateProductUOM is the app-layer update request.
type UpdateProductUOM struct {
	Name             *string  `json:"name"`
	Abbreviation     *string  `json:"abbreviation"`
	ConversionFactor *float64 `json:"conversion_factor"`
	IsBase           *bool    `json:"is_base"`
	IsApproximate    *bool    `json:"is_approximate"`
	Notes            *string  `json:"notes"`
}

// Decode implements web.Decoder.
func (app *UpdateProductUOM) Decode(data []byte) error {
	return json.Unmarshal(data, app)
}

// Validate checks the data in the model is considered clean.
func (app UpdateProductUOM) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusUpdateProductUOM(app UpdateProductUOM) productuombus.UpdateProductUOM {
	return productuombus.UpdateProductUOM{
		Name:             app.Name,
		Abbreviation:     app.Abbreviation,
		ConversionFactor: app.ConversionFactor,
		IsBase:           app.IsBase,
		IsApproximate:    app.IsApproximate,
		Notes:            app.Notes,
	}
}

// =============================================================================
// Query helpers

func parseFilter(qp QueryParams) (productuombus.QueryFilter, error) {
	filter := productuombus.QueryFilter{}

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return productuombus.QueryFilter{}, errs.Newf(errs.InvalidArgument, "invalid id: %s", err)
		}
		filter.ID = &id
	}

	if qp.ProductID != "" {
		pid, err := uuid.Parse(qp.ProductID)
		if err != nil {
			return productuombus.QueryFilter{}, errs.Newf(errs.InvalidArgument, "invalid product_id: %s", err)
		}
		filter.ProductID = &pid
	}

	if qp.Name != "" {
		filter.Name = &qp.Name
	}

	if qp.IsBase != "" {
		b, err := strconv.ParseBool(qp.IsBase)
		if err != nil {
			return productuombus.QueryFilter{}, errs.Newf(errs.InvalidArgument, "invalid is_base: %s", err)
		}
		filter.IsBase = &b
	}

	return filter, nil
}
