package lineitemfulfillmentstatusapp

import (
	"encoding/json"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/sales/lineitemfulfillmentstatusbus"
)

type QueryParams struct {
	Page    string
	Rows    string
	OrderBy string

	ID          string
	Name        string
	Description string
}

type LineItemFulfillmentStatus struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	PrimaryColor   string `json:"primary_color"`
	SecondaryColor string `json:"secondary_color"`
	Icon           string `json:"icon"`
}

func (app LineItemFulfillmentStatus) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppLineItemFulfillmentStatus(bus lineitemfulfillmentstatusbus.LineItemFulfillmentStatus) LineItemFulfillmentStatus {
	return LineItemFulfillmentStatus{
		ID:             bus.ID.String(),
		Name:           bus.Name,
		Description:    bus.Description,
		PrimaryColor:   bus.PrimaryColor,
		SecondaryColor: bus.SecondaryColor,
		Icon:           bus.Icon,
	}
}

func ToAppLineItemFulfillmentStatuses(bus []lineitemfulfillmentstatusbus.LineItemFulfillmentStatus) []LineItemFulfillmentStatus {
	appStatuses := make([]LineItemFulfillmentStatus, len(bus))
	for i, status := range bus {
		appStatuses[i] = ToAppLineItemFulfillmentStatus(status)
	}
	return appStatuses
}

type NewLineItemFulfillmentStatus struct {
	Name           string `json:"name" validate:"required,min=3"`
	Description    string `json:"description" validate:"omitempty"`
	PrimaryColor   string `json:"primary_color" validate:"omitempty,max=50"`
	SecondaryColor string `json:"secondary_color" validate:"omitempty,max=50"`
	Icon           string `json:"icon" validate:"omitempty,max=100"`
}

func (app *NewLineItemFulfillmentStatus) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewLineItemFulfillmentStatus) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusNewLineItemFulfillmentStatus(app NewLineItemFulfillmentStatus) (lineitemfulfillmentstatusbus.NewLineItemFulfillmentStatus, error) {
	bus := lineitemfulfillmentstatusbus.NewLineItemFulfillmentStatus{
		Name:           app.Name,
		Description:    app.Description,
		PrimaryColor:   app.PrimaryColor,
		SecondaryColor: app.SecondaryColor,
		Icon:           app.Icon,
	}

	return bus, nil
}

type UpdateLineItemFulfillmentStatus struct {
	Name           *string `json:"name" validate:"omitempty,min=3"`
	Description    *string `json:"description" validate:"omitempty"`
	PrimaryColor   *string `json:"primary_color" validate:"omitempty,max=50"`
	SecondaryColor *string `json:"secondary_color" validate:"omitempty,max=50"`
	Icon           *string `json:"icon" validate:"omitempty,max=100"`
}

func (app *UpdateLineItemFulfillmentStatus) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app UpdateLineItemFulfillmentStatus) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusUpdateLineItemFulfillmentStatus(app UpdateLineItemFulfillmentStatus) (lineitemfulfillmentstatusbus.UpdateLineItemFulfillmentStatus, error) {
	bus := lineitemfulfillmentstatusbus.UpdateLineItemFulfillmentStatus{
		Name:           app.Name,
		Description:    app.Description,
		PrimaryColor:   app.PrimaryColor,
		SecondaryColor: app.SecondaryColor,
		Icon:           app.Icon,
	}

	return bus, nil
}
