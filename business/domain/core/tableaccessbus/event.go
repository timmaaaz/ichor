package tableaccessbus

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
)

// DomainName represents the name of this domain.
const DomainName = "tableaccess"

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
	TableAccessID uuid.UUID
	TableAccess   TableAccess
}

// String returns a string representation of the action parameters.
func (ac *ActionCreatedParms) String() string {
	return fmt.Sprintf("&EventParamsCreated{TableAccessID:%v, RoleID:%v, TableName:%v}",
		ac.TableAccessID, ac.TableAccess.RoleID, ac.TableAccess.TableName)
}

// Marshal returns the event parameters encoded as JSON.
func (ac *ActionCreatedParms) Marshal() ([]byte, error) {
	return json.Marshal(ac)
}

// ActionCreatedData constructs the data for the created action.
func ActionCreatedData(tableAccess TableAccess) delegate.Data {
	params := ActionCreatedParms{
		TableAccessID: tableAccess.ID,
		TableAccess:   tableAccess,
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
	TableAccessID uuid.UUID
}

// String returns a string representation of the action parameters.
func (ar *ActionRetrievedParms) String() string {
	return fmt.Sprintf("&EventParamsRetrieved{TableAccessID:%v}", ar.TableAccessID)
}

// Marshal returns the event parameters encoded as JSON.
func (ar *ActionRetrievedParms) Marshal() ([]byte, error) {
	return json.Marshal(ar)
}

// ActionRetrievedData constructs the data for the retrieved action.
func ActionRetrievedData(tableAccessID uuid.UUID) delegate.Data {
	params := ActionRetrievedParms{
		TableAccessID: tableAccessID,
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
	TableAccess TableAccess
}

// String returns a string representation of the action parameters.
func (au *ActionUpdatedParms) String() string {
	return fmt.Sprintf("&EventParamsUpdated{TableAccessID:%+v}", au.TableAccess)
}

// Marshal returns the event parameters encoded as JSON.
func (au *ActionUpdatedParms) Marshal() ([]byte, error) {
	return json.Marshal(au)
}

// ActionUpdatedData constructs the data for the updated action.
func ActionUpdatedData(ta TableAccess) delegate.Data {
	params := ActionUpdatedParms{
		TableAccess: ta,
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
	TableAccessID uuid.UUID
}

// String returns a string representation of the action parameters.
func (ad *ActionDeletedParms) String() string {
	return fmt.Sprintf("&EventParamsDeleted{TableAccessID:%v}", ad.TableAccessID)
}

// Marshal returns the event parameters encoded as JSON.
func (ad *ActionDeletedParms) Marshal() ([]byte, error) {
	return json.Marshal(ad)
}

// ActionDeletedData constructs the data for the deleted action.
func ActionDeletedData(tableAccessID uuid.UUID) delegate.Data {
	params := ActionDeletedParms{
		TableAccessID: tableAccessID,
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
