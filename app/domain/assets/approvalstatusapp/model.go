package approvalstatusapp

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/assets/approvalstatusbus"
)

type QueryParams struct {
	Page    string
	Rows    string
	OrderBy string
	ID      string
	IconID  string
	Name    string
}

type ApprovalStatus struct {
	ID             string `json:"id"`
	IconID         string `json:"icon_id"`
	Name           string `json:"name"`
	PrimaryColor   string `json:"primary_color"`
	SecondaryColor string `json:"secondary_color"`
	Icon           string `json:"icon"`
}

func (app ApprovalStatus) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppApprovalStatus(bus approvalstatusbus.ApprovalStatus) ApprovalStatus {
	return ApprovalStatus{
		ID:             bus.ID.String(),
		IconID:         bus.IconID.String(),
		Name:           bus.Name,
		PrimaryColor:   bus.PrimaryColor,
		SecondaryColor: bus.SecondaryColor,
		Icon:           bus.Icon,
	}
}

func ToAppApprovalStatuses(bus []approvalstatusbus.ApprovalStatus) []ApprovalStatus {
	app := make([]ApprovalStatus, len(bus))
	for i, v := range bus {
		app[i] = ToAppApprovalStatus(v)
	}
	return app
}

// =============================================================================

type NewApprovalStatus struct {
	IconID         string `json:"icon_id" validate:"required"`
	Name           string `json:"name" validate:"required,min=3,max=100"`
	PrimaryColor   string `json:"primary_color" validate:"omitempty,max=50"`
	SecondaryColor string `json:"secondary_color" validate:"omitempty,max=50"`
	Icon           string `json:"icon" validate:"omitempty,max=100"`
}

func (app *NewApprovalStatus) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewApprovalStatus) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "Validate: %s", err)
	}

	return nil
}

func toBusNewApprovalStatus(app NewApprovalStatus) (approvalstatusbus.NewApprovalStatus, error) {
	var iconID uuid.UUID
	var err error

	if iconID, err = uuid.Parse(app.IconID); err != nil {
		return approvalstatusbus.NewApprovalStatus{}, err
	}

	return approvalstatusbus.NewApprovalStatus{
		IconID:         iconID,
		Name:           app.Name, // TODO: Look at defining custom type
		PrimaryColor:   app.PrimaryColor,
		SecondaryColor: app.SecondaryColor,
		Icon:           app.Icon,
	}, nil
}

type UpdateApprovalStatus struct {
	IconID         *string `json:"icon_id" validate:"omitempty"`
	Name           *string `json:"name" validate:"omitempty,min=3,max=100"`
	PrimaryColor   *string `json:"primary_color" validate:"omitempty,max=50"`
	SecondaryColor *string `json:"secondary_color" validate:"omitempty,max=50"`
	Icon           *string `json:"icon" validate:"omitempty,max=100"`
}

func (app *UpdateApprovalStatus) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app UpdateApprovalStatus) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusUpdateApprovalStatus(app UpdateApprovalStatus) (approvalstatusbus.UpdateApprovalStatus, error) {
	var iconID *uuid.UUID
	var name *string

	if app.IconID != nil {
		if id, err := uuid.Parse(*app.IconID); err != nil {
			return approvalstatusbus.UpdateApprovalStatus{}, err
		} else {
			iconID = &id
		}
	}

	if app.Name != nil {
		name = app.Name
	}

	return approvalstatusbus.UpdateApprovalStatus{
		IconID:         iconID,
		Name:           name,
		PrimaryColor:   app.PrimaryColor,
		SecondaryColor: app.SecondaryColor,
		Icon:           app.Icon,
	}, nil
}
