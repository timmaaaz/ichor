// Package workflow provides workflow automation capabilities.
package workflow

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
)

// DomainName represents the name of this domain for delegate events.
const DomainName = "workflow_rule"

// Delegate action constants for rule lifecycle events.
const (
	ActionRuleCreated     = "rule_created"
	ActionRuleUpdated     = "rule_updated"
	ActionRuleDeleted     = "rule_deleted"
	ActionRuleActivated   = "rule_activated"
	ActionRuleDeactivated = "rule_deactivated"
)

// Standard action names matching what domain event.go files use.
// Used by DelegateHandler implementations to register for CRUD events.
const (
	ActionCreated = "created"
	ActionUpdated = "updated"
	ActionDeleted = "deleted"
)

// DelegateEventParams is the standard structure for delegate event parameters.
// Domain event.go files use this structure (or compatible layouts) for their
// ActionXxxParms types. The UserID field identifies who triggered the action.
type DelegateEventParams struct {
	EntityID     uuid.UUID `json:"entityID"`
	UserID       uuid.UUID `json:"userID"`
	Entity       any       `json:"entity,omitempty"`
	BeforeEntity any       `json:"beforeEntity,omitempty"` // Pre-update entity state (for on_update events only)
}

// ActionRuleChangedParms represents the parameters for rule change events.
type ActionRuleChangedParms struct {
	RuleID uuid.UUID `json:"ruleID"`
}

// Marshal returns the event parameters encoded as JSON.
func (p *ActionRuleChangedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// ActionRuleChangedData constructs delegate data for rule change events.
func ActionRuleChangedData(action string, ruleID uuid.UUID) delegate.Data {
	params := ActionRuleChangedParms{
		RuleID: ruleID,
	}
	rawParams, err := params.Marshal()
	if err != nil {
		panic(err)
	}
	return delegate.Data{
		Domain:    DomainName,
		Action:    action,
		RawParams: rawParams,
	}
}

// AllocationResultDomainName is the delegate domain for allocation-result creation
// events (P4 M2). It is distinct from DomainName ("workflow_rule", rule-lifecycle):
// allocation_results is an insert-only entity, so only the "created" action is
// meaningful. The trigger matches on the bare entity name "allocation_results"
// (workflow.entities.name = c.relname), so the entity registered for this domain in
// all.go / the worker MUST be the bare name (P4 §E.4).
const AllocationResultDomainName = "allocation_result"

// ActionAllocationResultCreatedData constructs the delegate event fired after a new
// allocation_results row is written (P4 M2). The Action is the standard "created"
// constant so it registers under on_create; the full row rides Entity for any
// downstream field access.
func ActionAllocationResultCreatedData(ar AllocationResult) delegate.Data {
	params := DelegateEventParams{
		EntityID: ar.ID,
		Entity:   ar,
	}
	rawParams, err := json.Marshal(&params)
	if err != nil {
		panic(err)
	}
	return delegate.Data{
		Domain:    AllocationResultDomainName,
		Action:    ActionCreated,
		RawParams: rawParams,
	}
}

// SyntheticEventData constructs a delegate event for a generic raw-SQL write (P4 M1
// — update_field / create_entity / transition_status). The handler resolves the
// (domain, action) pair from its injected reverse map (P4 §E.2) and supplies the
// synthesized entity snapshot in params; firing under a real domain makes the write
// indistinguishable downstream from a typed bus Create/Update. The lineage visited-set
// rides ctx into delegate.Call, so the P1 loop guard applies for free.
func SyntheticEventData(domain, action string, params DelegateEventParams) delegate.Data {
	rawParams, err := json.Marshal(&params)
	if err != nil {
		panic(err)
	}
	return delegate.Data{
		Domain:    domain,
		Action:    action,
		RawParams: rawParams,
	}
}
