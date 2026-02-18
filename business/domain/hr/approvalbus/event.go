package approvalbus

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
)

// DomainName represents the name of this domain for delegate events.
const DomainName = "userapprovalstatus"

// EntityName is the workflow entity name used for event matching.
// This should match the entity name in workflow.entities table.
// The entity is stored as just the table name (not schema-qualified).
const EntityName = "user_approval_statuses"

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
	EntityID uuid.UUID          `json:"entityID"`
	UserID   uuid.UUID          `json:"userID"`
	Entity   UserApprovalStatus `json:"entity"`
}

// Marshal returns the event parameters encoded as JSON.
func (p *ActionCreatedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// ActionCreatedData constructs delegate data for user approval status creation events.
func ActionCreatedData(status UserApprovalStatus) delegate.Data {
	params := ActionCreatedParms{
		EntityID: status.ID,
		UserID:   uuid.Nil, // Reference table - no user tracking
		Entity:   status,
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
	EntityID     uuid.UUID          `json:"entityID"`
	UserID       uuid.UUID          `json:"userID"`
	Entity       UserApprovalStatus `json:"entity"`
	BeforeEntity UserApprovalStatus `json:"beforeEntity,omitempty"`
}

// Marshal returns the event parameters encoded as JSON.
func (p *ActionUpdatedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// ActionUpdatedData constructs delegate data for user approval status update events.
func ActionUpdatedData(before, after UserApprovalStatus) delegate.Data {
	params := ActionUpdatedParms{
		EntityID:     after.ID,
		UserID:       uuid.Nil, // Reference table - no user tracking
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
	EntityID uuid.UUID          `json:"entityID"`
	UserID   uuid.UUID          `json:"userID"`
	Entity   UserApprovalStatus `json:"entity"`
}

// Marshal returns the event parameters encoded as JSON.
func (p *ActionDeletedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// ActionDeletedData constructs delegate data for user approval status deletion events.
func ActionDeletedData(status UserApprovalStatus) delegate.Data {
	params := ActionDeletedParms{
		EntityID: status.ID,
		UserID:   uuid.Nil, // Reference table - no user tracking
		Entity:   status,
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
