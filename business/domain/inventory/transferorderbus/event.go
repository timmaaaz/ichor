package transferorderbus

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
)

// DomainName represents the name of this domain for delegate events.
const DomainName = "transferorder"

// EntityName is the workflow entity name used for event matching.
// This should match the entity name in workflow.entities table.
// The entity is stored as just the table name (not schema-qualified).
const EntityName = "transfer_orders"

// Delegate action constants.
const (
	ActionCreated = "created"
	ActionUpdated = "updated"
	ActionDeleted = "deleted"
)

// =============================================================================
// Created Event
// =============================================================================

// ActionCreatedParms represents the parameters for the created action.
// Note: UserID is taken from the entity's RequestedByID field which tracks who requested the transfer.
type ActionCreatedParms struct {
	EntityID uuid.UUID     `json:"entityID"`
	UserID   uuid.UUID     `json:"userID"`
	Entity   TransferOrder `json:"entity"`
}

// Marshal returns the event parameters encoded as JSON.
func (p *ActionCreatedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// ActionCreatedData constructs delegate data for transfer order creation events.
func ActionCreatedData(to TransferOrder) delegate.Data {
	params := ActionCreatedParms{
		EntityID: to.TransferID,
		UserID:   to.RequestedByID, // RequestedByID tracks who requested the transfer
		Entity:   to,
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

// =============================================================================
// Updated Event
// =============================================================================

// ActionUpdatedParms represents the parameters for the updated action.
type ActionUpdatedParms struct {
	EntityID     uuid.UUID     `json:"entityID"`
	UserID       uuid.UUID     `json:"userID"`
	Entity       TransferOrder `json:"entity"`
	BeforeEntity TransferOrder `json:"beforeEntity,omitempty"`
}

// Marshal returns the event parameters encoded as JSON.
func (p *ActionUpdatedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// ActionUpdatedData constructs delegate data for transfer order update events.
func ActionUpdatedData(before, after TransferOrder) delegate.Data {
	params := ActionUpdatedParms{
		EntityID:     after.TransferID,
		UserID:       after.RequestedByID, // RequestedByID tracks who requested the transfer
		Entity:       after,
		BeforeEntity: before,
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

// =============================================================================
// Deleted Event
// =============================================================================

// ActionDeletedParms represents the parameters for the deleted action.
type ActionDeletedParms struct {
	EntityID uuid.UUID     `json:"entityID"`
	UserID   uuid.UUID     `json:"userID"`
	Entity   TransferOrder `json:"entity"`
}

// Marshal returns the event parameters encoded as JSON.
func (p *ActionDeletedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// ActionDeletedData constructs delegate data for transfer order deletion events.
func ActionDeletedData(to TransferOrder) delegate.Data {
	params := ActionDeletedParms{
		EntityID: to.TransferID,
		UserID:   to.RequestedByID, // RequestedByID tracks who requested the transfer
		Entity:   to,
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
