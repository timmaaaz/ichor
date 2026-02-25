package lotlocationbus

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
)

// DomainName represents the name of this domain for delegate events.
const DomainName = "lotlocation"

// EntityName is the workflow entity name used for event matching.
const EntityName = "lot_locations"

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
	EntityID uuid.UUID   `json:"entityID"`
	UserID   uuid.UUID   `json:"userID"`
	Entity   LotLocation `json:"entity"`
}

func (p *ActionCreatedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

func ActionCreatedData(ll LotLocation) delegate.Data {
	params := ActionCreatedParms{
		EntityID: ll.ID,
		UserID:   uuid.Nil,
		Entity:   ll,
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
	EntityID     uuid.UUID   `json:"entityID"`
	UserID       uuid.UUID   `json:"userID"`
	Entity       LotLocation `json:"entity"`
	BeforeEntity LotLocation `json:"beforeEntity,omitempty"`
}

func (p *ActionUpdatedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

func ActionUpdatedData(before, after LotLocation) delegate.Data {
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

type ActionDeletedParms struct {
	EntityID uuid.UUID   `json:"entityID"`
	UserID   uuid.UUID   `json:"userID"`
	Entity   LotLocation `json:"entity"`
}

func (p *ActionDeletedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

func ActionDeletedData(ll LotLocation) delegate.Data {
	params := ActionDeletedParms{
		EntityID: ll.ID,
		UserID:   uuid.Nil,
		Entity:   ll,
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
