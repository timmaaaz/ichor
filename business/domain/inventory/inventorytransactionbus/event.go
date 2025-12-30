package inventorytransactionbus

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
)

// DomainName represents the name of this domain for delegate events.
const DomainName = "inventorytransaction"

// EntityName is the workflow entity name used for event matching.
// This should match the entity name in workflow.entities table.
// The entity is stored as just the table name (not schema-qualified).
const EntityName = "inventory_transactions"

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
// Note: UserID is taken from the entity's UserID field which tracks who performed the transaction.
type ActionCreatedParms struct {
	EntityID uuid.UUID            `json:"entityID"`
	UserID   uuid.UUID            `json:"userID"`
	Entity   InventoryTransaction `json:"entity"`
}

// Marshal returns the event parameters encoded as JSON.
func (p *ActionCreatedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// ActionCreatedData constructs delegate data for inventory transaction creation events.
func ActionCreatedData(it InventoryTransaction) delegate.Data {
	params := ActionCreatedParms{
		EntityID: it.InventoryTransactionID,
		UserID:   it.UserID, // UserID tracks who performed the transaction
		Entity:   it,
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
	EntityID uuid.UUID            `json:"entityID"`
	UserID   uuid.UUID            `json:"userID"`
	Entity   InventoryTransaction `json:"entity"`
}

// Marshal returns the event parameters encoded as JSON.
func (p *ActionUpdatedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// ActionUpdatedData constructs delegate data for inventory transaction update events.
func ActionUpdatedData(it InventoryTransaction) delegate.Data {
	params := ActionUpdatedParms{
		EntityID: it.InventoryTransactionID,
		UserID:   it.UserID, // UserID tracks who performed the transaction
		Entity:   it,
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
	EntityID uuid.UUID            `json:"entityID"`
	UserID   uuid.UUID            `json:"userID"`
	Entity   InventoryTransaction `json:"entity"`
}

// Marshal returns the event parameters encoded as JSON.
func (p *ActionDeletedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// ActionDeletedData constructs delegate data for inventory transaction deletion events.
func ActionDeletedData(it InventoryTransaction) delegate.Data {
	params := ActionDeletedParms{
		EntityID: it.InventoryTransactionID,
		UserID:   it.UserID, // UserID tracks who performed the transaction
		Entity:   it,
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
