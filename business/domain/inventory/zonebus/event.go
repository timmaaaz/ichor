package zonebus

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
)

// DomainName represents the name of this domain for delegate events.
const DomainName = "zone"

// EntityName is the workflow entity name used for event matching.
// This should match the entity name in workflow.entities table.
// The entity is stored as just the table name (not schema-qualified).
const EntityName = "zones"

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
// Note: This is a reference/lookup table without user tracking fields.
// UserID is set to uuid.Nil for system-level operations.
type ActionCreatedParms struct {
	EntityID uuid.UUID `json:"entityID"`
	UserID   uuid.UUID `json:"userID"`
	Entity   Zone      `json:"entity"`
}

// Marshal returns the event parameters encoded as JSON.
func (p *ActionCreatedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// ActionCreatedData constructs delegate data for zone creation events.
func ActionCreatedData(zone Zone) delegate.Data {
	params := ActionCreatedParms{
		EntityID: zone.ZoneID,
		UserID:   uuid.Nil, // Reference table - no user tracking
		Entity:   zone,
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
	Entity   Zone      `json:"entity"`
}

// Marshal returns the event parameters encoded as JSON.
func (p *ActionUpdatedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// ActionUpdatedData constructs delegate data for zone update events.
func ActionUpdatedData(zone Zone) delegate.Data {
	params := ActionUpdatedParms{
		EntityID: zone.ZoneID,
		UserID:   uuid.Nil, // Reference table - no user tracking
		Entity:   zone,
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
	Entity   Zone      `json:"entity"`
}

// Marshal returns the event parameters encoded as JSON.
func (p *ActionDeletedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// ActionDeletedData constructs delegate data for zone deletion events.
func ActionDeletedData(zone Zone) delegate.Data {
	params := ActionDeletedParms{
		EntityID: zone.ZoneID,
		UserID:   uuid.Nil, // Reference table - no user tracking
		Entity:   zone,
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
