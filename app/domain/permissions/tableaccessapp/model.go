package tableaccessapp

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/permissions/tableaccessbus"
)

type QueryParams struct {
	Page      string
	Rows      string
	OrderBy   string
	ID        string
	RoleID    string
	TableName string
	CanCreate string
	CanRead   string
	CanUpdate string
	CanDelete string
}

type TableAccess struct {
	ID        string `json:"id"`
	RoleID    string `json:"role_id"`
	TableName string `json:"table_name"`
	CanCreate bool   `json:"can_create"`
	CanRead   bool   `json:"can_read"`
	CanUpdate bool   `json:"can_update"`
	CanDelete bool   `json:"can_delete"`
}

func (app TableAccess) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

type TableAccesses []TableAccess

func (app TableAccesses) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppTableAccess(bus tableaccessbus.TableAccess) TableAccess {
	return TableAccess{
		ID:        bus.ID.String(),
		RoleID:    bus.RoleID.String(),
		TableName: bus.TableName,
		CanCreate: bus.CanCreate,
		CanRead:   bus.CanRead,
		CanUpdate: bus.CanUpdate,
		CanDelete: bus.CanDelete,
	}
}

func ToAppTableAccesses(bus []tableaccessbus.TableAccess) []TableAccess {
	app := make([]TableAccess, len(bus))
	for i, v := range bus {
		app[i] = ToAppTableAccess(v)
	}
	return app
}

// =============================================================================

type NewTableAccess struct {
	RoleID    string `json:"role_id" validate:"required"`
	TableName string `json:"table_name" validate:"required"`
	CanCreate bool   `json:"can_create"`
	CanRead   bool   `json:"can_read"`
	CanUpdate bool   `json:"can_update"`
	CanDelete bool   `json:"can_delete"`
}

func (app *NewTableAccess) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewTableAccess) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusNewTableAccess(app NewTableAccess) (tableaccessbus.NewTableAccess, error) {
	rID := uuid.MustParse(app.RoleID)
	return tableaccessbus.NewTableAccess{
		RoleID:    rID,
		TableName: app.TableName,
		CanCreate: app.CanCreate,
		CanRead:   app.CanRead,
		CanUpdate: app.CanUpdate,
		CanDelete: app.CanDelete,
	}, nil

}

// =============================================================================

type UpdateTableAccess struct {
	RoleID    *string `json:"role_id" validate:"omitempty,uuid"`
	TableName *string `json:"table_name"`
	CanCreate *bool   `json:"can_create"`
	CanRead   *bool   `json:"can_read"`
	CanUpdate *bool   `json:"can_update"`
	CanDelete *bool   `json:"can_delete"`
}

func (app *UpdateTableAccess) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app UpdateTableAccess) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusUpdateTableAccess(app UpdateTableAccess) (tableaccessbus.UpdateTableAccess, error) {
	// Create the business update object
	update := tableaccessbus.UpdateTableAccess{}

	// Handle role_id if provided
	if app.RoleID != nil {
		tmp, err := uuid.Parse(*app.RoleID)
		if err != nil {
			return tableaccessbus.UpdateTableAccess{}, errs.Newf(errs.InvalidArgument, "toBusUpdateTableAccess: %s", err)
		}
		update.RoleID = &tmp
	}

	// Handle table name if provided
	update.TableName = app.TableName

	// Handle permission fields if they are provided
	if app.CanCreate != nil {
		update.CanCreate = app.CanCreate
	}

	if app.CanRead != nil {
		update.CanRead = app.CanRead
	}

	if app.CanUpdate != nil {
		update.CanUpdate = app.CanUpdate
	}

	if app.CanDelete != nil {
		update.CanDelete = app.CanDelete
	}

	return update, nil
}
