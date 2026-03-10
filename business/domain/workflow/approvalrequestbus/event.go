package approvalrequestbus

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
)

// DomainName represents the name of this domain for delegate events.
const DomainName = "approvalrequest"

// Delegate action constants.
const (
	ActionUpdated = "updated"
)

// =============================================================================
// Updated Event
// =============================================================================

// ActionUpdatedParms represents the parameters for the updated action.
type ActionUpdatedParms struct {
	EntityID     uuid.UUID      `json:"entityID"`
	Entity       ApprovalRequest `json:"entity"`
	BeforeEntity ApprovalRequest `json:"beforeEntity,omitempty"`
}

// Marshal returns the event parameters encoded as JSON.
func (p *ActionUpdatedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// ActionUpdatedData constructs delegate data for approval request update events.
func ActionUpdatedData(before, after ApprovalRequest) delegate.Data {
	params := ActionUpdatedParms{
		EntityID:     after.ID,
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
