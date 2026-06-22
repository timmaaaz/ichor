package ruleapi

import (
	"encoding/json"
	"testing"

	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

// Test_resolveActionType covers the API read path that populates
// ActionResponse.ActionType. It must resolve the type identically to the
// write-time guard and the Temporal executor — in particular it must honor the
// legacy "type" key, which a bespoke "action_type"-only lookup reported as an
// empty type (an action that saved and dispatched fine showed a blank type in
// GET responses).
func Test_resolveActionType(t *testing.T) {
	tests := []struct {
		name string
		view workflow.RuleActionView
		want string
	}{
		{
			name: "template action type takes precedence over inline",
			view: workflow.RuleActionView{
				TemplateActionType: "send_email",
				ActionConfig:       json.RawMessage(`{"action_type":"ignored"}`),
			},
			want: "send_email",
		},
		{
			name: "inline action_type resolves when no template",
			view: workflow.RuleActionView{
				ActionConfig: json.RawMessage(`{"action_type":"create_alert"}`),
			},
			want: "create_alert",
		},
		{
			name: "legacy type key resolves when no template",
			view: workflow.RuleActionView{
				ActionConfig: json.RawMessage(`{"type":"send_notification"}`),
			},
			want: "send_notification",
		},
		{
			name: "action_type wins over legacy type when both present",
			view: workflow.RuleActionView{
				ActionConfig: json.RawMessage(`{"action_type":"create_alert","type":"send_notification"}`),
			},
			want: "create_alert",
		},
		{
			name: "empty when neither template nor inline type",
			view: workflow.RuleActionView{
				ActionConfig: json.RawMessage(`{"to":"x"}`),
			},
			want: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := resolveActionType(tc.view); got != tc.want {
				t.Errorf("resolveActionType() = %q, want %q", got, tc.want)
			}
		})
	}
}
