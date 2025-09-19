package physicalattributeapp

import (
	"encoding/json"
	"fmt"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/products/physicalattributebus"
	"github.com/timmaaaz/ichor/business/sdk/convert"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

type QueryParams struct {
	Page                string
	Rows                string
	OrderBy             string
	ID                  string
	ProductID           string
	Length              string
	Width               string
	Height              string
	Weight              string
	WeightUnit          string
	Color               string
	Size                string
	Material            string
	StorageRequirements string
	HazmatClass         string
	ShelfLifeDays       string
	CreatedDate         string
	UpdatedDate         string
}

type PhysicalAttribute struct {
	ID                  string `json:"attribute_id"`
	ProductID           string `json:"product_id"`
	Length              string `json:"length"`
	Width               string `json:"width"`
	Height              string `json:"height"`
	Weight              string `json:"weight"`
	WeightUnit          string `json:"weight_unit"`
	Color               string `json:"color"`
	Size                string `json:"size"`
	Material            string `json:"material"`
	StorageRequirements string `json:"storage_requirements"`
	HazmatClass         string `json:"hazmat_class"`
	ShelfLifeDays       string `json:"shelf_life_days"`
	CreatedDate         string `json:"created_date"`
	UpdatedDate         string `json:"updated_date"`
}

func (app PhysicalAttribute) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppPhysicalAttribute(bus physicalattributebus.PhysicalAttribute) PhysicalAttribute {
	return PhysicalAttribute{
		ID:                  bus.AttributeID.String(),
		ProductID:           bus.ProductID.String(),
		Length:              bus.Length.String(),
		Width:               bus.Width.String(),
		Height:              bus.Height.String(),
		Weight:              bus.Weight.String(),
		WeightUnit:          bus.WeightUnit,
		Color:               bus.Color,
		Size:                bus.Size,
		Material:            bus.Material,
		StorageRequirements: bus.StorageRequirements,
		HazmatClass:         bus.HazmatClass,
		ShelfLifeDays:       fmt.Sprintf("%d", bus.ShelfLifeDays),
		CreatedDate:         bus.CreatedDate.Format(timeutil.FORMAT),
		UpdatedDate:         bus.UpdatedDate.Format(timeutil.FORMAT),
	}
}

func ToAppPhysicalAttributes(bus []physicalattributebus.PhysicalAttribute) []PhysicalAttribute {
	app := make([]PhysicalAttribute, len(bus))
	for i, v := range bus {
		app[i] = ToAppPhysicalAttribute(v)
	}
	return app
}

type NewPhysicalAttribute struct {
	ProductID           string `json:"product_id" validate:"required,min=36,max=36"`
	Length              string `json:"length" validate:"required"`
	Width               string `json:"width" validate:"required"`
	Height              string `json:"height" validate:"required"`
	Weight              string `json:"weight" validate:"required"`
	WeightUnit          string `json:"weight_unit" validate:"required"`
	Color               string `json:"color" validate:"omitempty"`
	Size                string `json:"size" validate:"omitempty"`
	Material            string `json:"material" validate:"omitempty"`
	StorageRequirements string `json:"storage_requirements" validate:"required"`
	HazmatClass         string `json:"hazmat_class" validate:"required"`
	ShelfLifeDays       string `json:"shelf_life_days" validate:"required"`
}

func (app *NewPhysicalAttribute) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewPhysicalAttribute) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusNewPhysicalAttribute(app NewPhysicalAttribute) (physicalattributebus.NewPhysicalAttribute, error) {
	dest := physicalattributebus.NewPhysicalAttribute{}
	var err error

	dest.Weight, err = physicalattributebus.ParseDimension(app.Weight)
	if err != nil {
		return physicalattributebus.NewPhysicalAttribute{}, errs.NewFieldsError("Weight", err)
	}

	dest.Length, err = physicalattributebus.ParseDimension(app.Length)
	if err != nil {
		return physicalattributebus.NewPhysicalAttribute{}, errs.NewFieldsError("Length", err)
	}

	dest.Width, err = physicalattributebus.ParseDimension(app.Width)
	if err != nil {
		return physicalattributebus.NewPhysicalAttribute{}, errs.NewFieldsError("Width", err)
	}

	dest.Height, err = physicalattributebus.ParseDimension(app.Height)
	if err != nil {
		return physicalattributebus.NewPhysicalAttribute{}, errs.NewFieldsError("Height", err)
	}

	err = convert.PopulateTypesFromStrings(app, &dest)
	return dest, err
}

type UpdatePhysicalAttribute struct {
	ProductID           *string `json:"product_id" validate:"omitempty,min=36,max=36"`
	Length              *string `json:"length"`
	Width               *string `json:"width"`
	Height              *string `json:"height"`
	Weight              *string `json:"weight"`
	WeightUnit          *string `json:"weight_unit"`
	Color               *string `json:"color"`
	Size                *string `json:"size"`
	Material            *string `json:"material"`
	StorageRequirements *string `json:"storage_requirements"`
	HazmatClass         *string `json:"hazmat_class"`
	ShelfLifeDays       *string `json:"shelf_life_days"`
}

// Decode implements the decoder interface.
func (app *UpdatePhysicalAttribute) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks the data in the model is considered clean.
func (app UpdatePhysicalAttribute) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusUpdatePhysicalAttribute(app UpdatePhysicalAttribute) (physicalattributebus.UpdatePhysicalAttribute, error) {
	dest := physicalattributebus.UpdatePhysicalAttribute{}

	if app.Weight != nil {
		w, err := physicalattributebus.ParseDimension(*app.Weight)
		if err != nil {
			return physicalattributebus.UpdatePhysicalAttribute{}, errs.NewFieldsError("Weight", err)
		}
		dest.Weight = &w
	}

	if app.Length != nil {
		l, err := physicalattributebus.ParseDimension(*app.Length)
		if err != nil {
			return physicalattributebus.UpdatePhysicalAttribute{}, errs.NewFieldsError("Length", err)
		}

		dest.Length = &l
	}

	if app.Width != nil {
		w, err := physicalattributebus.ParseDimension(*app.Width)
		if err != nil {
			return physicalattributebus.UpdatePhysicalAttribute{}, errs.NewFieldsError("Width", err)
		}
		dest.Width = &w
	}

	if app.Height != nil {
		h, err := physicalattributebus.ParseDimension(*app.Height)
		if err != nil {
			return physicalattributebus.UpdatePhysicalAttribute{}, errs.NewFieldsError("Height", err)
		}
		dest.Height = &h
	}

	err := convert.PopulateTypesFromStrings(app, &dest)

	return dest, err
}
