package inspectionapp

import (
	"encoding/json"
	"fmt"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/quality/inspectionbus"
	"github.com/timmaaaz/ichor/business/sdk/convert"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

type QueryParams struct {
	Page    string
	Rows    string
	OrderBy string

	InspectionID       string
	ProductID          string
	InspectorID        string
	LotID              string
	Status             string
	Notes              string
	InspectionDate     string
	NextInspectionDate string
	UpdatedDate        string
	CreatedDate        string
}

type Inspection struct {
	InspectionID       string `json:"inspection_id"`
	ProductID          string `json:"product_id"`
	InspectorID        string `json:"inspector_id"`
	LotID              string `json:"lot_id"`
	Status             string `json:"status"`
	Notes              string `json:"notes"`
	InspectionDate     string `json:"inspection_date"`
	NextInspectionDate string `json:"next_inspection_date"`
	UpdatedDate        string `json:"updated_date"`
	CreatedDate        string `json:"created_date"`
}

func (app Inspection) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppInspection(bus inspectionbus.Inspection) Inspection {
	return Inspection{
		InspectionID:       bus.InspectionID.String(),
		ProductID:          bus.ProductID.String(),
		InspectorID:        bus.InspectorID.String(),
		LotID:              bus.LotID.String(),
		Status:             bus.Status,
		Notes:              bus.Notes,
		InspectionDate:     bus.InspectionDate.Format(timeutil.FORMAT),
		NextInspectionDate: bus.NextInspectionDate.Format(timeutil.FORMAT),
		UpdatedDate:        bus.UpdatedDate.Format(timeutil.FORMAT),
		CreatedDate:        bus.CreatedDate.Format(timeutil.FORMAT),
	}
}

func ToAppInspections(bus []inspectionbus.Inspection) []Inspection {
	app := make([]Inspection, len(bus))
	for i, v := range bus {
		app[i] = ToAppInspection(v)
	}
	return app
}

type NewInspection struct {
	ProductID          string `json:"product_id" validate:"required,min=36,max=36"`
	InspectorID        string `json:"inspector_id" validate:"required,min=36,max=36"`
	LotID              string `json:"lot_id" validate:"required,min=36,max=36"`
	Status             string `json:"status" validate:"required"`
	Notes              string `json:"notes" validate:"required"`
	InspectionDate     string `json:"inspection_date" validate:"required"`
	NextInspectionDate string `json:"next_inspection_date" validate:"required"`
}

func (app *NewInspection) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewInspection) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusNewInspection(app NewInspection) (inspectionbus.NewInspection, error) {
	dest := inspectionbus.NewInspection{}

	err := convert.PopulateTypesFromStrings(app, &dest)
	if err != nil {
		return inspectionbus.NewInspection{}, fmt.Errorf("toBusNewInspection: %w", err)
	}

	return dest, nil
}

type UpdateInspection struct {
	ProductID          *string `json:"product_id" validate:"omitempty,min=36,max=36"`
	InspectorID        *string `json:"inspector_id" validate:"omitempty,min=36,max=36"`
	LotID              *string `json:"lot_id" validate:"omitempty,min=36,max=36"`
	Status             *string `json:"status" validate:"omitempty"`
	Notes              *string `json:"notes" validate:"omitempty"`
	InspectionDate     *string `json:"inspection_date" validate:"omitempty"`
	NextInspectionDate *string `json:"next_inspection_date" validate:"omitempty"`
}

func (app *UpdateInspection) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app UpdateInspection) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusUpdateInspection(app UpdateInspection) (inspectionbus.UpdateInspection, error) {
	dest := inspectionbus.UpdateInspection{}

	err := convert.PopulateTypesFromStrings(app, &dest)
	if err != nil {
		return inspectionbus.UpdateInspection{}, fmt.Errorf("toBusUpdateInspection: %w", err)
	}

	return dest, nil
}
