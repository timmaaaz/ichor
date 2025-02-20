package brandapp

import (
	"encoding/json"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/brandbus"
	"github.com/timmaaaz/ichor/foundation/convert"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

type QueryParams struct {
	Page          string
	Rows          string
	OrderBy       string
	ID            string
	Name          string
	ContactInfoID string
	CreatedDate   string
	UpdatedDate   string
}

type Brand struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	ContactInfoID string `json:"contact_info_id"`
	CreatedDate   string `json:"created_date"`
	UpdatedDate   string `json:"updated_date"`
}

func (app Brand) Encode() ([]byte, string, error) {

	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppBrand(bus brandbus.Brand) Brand {
	return Brand{
		ID:            bus.BrandID.String(),
		Name:          bus.Name,
		ContactInfoID: bus.ContactInfoID.String(),
		CreatedDate:   bus.CreatedDate.Format(timeutil.FORMAT),
		UpdatedDate:   bus.UpdatedDate.Format(timeutil.FORMAT),
	}
}

func ToAppBrands(bus []brandbus.Brand) []Brand {
	app := make([]Brand, len(bus))
	for i, v := range bus {
		app[i] = ToAppBrand(v)
	}
	return app
}

// =========================================================================

type NewBrand struct {
	Name          string `json:"name" validate:"required"`
	ContactInfoID string `json:"contact_info_id" validate:"required,min=36,max=36"`
}

func (app *NewBrand) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewBrand) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusNewBrand(app NewBrand) (brandbus.NewBrand, error) {
	dest := brandbus.NewBrand{}

	err := convert.PopulateTypesFromStrings(app, &dest)
	return dest, err
}

type UpdateBrand struct {
	Name          *string `json:"name" validate:"omitempty"`
	ContactInfoID *string `json:"contact_info_id" validate:"omitempty,min=36,max=36"`
}

// Decode implements the decoder interface.
func (app *UpdateBrand) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks the data in the model is considered clean.
func (app UpdateBrand) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusUpdateBrand(app UpdateBrand) (brandbus.UpdateBrand, error) {
	dest := brandbus.UpdateBrand{}

	err := convert.PopulateTypesFromStrings(app, &dest)

	return dest, err
}
