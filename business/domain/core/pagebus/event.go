package pagebus

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
)

// DomainName represents the name of this domain.
const DomainName = "page"

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
	PageID uuid.UUID
	Page   Page
}

// String returns a string representation of the action parameters.
func (ac *ActionCreatedParms) String() string {
	return fmt.Sprintf("&EventParamsCreated{PageID:%v, Name:%v}", ac.PageID, ac.Page.Name)
}

// Marshal returns the event parameters encoded as JSON.
func (ac *ActionCreatedParms) Marshal() ([]byte, error) {
	return json.Marshal(ac)
}

// ActionCreatedData constructs the data for the created action.
func ActionCreatedData(page Page) delegate.Data {
	params := ActionCreatedParms{
		PageID: page.ID,
		Page:   page,
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
	PageID uuid.UUID
}

// String returns a string representation of the action parameters.
func (ar *ActionRetrievedParms) String() string {
	return fmt.Sprintf("&EventParamsRetrieved{PageID:%v}", ar.PageID)
}

// Marshal returns the event parameters encoded as JSON.
func (ar *ActionRetrievedParms) Marshal() ([]byte, error) {
	return json.Marshal(ar)
}

// ActionRetrievedData constructs the data for the retrieved action.
func ActionRetrievedData(pageID uuid.UUID) delegate.Data {
	params := ActionRetrievedParms{
		PageID: pageID,
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
	Page Page
}

// String returns a string representation of the action parameters.
func (au *ActionUpdatedParms) String() string {
	return fmt.Sprintf("&EventParamsUpdated{PageID:%v}", au.Page.ID)
}

// Marshal returns the event parameters encoded as JSON.
func (au *ActionUpdatedParms) Marshal() ([]byte, error) {
	return json.Marshal(au)
}

// ActionUpdatedData constructs the data for the updated action.
func ActionUpdatedData(p Page) delegate.Data {
	params := ActionUpdatedParms{
		Page: p,
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
	PageID uuid.UUID
}

// String returns a string representation of the action parameters.
func (ad *ActionDeletedParms) String() string {
	return fmt.Sprintf("&EventParamsDeleted{PageID:%v}", ad.PageID)
}

// Marshal returns the event parameters encoded as JSON.
func (ad *ActionDeletedParms) Marshal() ([]byte, error) {
	return json.Marshal(ad)
}

// ActionDeletedData constructs the data for the deleted action.
func ActionDeletedData(pageID uuid.UUID) delegate.Data {
	params := ActionDeletedParms{
		PageID: pageID,
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
