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
	EntityID uuid.UUID `json:"entityID"`
	UserID   uuid.UUID `json:"userID"`
	Entity   any       `json:"entity,omitempty"`
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
