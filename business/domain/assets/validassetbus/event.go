package validassetbus

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
)

// DomainName represents the name of this domain for delegate events.
const DomainName = "validasset"

// EntityName is the workflow entity name used for event matching.
// This should match the entity name in workflow.entities table.
// The entity is stored as just the table name (not schema-qualified).
const EntityName = "valid_assets"

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
	EntityID uuid.UUID  `json:"entityID"`
	UserID   uuid.UUID  `json:"userID"`
	Entity   ValidAsset `json:"entity"`
}

// Marshal returns the event parameters encoded as JSON.
func (p *ActionCreatedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// ActionCreatedData constructs delegate data for valid asset creation events.
func ActionCreatedData(validAsset ValidAsset) delegate.Data {
	params := ActionCreatedParms{
		EntityID: validAsset.ID,
		UserID:   validAsset.CreatedBy,
		Entity:   validAsset,
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
	EntityID uuid.UUID  `json:"entityID"`
	UserID   uuid.UUID  `json:"userID"`
	Entity   ValidAsset `json:"entity"`
}

// Marshal returns the event parameters encoded as JSON.
func (p *ActionUpdatedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// ActionUpdatedData constructs delegate data for valid asset update events.
func ActionUpdatedData(validAsset ValidAsset) delegate.Data {
	params := ActionUpdatedParms{
		EntityID: validAsset.ID,
		UserID:   validAsset.UpdatedBy,
		Entity:   validAsset,
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
	Entity   ValidAsset `json:"entity"`
}

// Marshal returns the event parameters encoded as JSON.
func (p *ActionDeletedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// ActionDeletedData constructs delegate data for valid asset deletion events.
// Note: For delete, we use UpdatedBy as the user who performed the delete.
func ActionDeletedData(validAsset ValidAsset) delegate.Data {
	params := ActionDeletedParms{
		EntityID: validAsset.ID,
		UserID:   validAsset.UpdatedBy, // UpdatedBy tracks who performed the delete
		Entity:   validAsset,
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
