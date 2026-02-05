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
