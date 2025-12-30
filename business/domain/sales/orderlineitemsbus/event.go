package orderlineitemsbus

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
)

// DomainName represents the name of this domain for delegate events.
const DomainName = "orderlineitem"

// EntityName is the workflow entity name used for event matching.
// This should match the entity name in workflow.entities table.
// The entity is stored as just the table name (not schema-qualified).
const EntityName = "order_line_items"

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
	Entity   OrderLineItem `json:"entity"`
}

// Marshal returns the event parameters encoded as JSON.
func (p *ActionCreatedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// ActionCreatedData constructs delegate data for order line item creation events.
func ActionCreatedData(lineItem OrderLineItem) delegate.Data {
	params := ActionCreatedParms{
		EntityID: lineItem.ID,
		UserID:   lineItem.CreatedBy,
		Entity:   lineItem,
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
	EntityID uuid.UUID     `json:"entityID"`
	UserID   uuid.UUID     `json:"userID"`
	Entity   OrderLineItem `json:"entity"`
}

// Marshal returns the event parameters encoded as JSON.
func (p *ActionUpdatedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// ActionUpdatedData constructs delegate data for order line item update events.
func ActionUpdatedData(lineItem OrderLineItem) delegate.Data {
	params := ActionUpdatedParms{
		EntityID: lineItem.ID,
		UserID:   lineItem.UpdatedBy,
		Entity:   lineItem,
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
	Entity   OrderLineItem `json:"entity"`
}

// Marshal returns the event parameters encoded as JSON.
func (p *ActionDeletedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// ActionDeletedData constructs delegate data for order line item deletion events.
// Note: For delete, we use UpdatedBy as the user who performed the delete.
func ActionDeletedData(lineItem OrderLineItem) delegate.Data {
	params := ActionDeletedParms{
		EntityID: lineItem.ID,
		UserID:   lineItem.UpdatedBy, // UpdatedBy tracks who performed the delete
		Entity:   lineItem,
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
