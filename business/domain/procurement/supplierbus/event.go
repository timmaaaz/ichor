package supplierbus

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
)

// DomainName represents the name of this domain for delegate events.
const DomainName = "supplier"

// EntityName is the workflow entity name used for event matching.
// This should match the entity name in workflow.entities table.
// The entity is stored as just the table name (not schema-qualified).
const EntityName = "suppliers"

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
// Note: This entity does not have user tracking fields (CreatedBy/UpdatedBy).
// UserID is set to uuid.Nil for system-level operations.
type ActionCreatedParms struct {
	EntityID uuid.UUID `json:"entityID"`
	UserID   uuid.UUID `json:"userID"`
	Entity   Supplier  `json:"entity"`
}

// Marshal returns the event parameters encoded as JSON.
func (p *ActionCreatedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// ActionCreatedData constructs delegate data for supplier creation events.
func ActionCreatedData(supplier Supplier) delegate.Data {
	params := ActionCreatedParms{
		EntityID: supplier.SupplierID,
		UserID:   uuid.Nil, // No user tracking on this entity
		Entity:   supplier,
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
	EntityID uuid.UUID `json:"entityID"`
	UserID   uuid.UUID `json:"userID"`
	Entity   Supplier  `json:"entity"`
}

// Marshal returns the event parameters encoded as JSON.
func (p *ActionUpdatedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// ActionUpdatedData constructs delegate data for supplier update events.
func ActionUpdatedData(supplier Supplier) delegate.Data {
	params := ActionUpdatedParms{
		EntityID: supplier.SupplierID,
		UserID:   uuid.Nil, // No user tracking on this entity
		Entity:   supplier,
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
	EntityID uuid.UUID `json:"entityID"`
	UserID   uuid.UUID `json:"userID"`
	Entity   Supplier  `json:"entity"`
}

// Marshal returns the event parameters encoded as JSON.
func (p *ActionDeletedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// ActionDeletedData constructs delegate data for supplier deletion events.
func ActionDeletedData(supplier Supplier) delegate.Data {
	params := ActionDeletedParms{
		EntityID: supplier.SupplierID,
		UserID:   uuid.Nil, // No user tracking on this entity
		Entity:   supplier,
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
