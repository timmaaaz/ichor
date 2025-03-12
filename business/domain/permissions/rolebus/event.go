package rolebus

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
)

// DomainName represents the name of this domain.
const DomainName = "role"

// Set of delegate actions for CRUD operations.
const (
	ActionCreated   = "created"
	ActionRetrieved = "retrieved"
	ActionUpdated   = "updated"
	ActionDeleted   = "deleted"
)

// ===============================================================
// Create Event

// ActionCreatedParms represents the parameters for the created action.
type ActionCreatedParms struct {
	RoleID uuid.UUID
	Role   Role
}

// String returns a string representation of the action parameters.
func (ac *ActionCreatedParms) String() string {
	return fmt.Sprintf("&EventParamsCreated{RoleID:%v, Name:%v}", ac.RoleID, ac.Role.Name)
}

// Marshal returns the event parameters encoded as JSON.
func (ac *ActionCreatedParms) Marshal() ([]byte, error) {
	return json.Marshal(ac)
}

// ActionCreatedData constructs the data for the created action.
func ActionCreatedData(role Role) delegate.Data {
	params := ActionCreatedParms{
		RoleID: role.ID,
		Role:   role,
	}

	rawParams, err := params.Marshal()
	if err != nil {
		panic(err)
	}

	return delegate.Data{
		Domain:    DomainName,
		Action:    ActionCreated,
		RawParams: rawParams,
	}
}

// ===============================================================
// Retrieved Event

// ActionRetrievedParms represents the parameters for the retrieved action.
type ActionRetrievedParms struct {
	RoleID uuid.UUID
}

// String returns a string representation of the action parameters.
func (ar *ActionRetrievedParms) String() string {
	return fmt.Sprintf("&EventParamsRetrieved{RoleID:%v}", ar.RoleID)
}

// Marshal returns the event parameters encoded as JSON.
func (ar *ActionRetrievedParms) Marshal() ([]byte, error) {
	return json.Marshal(ar)
}

// ActionRetrievedData constructs the data for the retrieved action.
func ActionRetrievedData(roleID uuid.UUID) delegate.Data {
	params := ActionRetrievedParms{
		RoleID: roleID,
	}

	rawParams, err := params.Marshal()
	if err != nil {
		panic(err)
	}

	return delegate.Data{
		Domain:    DomainName,
		Action:    ActionRetrieved,
		RawParams: rawParams,
	}
}

// ===============================================================
// Updated Event

// ActionUpdatedParms represents the parameters for the updated action.
type ActionUpdatedParms struct {
	Role Role
}

// String returns a string representation of the action parameters.
func (au *ActionUpdatedParms) String() string {
	return fmt.Sprintf("&EventParamsUpdated{RoleID:%v}", au.Role.ID)
}

// Marshal returns the event parameters encoded as JSON.
func (au *ActionUpdatedParms) Marshal() ([]byte, error) {
	return json.Marshal(au)
}

// ActionUpdatedData constructs the data for the updated action.
func ActionUpdatedData(r Role) delegate.Data {
	params := ActionUpdatedParms{
		Role: r,
	}

	rawParams, err := params.Marshal()
	if err != nil {
		panic(err)
	}

	return delegate.Data{
		Domain:    DomainName,
		Action:    ActionUpdated,
		RawParams: rawParams,
	}
}

// ===============================================================
// Deleted Event

// ActionDeletedParms represents the parameters for the deleted action.
type ActionDeletedParms struct {
	RoleID uuid.UUID
}

// String returns a string representation of the action parameters.
func (ad *ActionDeletedParms) String() string {
	return fmt.Sprintf("&EventParamsDeleted{RoleID:%v}", ad.RoleID)
}

// Marshal returns the event parameters encoded as JSON.
func (ad *ActionDeletedParms) Marshal() ([]byte, error) {
	return json.Marshal(ad)
}

// ActionDeletedData constructs the data for the deleted action.
func ActionDeletedData(roleID uuid.UUID) delegate.Data {
	params := ActionDeletedParms{
		RoleID: roleID,
	}

	rawParams, err := params.Marshal()
	if err != nil {
		panic(err)
	}

	return delegate.Data{
		Domain:    DomainName,
		Action:    ActionDeleted,
		RawParams: rawParams,
	}
}
