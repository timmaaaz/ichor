package picktaskbus

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
)

// DomainName represents the name of this domain for delegate events.
const DomainName = "picktask"

// EntityName is the workflow entity name used for event matching.
const EntityName = "pick_tasks"

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
	EntityID uuid.UUID `json:"entityID"`
	UserID   uuid.UUID `json:"userID"`
	Entity   PickTask  `json:"entity"`
}

func (p *ActionCreatedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

func ActionCreatedData(pt PickTask) delegate.Data {
	params := ActionCreatedParms{
		EntityID: pt.ID,
		UserID:   pt.CreatedBy,
		Entity:   pt,
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
	EntityID     uuid.UUID `json:"entityID"`
	UserID       uuid.UUID `json:"userID"`
	Entity       PickTask  `json:"entity"`
	BeforeEntity PickTask  `json:"beforeEntity,omitempty"`
}

func (p *ActionUpdatedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

func ActionUpdatedData(before, after PickTask) delegate.Data {
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
	EntityID uuid.UUID `json:"entityID"`
	UserID   uuid.UUID `json:"userID"`
	Entity   PickTask  `json:"entity"`
}

func (p *ActionDeletedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

func ActionDeletedData(pt PickTask) delegate.Data {
	params := ActionDeletedParms{
		EntityID: pt.ID,
		UserID:   pt.CreatedBy,
		Entity:   pt,
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
