package rolepageapp

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/core/rolepagebus"
)

type QueryParams struct {
	Page      string
	Rows      string
	OrderBy   string
	ID        string
	RoleID    string
	PageID    string
	CanAccess string
}

type RolePage struct {
	ID        string `json:"id"`
	RoleID    string `json:"role_id"`
	PageID    string `json:"page_id"`
	CanAccess bool   `json:"can_access"`
}

func (app RolePage) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppRolePage(bus rolepagebus.RolePage) RolePage {
	return RolePage{
		ID:        bus.ID.String(),
		RoleID:    bus.RoleID.String(),
		PageID:    bus.PageID.String(),
		CanAccess: bus.CanAccess,
	}
}

func ToAppRolePages(bus []rolepagebus.RolePage) []RolePage {
	app := make([]RolePage, len(bus))
	for i, v := range bus {
		app[i] = ToAppRolePage(v)
	}
	return app
}

// =============================================================================

type NewRolePage struct {
	RoleID    string `json:"role_id" validate:"required"`
	PageID    string `json:"page_id" validate:"required"`
	CanAccess bool   `json:"can_access"`
}

func (app *NewRolePage) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewRolePage) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusNewRolePage(app NewRolePage) (rolepagebus.NewRolePage, error) {
	roleID, err := uuid.Parse(app.RoleID)
	if err != nil {
		return rolepagebus.NewRolePage{}, errs.Newf(errs.InvalidArgument, "parse roleID: %s", err)
	}

	pageID, err := uuid.Parse(app.PageID)
	if err != nil {
		return rolepagebus.NewRolePage{}, errs.Newf(errs.InvalidArgument, "parse pageID: %s", err)
	}

	bus := rolepagebus.NewRolePage{
		RoleID:    roleID,
		PageID:    pageID,
		CanAccess: app.CanAccess,
	}
	return bus, nil
}

// =============================================================================

type UpdateRolePage struct {
	CanAccess *bool `json:"can_access"`
}

// Decode implements the decoder interface.
func (app *UpdateRolePage) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks the data in the model is considered clean.
func (app UpdateRolePage) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusUpdateRolePage(app UpdateRolePage) (rolepagebus.UpdateRolePage, error) {
	bus := rolepagebus.UpdateRolePage{
		CanAccess: app.CanAccess,
	}
	return bus, nil
}
