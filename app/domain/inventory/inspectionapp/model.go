package inspectionapp

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/inventory/inspectionbus"
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
	productID, err := uuid.Parse(app.ProductID)
	if err != nil {
		return inspectionbus.NewInspection{}, errs.Newf(errs.InvalidArgument, "parse productID: %s", err)
	}

	inspectorID, err := uuid.Parse(app.InspectorID)
	if err != nil {
		return inspectionbus.NewInspection{}, errs.Newf(errs.InvalidArgument, "parse inspectorID: %s", err)
	}

	lotID, err := uuid.Parse(app.LotID)
	if err != nil {
		return inspectionbus.NewInspection{}, errs.Newf(errs.InvalidArgument, "parse lotID: %s", err)
	}

	inspectionDate, err := time.Parse(timeutil.FORMAT, app.InspectionDate)
	if err != nil {
		return inspectionbus.NewInspection{}, errs.Newf(errs.InvalidArgument, "parse inspectionDate: %s", err)
	}

	nextInspectionDate, err := time.Parse(timeutil.FORMAT, app.NextInspectionDate)
	if err != nil {
		return inspectionbus.NewInspection{}, errs.Newf(errs.InvalidArgument, "parse nextInspectionDate: %s", err)
	}

	bus := inspectionbus.NewInspection{
		ProductID:          productID,
		InspectorID:        inspectorID,
		LotID:              lotID,
		Status:             app.Status,
		Notes:              app.Notes,
		InspectionDate:     inspectionDate,
		NextInspectionDate: nextInspectionDate,
	}
	return bus, nil
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
	var productID *uuid.UUID
	if app.ProductID != nil {
		id, err := uuid.Parse(*app.ProductID)
		if err != nil {
			return inspectionbus.UpdateInspection{}, errs.Newf(errs.InvalidArgument, "parse productID: %s", err)
		}
		productID = &id
	}

	var inspectorID *uuid.UUID
	if app.InspectorID != nil {
		id, err := uuid.Parse(*app.InspectorID)
		if err != nil {
			return inspectionbus.UpdateInspection{}, errs.Newf(errs.InvalidArgument, "parse inspectorID: %s", err)
		}
		inspectorID = &id
	}

	var lotID *uuid.UUID
	if app.LotID != nil {
		id, err := uuid.Parse(*app.LotID)
		if err != nil {
			return inspectionbus.UpdateInspection{}, errs.Newf(errs.InvalidArgument, "parse lotID: %s", err)
		}
		lotID = &id
	}

	var inspectionDate *time.Time
	if app.InspectionDate != nil {
		t, err := time.Parse(timeutil.FORMAT, *app.InspectionDate)
		if err != nil {
			return inspectionbus.UpdateInspection{}, errs.Newf(errs.InvalidArgument, "parse inspectionDate: %s", err)
		}
		inspectionDate = &t
	}

	var nextInspectionDate *time.Time
	if app.NextInspectionDate != nil {
		t, err := time.Parse(timeutil.FORMAT, *app.NextInspectionDate)
		if err != nil {
			return inspectionbus.UpdateInspection{}, errs.Newf(errs.InvalidArgument, "parse nextInspectionDate: %s", err)
		}
		nextInspectionDate = &t
	}

	bus := inspectionbus.UpdateInspection{
		ProductID:          productID,
		InspectorID:        inspectorID,
		LotID:              lotID,
		Status:             app.Status,
		Notes:              app.Notes,
		InspectionDate:     inspectionDate,
		NextInspectionDate: nextInspectionDate,
	}
	return bus, nil
}
