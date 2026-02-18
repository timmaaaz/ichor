package inventoryadjustmentbus

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
)

// DomainName represents the name of this domain for delegate events.
const DomainName = "inventoryadjustment"

// EntityName is the workflow entity name used for event matching.
// This should match the entity name in workflow.entities table.
// The entity is stored as just the table name (not schema-qualified).
const EntityName = "inventory_adjustments"

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
// Note: UserID is taken from the entity's AdjustedBy field which tracks who made the adjustment.
type ActionCreatedParms struct {
	EntityID uuid.UUID           `json:"entityID"`
	UserID   uuid.UUID           `json:"userID"`
	Entity   InventoryAdjustment `json:"entity"`
}

// Marshal returns the event parameters encoded as JSON.
func (p *ActionCreatedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// ActionCreatedData constructs delegate data for inventory adjustment creation events.
func ActionCreatedData(ia InventoryAdjustment) delegate.Data {
	params := ActionCreatedParms{
		EntityID: ia.InventoryAdjustmentID,
		UserID:   ia.AdjustedBy, // AdjustedBy tracks who made the adjustment
		Entity:   ia,
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
	EntityID     uuid.UUID           `json:"entityID"`
	UserID       uuid.UUID           `json:"userID"`
	Entity       InventoryAdjustment `json:"entity"`
	BeforeEntity InventoryAdjustment `json:"beforeEntity,omitempty"`
}

// Marshal returns the event parameters encoded as JSON.
func (p *ActionUpdatedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// ActionUpdatedData constructs delegate data for inventory adjustment update events.
func ActionUpdatedData(before, after InventoryAdjustment) delegate.Data {
	params := ActionUpdatedParms{
		EntityID:     after.InventoryAdjustmentID,
		UserID:       after.AdjustedBy, // AdjustedBy tracks who made the adjustment
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
	EntityID uuid.UUID           `json:"entityID"`
	UserID   uuid.UUID           `json:"userID"`
	Entity   InventoryAdjustment `json:"entity"`
}

// Marshal returns the event parameters encoded as JSON.
func (p *ActionDeletedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// ActionDeletedData constructs delegate data for inventory adjustment deletion events.
func ActionDeletedData(ia InventoryAdjustment) delegate.Data {
	params := ActionDeletedParms{
		EntityID: ia.InventoryAdjustmentID,
		UserID:   ia.AdjustedBy, // AdjustedBy tracks who made the adjustment
		Entity:   ia,
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
