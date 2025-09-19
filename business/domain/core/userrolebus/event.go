package userrolebus

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
)

// DomainName represents the name of this domain.
const DomainName = "userrole"

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
	UserRoleID uuid.UUID
	UserRole   UserRole
}

// String returns a string representation of the action parameters.
func (ac *ActionCreatedParms) String() string {
	return fmt.Sprintf("&EventParamsCreated{UserRoleID:%v, UserID:%v, RoleID:%v}",
		ac.UserRoleID, ac.UserRole.UserID, ac.UserRole.RoleID)
}

// Marshal returns the event parameters encoded as JSON.
func (ac *ActionCreatedParms) Marshal() ([]byte, error) {
	return json.Marshal(ac)
}

// ActionCreatedData constructs the data for the created action.
func ActionCreatedData(userRole UserRole) delegate.Data {
	params := ActionCreatedParms{
		UserRoleID: userRole.ID,
		UserRole:   userRole,
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
	UserRoleID uuid.UUID
}

// String returns a string representation of the action parameters.
func (ar *ActionRetrievedParms) String() string {
	return fmt.Sprintf("&EventParamsRetrieved{UserRoleID:%v}", ar.UserRoleID)
}

// Marshal returns the event parameters encoded as JSON.
func (ar *ActionRetrievedParms) Marshal() ([]byte, error) {
	return json.Marshal(ar)
}

// ActionRetrievedData constructs the data for the retrieved action.
func ActionRetrievedData(userRoleID uuid.UUID) delegate.Data {
	params := ActionRetrievedParms{
		UserRoleID: userRoleID,
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
	UserRole UserRole
}

// String returns a string representation of the action parameters.
func (au *ActionUpdatedParms) String() string {
	return fmt.Sprintf("&EventParamsUpdated{UserRoleID:%+v}", au.UserRole)
}

// Marshal returns the event parameters encoded as JSON.
func (au *ActionUpdatedParms) Marshal() ([]byte, error) {
	return json.Marshal(au)
}

// ActionUpdatedData constructs the data for the updated action.
func ActionUpdatedData(ur UserRole) delegate.Data {
	params := ActionUpdatedParms{
		UserRole: ur,
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
	UserRoleID uuid.UUID
}

// String returns a string representation of the action parameters.
func (ad *ActionDeletedParms) String() string {
	return fmt.Sprintf("&EventParamsDeleted{UserRoleID:%v}", ad.UserRoleID)
}

// Marshal returns the event parameters encoded as JSON.
func (ad *ActionDeletedParms) Marshal() ([]byte, error) {
	return json.Marshal(ad)
}

// ActionDeletedData constructs the data for the deleted action.
func ActionDeletedData(userRoleID uuid.UUID) delegate.Data {
	params := ActionDeletedParms{
		UserRoleID: userRoleID,
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
