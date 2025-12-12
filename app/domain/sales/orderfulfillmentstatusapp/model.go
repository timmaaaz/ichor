package orderfulfillmentstatusapp

import (
	"encoding/json"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/sales/orderfulfillmentstatusbus"
)

type QueryParams struct {
	Page    string
	Rows    string
	OrderBy string

	ID          string
	Name        string
	Description string
}

type OrderFulfillmentStatus struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	PrimaryColor   string `json:"primary_color"`
	SecondaryColor string `json:"secondary_color"`
	Icon           string `json:"icon"`
}

func (app OrderFulfillmentStatus) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppOrderFulfillmentStatus(bus orderfulfillmentstatusbus.OrderFulfillmentStatus) OrderFulfillmentStatus {
	return OrderFulfillmentStatus{
		ID:             bus.ID.String(),
		Name:           bus.Name,
		Description:    bus.Description,
		PrimaryColor:   bus.PrimaryColor,
		SecondaryColor: bus.SecondaryColor,
		Icon:           bus.Icon,
	}
}

func ToAppOrderFulfillmentStatuses(bus []orderfulfillmentstatusbus.OrderFulfillmentStatus) []OrderFulfillmentStatus {
	appStatuses := make([]OrderFulfillmentStatus, len(bus))
	for i, status := range bus {
		appStatuses[i] = ToAppOrderFulfillmentStatus(status)
	}
	return appStatuses
}

type NewOrderFulfillmentStatus struct {
	Name           string `json:"name" validate:"required,min=3"`
	Description    string `json:"description" validate:"omitempty"`
	PrimaryColor   string `json:"primary_color" validate:"omitempty,max=50"`
	SecondaryColor string `json:"secondary_color" validate:"omitempty,max=50"`
	Icon           string `json:"icon" validate:"omitempty,max=100"`
}

func (app *NewOrderFulfillmentStatus) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewOrderFulfillmentStatus) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusNewOrderFulfillmentStatus(app NewOrderFulfillmentStatus) (orderfulfillmentstatusbus.NewOrderFulfillmentStatus, error) {
	bus := orderfulfillmentstatusbus.NewOrderFulfillmentStatus{
		Name:           app.Name,
		Description:    app.Description,
		PrimaryColor:   app.PrimaryColor,
		SecondaryColor: app.SecondaryColor,
		Icon:           app.Icon,
	}

	return bus, nil
}

type UpdateOrderFulfillmentStatus struct {
	Name           *string `json:"name" validate:"omitempty,min=3"`
	Description    *string `json:"description" validate:"omitempty"`
	PrimaryColor   *string `json:"primary_color" validate:"omitempty,max=50"`
	SecondaryColor *string `json:"secondary_color" validate:"omitempty,max=50"`
	Icon           *string `json:"icon" validate:"omitempty,max=100"`
}

func (app *UpdateOrderFulfillmentStatus) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app UpdateOrderFulfillmentStatus) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusUpdateOrderFulfillmentStatus(app UpdateOrderFulfillmentStatus) (orderfulfillmentstatusbus.UpdateOrderFulfillmentStatus, error) {
	bus := orderfulfillmentstatusbus.UpdateOrderFulfillmentStatus{
		Name:           app.Name,
		Description:    app.Description,
		PrimaryColor:   app.PrimaryColor,
		SecondaryColor: app.SecondaryColor,
		Icon:           app.Icon,
	}

	return bus, nil
}
