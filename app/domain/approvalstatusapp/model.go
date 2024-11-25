package approvalstatusapp

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/approvalstatusbus"
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
	ID     string `json:"id"`
	IconID string `json:"icon_id"`
	Name   string `json:"name"`
}

func (app ApprovalStatus) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppApprovalStatus(bus approvalstatusbus.ApprovalStatus) ApprovalStatus {
	return ApprovalStatus{
		ID:     bus.ID.String(),
		IconID: bus.IconID.String(),
		Name:   bus.Name,
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
	IconID string `json:"iconID" validate:"required"`
	Name   string `json:"name" validate:"required,min=3,max=100"`
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
		IconID: iconID,
		Name:   app.Name, // TODO: Look at defining custom type
	}, nil
}

type UpdateApprovalStatus struct {
	IconID *string `json:"iconID"`
	Name   *string `json:"name"`
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
		IconID: iconID,
		Name:   name,
	}, nil
}
