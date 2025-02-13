package fulfillmentstatusapp

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/assets/fulfillmentstatusbus"
)

type QueryParams struct {
	Page    string
	Rows    string
	OrderBy string
	ID      string
	IconID  string
	Name    string
}

type FulfillmentStatus struct {
	ID     string `json:"id"`
	IconID string `json:"icon_id"`
	Name   string `json:"name"`
}

func (app FulfillmentStatus) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppFulfillmentStatus(bus fulfillmentstatusbus.FulfillmentStatus) FulfillmentStatus {
	return FulfillmentStatus{
		ID:     bus.ID.String(),
		IconID: bus.IconID.String(),
		Name:   bus.Name,
	}
}

func ToAppFulfillmentStatuses(bus []fulfillmentstatusbus.FulfillmentStatus) []FulfillmentStatus {
	app := make([]FulfillmentStatus, len(bus))
	for i, v := range bus {
		app[i] = ToAppFulfillmentStatus(v)
	}
	return app
}

// =============================================================================

type NewFulfillmentStatus struct {
	IconId string `json:"icon_id" validate:"required"`
	Name   string `json:"name" validate:"required,min=3,max=100"`
}

func (app *NewFulfillmentStatus) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewFulfillmentStatus) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusNewFulfillmentStatus(app NewFulfillmentStatus) (fulfillmentstatusbus.NewFulfillmentStatus, error) {
	var iconID uuid.UUID
	var err error

	if iconID, err = uuid.Parse(app.IconId); err != nil {
		return fulfillmentstatusbus.NewFulfillmentStatus{}, err
	}

	return fulfillmentstatusbus.NewFulfillmentStatus{
		IconID: iconID,
		Name:   app.Name,
	}, nil
}

type UpdateFulfillmentStatus struct {
	IconID *string `json:"icon_id" validate:"required"`
	Name   *string `json:"name" validate:"required,min=3,max=100"`
}

func (app *UpdateFulfillmentStatus) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app UpdateFulfillmentStatus) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusUpdateFulfillmentStatus(app UpdateFulfillmentStatus) (fulfillmentstatusbus.UpdateFulfillmentStatus, error) {
	var iconID *uuid.UUID
	var name *string

	if app.IconID != nil {
		if id, err := uuid.Parse(*app.IconID); err != nil {
			return fulfillmentstatusbus.UpdateFulfillmentStatus{}, err
		} else {
			iconID = &id
		}
	}

	if app.Name != nil {
		name = app.Name
	}

	return fulfillmentstatusbus.UpdateFulfillmentStatus{
		IconID: iconID,
		Name:   name,
	}, nil
}
