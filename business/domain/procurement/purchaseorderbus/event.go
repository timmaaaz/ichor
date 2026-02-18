package purchaseorderbus

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
)

// DomainName represents the name of this domain for delegate events.
const DomainName = "purchaseorder"

// EntityName is the workflow entity name used for event matching.
// This should match the entity name in workflow.entities table.
// The entity is stored as just the table name (not schema-qualified).
const EntityName = "purchase_orders"

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
type ActionCreatedParms struct {
	EntityID uuid.UUID     `json:"entityID"`
	UserID   uuid.UUID     `json:"userID"`
	Entity   PurchaseOrder `json:"entity"`
}

// Marshal returns the event parameters encoded as JSON.
func (p *ActionCreatedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// ActionCreatedData constructs delegate data for purchase order creation events.
func ActionCreatedData(po PurchaseOrder) delegate.Data {
	params := ActionCreatedParms{
		EntityID: po.ID,
		UserID:   po.CreatedBy,
		Entity:   po,
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
	Entity       PurchaseOrder `json:"entity"`
	BeforeEntity PurchaseOrder `json:"beforeEntity,omitempty"`
}

// Marshal returns the event parameters encoded as JSON.
func (p *ActionUpdatedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// ActionUpdatedData constructs delegate data for purchase order update events.
func ActionUpdatedData(before, after PurchaseOrder) delegate.Data {
	params := ActionUpdatedParms{
		EntityID:     after.ID,
		UserID:       after.UpdatedBy,
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
	Entity   PurchaseOrder `json:"entity"`
}

// Marshal returns the event parameters encoded as JSON.
func (p *ActionDeletedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// ActionDeletedData constructs delegate data for purchase order deletion events.
// Note: For delete, we use UpdatedBy as the user who performed the delete.
func ActionDeletedData(po PurchaseOrder) delegate.Data {
	params := ActionDeletedParms{
		EntityID: po.ID,
		UserID:   po.UpdatedBy, // UpdatedBy tracks who performed the delete
		Entity:   po,
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
