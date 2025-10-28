package rolepagebus

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
)

// DomainName represents the name of this domain.
const DomainName = "rolepage"

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
	RolePageID uuid.UUID
	RolePage   RolePage
}

// String returns a string representation of the action parameters.
func (ac *ActionCreatedParms) String() string {
	return fmt.Sprintf("&EventParamsCreated{RolePageID:%v, RoleID:%v, PageID:%v}", ac.RolePageID, ac.RolePage.RoleID, ac.RolePage.PageID)
}

// Marshal returns the event parameters encoded as JSON.
func (ac *ActionCreatedParms) Marshal() ([]byte, error) {
	return json.Marshal(ac)
}

// ActionCreatedData constructs the data for the created action.
func ActionCreatedData(rolePage RolePage) delegate.Data {
	params := ActionCreatedParms{
		RolePageID: rolePage.ID,
		RolePage:   rolePage,
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
	RolePageID uuid.UUID
}

// String returns a string representation of the action parameters.
func (ar *ActionRetrievedParms) String() string {
	return fmt.Sprintf("&EventParamsRetrieved{RolePageID:%v}", ar.RolePageID)
}

// Marshal returns the event parameters encoded as JSON.
func (ar *ActionRetrievedParms) Marshal() ([]byte, error) {
	return json.Marshal(ar)
}

// ActionRetrievedData constructs the data for the retrieved action.
func ActionRetrievedData(rolePageID uuid.UUID) delegate.Data {
	params := ActionRetrievedParms{
		RolePageID: rolePageID,
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
	RolePage RolePage
}

// String returns a string representation of the action parameters.
func (au *ActionUpdatedParms) String() string {
	return fmt.Sprintf("&EventParamsUpdated{RolePageID:%v}", au.RolePage.ID)
}

// Marshal returns the event parameters encoded as JSON.
func (au *ActionUpdatedParms) Marshal() ([]byte, error) {
	return json.Marshal(au)
}

// ActionUpdatedData constructs the data for the updated action.
func ActionUpdatedData(rp RolePage) delegate.Data {
	params := ActionUpdatedParms{
		RolePage: rp,
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
	RolePageID uuid.UUID
}

// String returns a string representation of the action parameters.
func (ad *ActionDeletedParms) String() string {
	return fmt.Sprintf("&EventParamsDeleted{RolePageID:%v}", ad.RolePageID)
}

// Marshal returns the event parameters encoded as JSON.
func (ad *ActionDeletedParms) Marshal() ([]byte, error) {
	return json.Marshal(ad)
}

// ActionDeletedData constructs the data for the deleted action.
func ActionDeletedData(rolePageID uuid.UUID) delegate.Data {
	params := ActionDeletedParms{
		RolePageID: rolePageID,
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
