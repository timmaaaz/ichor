package lineitemfulfillmentstatusapp

import (
	"encoding/json"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/order/lineitemfulfillmentstatusbus"
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

type LineItemFulfillmentStatus struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (app LineItemFulfillmentStatus) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppLineItemFulfillmentStatus(bus lineitemfulfillmentstatusbus.LineItemFulfillmentStatus) LineItemFulfillmentStatus {
	return LineItemFulfillmentStatus{
		ID:          bus.ID.String(),
		Name:        bus.Name,
		Description: bus.Description,
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
	Name        string `json:"name" validate:"required,min=3"`
	Description string `json:"description" validate:"omitempty"`
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
	var dest lineitemfulfillmentstatusbus.NewLineItemFulfillmentStatus
	err := convert.PopulateTypesFromStrings(app, &dest)
	if err != nil {
		return lineitemfulfillmentstatusbus.NewLineItemFulfillmentStatus{}, errs.Newf(errs.InvalidArgument, "toBusNewLineItemFulfillmentStatus: %s", err)
	}

	return dest, nil
}

type UpdateLineItemFulfillmentStatus struct {
	Name        *string `json:"name" validate:"omitempty,min=3"`
	Description *string `json:"description" validate:"omitempty"`
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
	var dest lineitemfulfillmentstatusbus.UpdateLineItemFulfillmentStatus
	err := convert.PopulateTypesFromStrings(app, &dest)
	if err != nil {
		return lineitemfulfillmentstatusbus.UpdateLineItemFulfillmentStatus{}, errs.Newf(errs.InvalidArgument, "toBusUpdateLineItemFulfillmentStatus: %s", err)
	}

	return dest, nil
}
