package inspectionbus

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
)

// DomainName represents the name of this domain for delegate events.
const DomainName = "inspection"

// EntityName is the workflow entity name used for event matching.
// This should match the entity name in workflow.entities table.
// The entity is stored as just the table name (not schema-qualified).
const EntityName = "inspections"

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
// Note: UserID is taken from the entity's InspectorID field which tracks who performed the inspection.
type ActionCreatedParms struct {
	EntityID uuid.UUID  `json:"entityID"`
	UserID   uuid.UUID  `json:"userID"`
	Entity   Inspection `json:"entity"`
}

// Marshal returns the event parameters encoded as JSON.
func (p *ActionCreatedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// ActionCreatedData constructs delegate data for inspection creation events.
func ActionCreatedData(inspection Inspection) delegate.Data {
	params := ActionCreatedParms{
		EntityID: inspection.InspectionID,
		UserID:   inspection.InspectorID, // InspectorID tracks who performed the inspection
		Entity:   inspection,
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
	Entity       Inspection `json:"entity"`
	BeforeEntity Inspection `json:"beforeEntity,omitempty"`
}

// Marshal returns the event parameters encoded as JSON.
func (p *ActionUpdatedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// ActionUpdatedData constructs delegate data for inspection update events.
func ActionUpdatedData(before, after Inspection) delegate.Data {
	params := ActionUpdatedParms{
		EntityID:     after.InspectionID,
		UserID:       after.InspectorID, // InspectorID tracks who performed the inspection
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
	Entity   Inspection `json:"entity"`
}

// Marshal returns the event parameters encoded as JSON.
func (p *ActionDeletedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// ActionDeletedData constructs delegate data for inspection deletion events.
func ActionDeletedData(inspection Inspection) delegate.Data {
	params := ActionDeletedParms{
		EntityID: inspection.InspectionID,
		UserID:   inspection.InspectorID, // InspectorID tracks who performed the inspection
		Entity:   inspection,
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
