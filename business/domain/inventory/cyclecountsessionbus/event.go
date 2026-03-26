package cyclecountsessionbus

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
)

// DomainName represents the name of this domain for delegate events.
const DomainName = "cyclecountsession"

// EntityName is the workflow entity name used for event matching.
const EntityName = "cycle_count_sessions"

// Delegate action constants.
const (
	ActionCreated = "created"
	ActionUpdated = "updated"
	ActionDeleted = "deleted"
)

// =============================================================================
// Created Event
// =============================================================================

type ActionCreatedParms struct {
	EntityID uuid.UUID         `json:"entityID"`
	UserID   uuid.UUID         `json:"userID"`
	Entity   CycleCountSession `json:"entity"`
}

func (p *ActionCreatedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

func ActionCreatedData(ccs CycleCountSession) delegate.Data {
	params := ActionCreatedParms{
		EntityID: ccs.ID,
		UserID:   ccs.CreatedBy,
		Entity:   ccs,
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

type ActionUpdatedParms struct {
	EntityID     uuid.UUID         `json:"entityID"`
	UserID       uuid.UUID         `json:"userID"`
	Entity       CycleCountSession `json:"entity"`
	BeforeEntity CycleCountSession `json:"beforeEntity,omitempty"`
}

func (p *ActionUpdatedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

func ActionUpdatedData(before, after CycleCountSession) delegate.Data {
	params := ActionUpdatedParms{
		EntityID:     after.ID,
		UserID:       after.CreatedBy,
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

type ActionDeletedParms struct {
	EntityID uuid.UUID         `json:"entityID"`
	UserID   uuid.UUID         `json:"userID"`
	Entity   CycleCountSession `json:"entity"`
}

func (p *ActionDeletedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

func ActionDeletedData(ccs CycleCountSession) delegate.Data {
	params := ActionDeletedParms{
		EntityID: ccs.ID,
		UserID:   ccs.CreatedBy,
		Entity:   ccs,
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
