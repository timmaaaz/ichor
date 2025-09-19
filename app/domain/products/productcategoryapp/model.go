package productcategoryapp

import (
	"encoding/json"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/products/productcategorybus"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

type QueryParams struct {
	Page        string
	Rows        string
	OrderBy     string
	ID          string
	Name        string
	Description string
	CreatedDate string
	UpdatedDate string
}

type ProductCategory struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedDate string `json:"created_date"`
	UpdatedDate string `json:"updated_date"`
}

func (app ProductCategory) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppProductCategory(bus productcategorybus.ProductCategory) ProductCategory {
	return ProductCategory{
		ID:          bus.ProductCategoryID.String(),
		Name:        bus.Name,
		Description: bus.Description,
		CreatedDate: bus.CreatedDate.Format(timeutil.FORMAT),
		UpdatedDate: bus.UpdatedDate.Format(timeutil.FORMAT),
	}
}

func ToAppProductCategories(bus []productcategorybus.ProductCategory) []ProductCategory {
	app := make([]ProductCategory, len(bus))
	for i, v := range bus {
		app[i] = ToAppProductCategory(v)
	}
	return app
}

type NewProductCategory struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description" validate:"required"`
}

func (app *NewProductCategory) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewProductCategory) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusNewProductCategory(app NewProductCategory) productcategorybus.NewProductCategory {
	return productcategorybus.NewProductCategory{
		Name:        app.Name,
		Description: app.Description,
	}
}

type UpdateProductCategory struct {
	Name        *string `json:"name" validate:"omitempty"`
	Description *string `json:"description" validate:"omitempty"`
}

func (app *UpdateProductCategory) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app UpdateProductCategory) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusUpdateProductCategory(app UpdateProductCategory) productcategorybus.UpdateProductCategory {
	return productcategorybus.UpdateProductCategory{
		Name:        app.Name,
		Description: app.Description,
	}
}
