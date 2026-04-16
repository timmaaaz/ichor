package labelbus

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
)

// DomainName represents the name of this domain for delegate events.
const DomainName = "label"

// EntityName is the workflow entity name used for event matching.
// Must match the entity name in workflow.entities (the bare table
// name, not schema-qualified).
const EntityName = "label_catalog"

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
// LabelCatalog is a reference/lookup table without user tracking fields,
// so UserID is set to uuid.Nil (matches pagebus/rolebus convention).
type ActionCreatedParms struct {
	EntityID uuid.UUID    `json:"entityID"`
	UserID   uuid.UUID    `json:"userID"`
	Entity   LabelCatalog `json:"entity"`
}

// Marshal returns the event parameters encoded as JSON.
func (p *ActionCreatedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// ActionCreatedData constructs delegate data for label creation events.
func ActionCreatedData(lc LabelCatalog) delegate.Data {
	params := ActionCreatedParms{
		EntityID: lc.ID,
		UserID:   uuid.Nil,
		Entity:   lc,
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
	EntityID     uuid.UUID    `json:"entityID"`
	UserID       uuid.UUID    `json:"userID"`
	Entity       LabelCatalog `json:"entity"`
	BeforeEntity LabelCatalog `json:"beforeEntity,omitempty"`
}

// Marshal returns the event parameters encoded as JSON.
func (p *ActionUpdatedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// ActionUpdatedData constructs delegate data for label update events.
func ActionUpdatedData(before, after LabelCatalog) delegate.Data {
	params := ActionUpdatedParms{
		EntityID:     after.ID,
		UserID:       uuid.Nil,
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
	EntityID uuid.UUID    `json:"entityID"`
	UserID   uuid.UUID    `json:"userID"`
	Entity   LabelCatalog `json:"entity"`
}

// Marshal returns the event parameters encoded as JSON.
func (p *ActionDeletedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// ActionDeletedData constructs delegate data for label deletion events.
func ActionDeletedData(lc LabelCatalog) delegate.Data {
	params := ActionDeletedParms{
		EntityID: lc.ID,
		UserID:   uuid.Nil,
		Entity:   lc,
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
