package cyclecountitembus

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
)

// DomainName represents the name of this domain for delegate events.
const DomainName = "cyclecountitem"

// EntityName is the workflow entity name used for event matching.
const EntityName = "cycle_count_items"

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
	EntityID uuid.UUID      `json:"entityID"`
	UserID   uuid.UUID      `json:"userID"`
	Entity   CycleCountItem `json:"entity"`
}

func (p *ActionCreatedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

func ActionCreatedData(cci CycleCountItem) delegate.Data {
	params := ActionCreatedParms{
		EntityID: cci.ID,
		UserID:   uuid.UUID{},
		Entity:   cci,
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
	EntityID     uuid.UUID      `json:"entityID"`
	UserID       uuid.UUID      `json:"userID"`
	Entity       CycleCountItem `json:"entity"`
	BeforeEntity CycleCountItem `json:"beforeEntity,omitempty"`
}

func (p *ActionUpdatedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

func ActionUpdatedData(before, after CycleCountItem) delegate.Data {
	params := ActionUpdatedParms{
		EntityID:     after.ID,
		UserID:       uuid.UUID{},
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
	EntityID uuid.UUID      `json:"entityID"`
	UserID   uuid.UUID      `json:"userID"`
	Entity   CycleCountItem `json:"entity"`
}

func (p *ActionDeletedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

func ActionDeletedData(cci CycleCountItem) delegate.Data {
	params := ActionDeletedParms{
		EntityID: cci.ID,
		UserID:   uuid.UUID{},
		Entity:   cci,
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
