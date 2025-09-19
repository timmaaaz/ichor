package roleapp

import (
	"encoding/json"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/core/rolebus"
	"github.com/timmaaaz/ichor/business/sdk/convert"
)

type QueryParams struct {
	Page        string
	Rows        string
	OrderBy     string
	ID          string
	Name        string
	Description string
}

type Role struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (app Role) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppRole(bus rolebus.Role) Role {
	return Role{
		ID:          bus.ID.String(),
		Name:        bus.Name,
		Description: bus.Description,
	}
}

func ToAppRoles(bus []rolebus.Role) []Role {
	app := make([]Role, len(bus))
	for i, v := range bus {
		app[i] = ToAppRole(v)
	}
	return app
}

// =============================================================================

type NewRole struct {
	Name        string `json:"name" validate:"required,min=3,max=50"`
	Description string `json:"description" validate:"required"`
}

func (app *NewRole) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewRole) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusNewRole(app NewRole) (rolebus.NewRole, error) {
	dest := rolebus.NewRole{}
	err := convert.PopulateTypesFromStrings(app, &dest)
	return dest, err
}

// =============================================================================

type UpdateRole struct {
	Name        *string `json:"name" validate:"omitempty,min=3,max=50"`
	Description *string `json:"description" validate:"omitempty"`
}

// Decode implements the decoder interface.
func (app *UpdateRole) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks the data in the model is considered clean.
func (app UpdateRole) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusUpdateRole(app UpdateRole) (rolebus.UpdateRole, error) {
	dest := rolebus.UpdateRole{}
	err := convert.PopulateTypesFromStrings(app, &dest)

	return dest, err
}
