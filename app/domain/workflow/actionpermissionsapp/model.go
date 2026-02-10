// Package actionpermissionsapp provides the application layer for managing
// action permissions that control which roles can execute workflow actions manually.
package actionpermissionsapp

import (
	"encoding/json"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/workflow/actionpermissionsbus"
)

// QueryParams holds the query parameters for listing action permissions.
type QueryParams struct {
	Page       string
	Rows       string
	OrderBy    string
	ID         string
	RoleID     string
	ActionType string
	IsAllowed  string
}

// =============================================================================
// ActionPermission Response
// =============================================================================

// ActionPermission represents a permission for a role to execute a specific action type.
type ActionPermission struct {
	ID          string          `json:"id"`
	RoleID      string          `json:"roleId"`
	ActionType  string          `json:"actionType"`
	IsAllowed   bool            `json:"isAllowed"`
	Constraints json.RawMessage `json:"constraints"`
	CreatedAt   string          `json:"createdAt"`
	UpdatedAt   string          `json:"updatedAt"`
}

// Encode implements the encoder interface.
func (app ActionPermission) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// ToAppActionPermission converts a business layer ActionPermission to an app ActionPermission.
func ToAppActionPermission(bus actionpermissionsbus.ActionPermission) ActionPermission {
	return ActionPermission{
		ID:          bus.ID.String(),
		RoleID:      bus.RoleID.String(),
		ActionType:  bus.ActionType,
		IsAllowed:   bus.IsAllowed,
		Constraints: bus.Constraints,
		CreatedAt:   bus.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   bus.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// ToAppActionPermissions converts a slice of business layer ActionPermissions to app ActionPermissions.
func ToAppActionPermissions(bus []actionpermissionsbus.ActionPermission) []ActionPermission {
	app := make([]ActionPermission, len(bus))
	for i, v := range bus {
		app[i] = ToAppActionPermission(v)
	}
	return app
}

// =============================================================================
// NewActionPermission Request
// =============================================================================

// NewActionPermission contains information needed to create a new action permission.
type NewActionPermission struct {
	RoleID      string          `json:"roleId" validate:"required,uuid"`
	ActionType  string          `json:"actionType" validate:"required,min=1,max=100"`
	IsAllowed   bool            `json:"isAllowed"`
	Constraints json.RawMessage `json:"constraints,omitempty"`
}

// Decode implements the decoder interface.
func (app *NewActionPermission) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks the data in the model is considered clean.
func (app NewActionPermission) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

// =============================================================================
// UpdateActionPermission Request
// =============================================================================

// UpdateActionPermission contains information needed to update an action permission.
type UpdateActionPermission struct {
	IsAllowed   *bool            `json:"isAllowed,omitempty"`
	Constraints *json.RawMessage `json:"constraints,omitempty"`
}

// Decode implements the decoder interface.
func (app *UpdateActionPermission) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks the data in the model is considered clean.
func (app UpdateActionPermission) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}
