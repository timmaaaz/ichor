package approvalapp

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/hr/approvalbus"
)

type QueryParams struct {
	Page    string
	Rows    string
	OrderBy string
	ID      string
	IconID  string
	Name    string
}

type UserApprovalStatus struct {
	ID             string `json:"id"`
	IconID         string `json:"icon_id"`
	Name           string `json:"name"`
	PrimaryColor   string `json:"primary_color"`
	SecondaryColor string `json:"secondary_color"`
	Icon           string `json:"icon"`
}

func (app UserApprovalStatus) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppUserApprovalStatus(bus approvalbus.UserApprovalStatus) UserApprovalStatus {
	return UserApprovalStatus{
		ID:             bus.ID.String(),
		IconID:         bus.IconID.String(),
		Name:           bus.Name,
		PrimaryColor:   bus.PrimaryColor,
		SecondaryColor: bus.SecondaryColor,
		Icon:           bus.Icon,
	}
}

func ToAppUserApprovalStatuses(bus []approvalbus.UserApprovalStatus) []UserApprovalStatus {
	app := make([]UserApprovalStatus, len(bus))
	for i, v := range bus {
		app[i] = ToAppUserApprovalStatus(v)
	}
	return app
}

// =============================================================================

type NewUserApprovalStatus struct {
	IconID         string `json:"icon_id" validate:"omitempty,uuid"`
	Name           string `json:"name" validate:"required,min=3,max=100"`
	PrimaryColor   string `json:"primary_color" validate:"omitempty,max=50"`
	SecondaryColor string `json:"secondary_color" validate:"omitempty,max=50"`
	Icon           string `json:"icon" validate:"omitempty,max=100"`
}

func (app *NewUserApprovalStatus) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewUserApprovalStatus) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "Validate: %s", err)
	}

	return nil
}

func toBusNewUserApprovalStatus(app NewUserApprovalStatus) (approvalbus.NewUserApprovalStatus, error) {
	var iconID uuid.UUID

	if app.IconID != "" {
		var err error
		if iconID, err = uuid.Parse(app.IconID); err != nil {
			return approvalbus.NewUserApprovalStatus{}, err
		}
	}
	// If app.IconID is empty, iconID remains the zero value (uuid.Nil)

	return approvalbus.NewUserApprovalStatus{
		IconID:         iconID,
		Name:           app.Name,
		PrimaryColor:   app.PrimaryColor,
		SecondaryColor: app.SecondaryColor,
		Icon:           app.Icon,
	}, nil
}

type UpdateUserApprovalStatus struct {
	IconID         *string `json:"icon_id" validate:"omitempty,uuid"`
	Name           *string `json:"name" validate:"omitempty,min=3,max=100"`
	PrimaryColor   *string `json:"primary_color" validate:"omitempty,max=50"`
	SecondaryColor *string `json:"secondary_color" validate:"omitempty,max=50"`
	Icon           *string `json:"icon" validate:"omitempty,max=100"`
}

func (app *UpdateUserApprovalStatus) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app UpdateUserApprovalStatus) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusUpdateUserApprovalStatus(app UpdateUserApprovalStatus) (approvalbus.UpdateUserApprovalStatus, error) {
	var iconID *uuid.UUID
	var name *string

	if app.IconID != nil {
		if id, err := uuid.Parse(*app.IconID); err != nil {
			return approvalbus.UpdateUserApprovalStatus{}, err
		} else {
			iconID = &id
		}
	}

	if app.Name != nil {
		name = app.Name
	}

	return approvalbus.UpdateUserApprovalStatus{
		IconID:         iconID,
		Name:           name,
		PrimaryColor:   app.PrimaryColor,
		SecondaryColor: app.SecondaryColor,
		Icon:           app.Icon,
	}, nil
}
