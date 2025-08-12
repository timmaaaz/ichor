package orderfulfillmentstatusapp

import (
	"encoding/json"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/order/orderfulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/sdk/convert"
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
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (app OrderFulfillmentStatus) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppOrderFulfillmentStatus(bus orderfulfillmentstatusbus.OrderFulfillmentStatus) OrderFulfillmentStatus {
	return OrderFulfillmentStatus{
		ID:          bus.ID.String(),
		Name:        bus.Name,
		Description: bus.Description,
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
	Name        string `json:"name" validate:"required,min=3"`
	Description string `json:"description" validate:"omitempty"`
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
	var dest orderfulfillmentstatusbus.NewOrderFulfillmentStatus
	err := convert.PopulateTypesFromStrings(app, &dest)
	if err != nil {
		return orderfulfillmentstatusbus.NewOrderFulfillmentStatus{}, errs.Newf(errs.InvalidArgument, "toBusNewOrderFulfillmentStatus: %s", err)
	}

	return dest, nil
}

type UpdateOrderFulfillmentStatus struct {
	Name        *string `json:"name" validate:"omitempty,min=3"`
	Description *string `json:"description" validate:"omitempty"`
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
	var dest orderfulfillmentstatusbus.UpdateOrderFulfillmentStatus
	err := convert.PopulateTypesFromStrings(app, &dest)
	if err != nil {
		return orderfulfillmentstatusbus.UpdateOrderFulfillmentStatus{}, errs.Newf(errs.InvalidArgument, "toBusUpdateOrderFulfillmentStatus: %s", err)
	}

	return dest, nil
}
