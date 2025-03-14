package userroleapp

import (
	"encoding/json"

	"github.com/timmaaaz/ichor/business/domain/permissions/userrolebus"
	"github.com/timmaaaz/ichor/foundation/convert"
)

type QueryParams struct {
	Page    string
	Rows    string
	OrderBy string
	ID      string
	UserID  string
	RoleID  string
}

type UserRole struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
	RoleID string `json:"role_id"`
}

func (app UserRole) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppUserRole(bus userrolebus.UserRole) UserRole {
	return UserRole{
		ID:     bus.ID.String(),
		UserID: bus.UserID.String(),
		RoleID: bus.RoleID.String(),
	}
}

func ToAppUserRoles(bus []userrolebus.UserRole) []UserRole {
	app := make([]UserRole, len(bus))
	for i, v := range bus {
		app[i] = ToAppUserRole(v)
	}
	return app
}

// =============================================================================

type NewUserRole struct {
	UserID string `json:"user_id" validate:"required"`
	RoleID string `json:"role_id" validate:"required"`
}

func (app *NewUserRole) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewUserRole) Validate() error {
	return nil
}

func toBusNewUserRole(app NewUserRole) (userrolebus.NewUserRole, error) {
	dest := userrolebus.NewUserRole{}
	err := convert.PopulateTypesFromStrings(app, &dest)
	return dest, err
}
