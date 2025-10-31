package rolepageapp

import (
	"encoding/json"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/core/rolepagebus"
	"github.com/timmaaaz/ichor/business/sdk/convert"
)

type QueryParams struct {
	Page       string
	Rows       string
	OrderBy    string
	ID         string
	RoleID     string
	PageID     string
	CanAccess  string
	ShowInMenu string
}

type RolePage struct {
	ID         string `json:"id"`
	RoleID     string `json:"roleId"`
	PageID     string `json:"pageId"`
	CanAccess  bool   `json:"canAccess"`
	ShowInMenu bool   `json:"showInMenu"`
}

func (app RolePage) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppRolePage(bus rolepagebus.RolePage) RolePage {
	return RolePage{
		ID:         bus.ID.String(),
		RoleID:     bus.RoleID.String(),
		PageID:     bus.PageID.String(),
		CanAccess:  bus.CanAccess,
		ShowInMenu: bus.ShowInMenu,
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
	RoleID     string `json:"roleId" validate:"required"`
	PageID     string `json:"pageId" validate:"required"`
	CanAccess  bool   `json:"canAccess"`
	ShowInMenu bool   `json:"showInMenu"`
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
	dest := rolepagebus.NewRolePage{}
	err := convert.PopulateTypesFromStrings(app, &dest)
	return dest, err
}

// =============================================================================

type UpdateRolePage struct {
	CanAccess  *bool `json:"canAccess"`
	ShowInMenu *bool `json:"showInMenu"`
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
	dest := rolepagebus.UpdateRolePage{}
	err := convert.PopulateTypesFromStrings(app, &dest)

	return dest, err
}
