package pageactionbus

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
)

// DomainName represents the name of this domain for delegate events.
const DomainName = "pageaction"

// EntityName is the workflow entity name used for event matching.
// This should match the entity name in workflow.entities table.
// The entity is stored as just the table name (not schema-qualified).
const EntityName = "page_actions"

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
// Note: This is a reference/config table without user tracking fields.
// UserID is set to uuid.Nil for system-level operations.
type ActionCreatedParms struct {
	EntityID uuid.UUID  `json:"entityID"`
	UserID   uuid.UUID  `json:"userID"`
	Entity   PageAction `json:"entity"`
}

// Marshal returns the event parameters encoded as JSON.
func (p *ActionCreatedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// ActionCreatedData constructs delegate data for page action creation events.
func ActionCreatedData(action PageAction) delegate.Data {
	params := ActionCreatedParms{
		EntityID: action.ID,
		UserID:   uuid.Nil, // Config table - no user tracking
		Entity:   action,
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
	EntityID     uuid.UUID  `json:"entityID"`
	UserID       uuid.UUID  `json:"userID"`
	Entity       PageAction `json:"entity"`
	BeforeEntity PageAction `json:"beforeEntity,omitempty"`
}

// Marshal returns the event parameters encoded as JSON.
func (p *ActionUpdatedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// ActionUpdatedData constructs delegate data for page action update events.
func ActionUpdatedData(before, after PageAction) delegate.Data {
	params := ActionUpdatedParms{
		EntityID:     after.ID,
		UserID:       uuid.Nil, // Config table - no user tracking
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
	EntityID uuid.UUID  `json:"entityID"`
	UserID   uuid.UUID  `json:"userID"`
	Entity   PageAction `json:"entity"`
}

// Marshal returns the event parameters encoded as JSON.
func (p *ActionDeletedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// ActionDeletedData constructs delegate data for page action deletion events.
func ActionDeletedData(action PageAction) delegate.Data {
	params := ActionDeletedParms{
		EntityID: action.ID,
		UserID:   uuid.Nil, // Config table - no user tracking
		Entity:   action,
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
