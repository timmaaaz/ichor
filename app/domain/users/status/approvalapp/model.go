package approvalapp

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/users/status/approvalbus"
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
	ID     string `json:"id"`
	IconID string `json:"icon_id"`
	Name   string `json:"name"`
}

func (app UserApprovalStatus) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppUserApprovalStatus(bus approvalbus.UserApprovalStatus) UserApprovalStatus {
	return UserApprovalStatus{
		ID:     bus.ID.String(),
		IconID: bus.IconID.String(),
		Name:   bus.Name,
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
	IconID string `json:"iconID" validate:"required"`
	Name   string `json:"name" validate:"required,min=3,max=100"`
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
	var err error

	if iconID, err = uuid.Parse(app.IconID); err != nil {
		return approvalbus.NewUserApprovalStatus{}, err
	}

	return approvalbus.NewUserApprovalStatus{
		IconID: iconID,
		Name:   app.Name, // TODO: Look at defining custom type
	}, nil
}

type UpdateUserApprovalStatus struct {
	IconID *string `json:"icon_id" validate:"required"`
	Name   *string `json:"name" validate:"required,min=3,max=100"`
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
		IconID: iconID,
		Name:   name,
	}, nil
}
