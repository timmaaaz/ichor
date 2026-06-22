package workflow

import (
	"encoding/json"
	"errors"

	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/app/sdk/errs"
)

// ValidateActionExecutable rejects a rule action that can never execute: one
// with no template AND no inline "action_type" in its config. The executor
// resolves an action's type from its template (action_templates.action_type)
// or, failing that, from an inline action_type in action_config (see
// edgedb.toActionNode). An action with neither saves cleanly but then fails
// every Temporal dispatch with "action_type is required" — and the only
// evidence is in worker logs. Reject it at write time instead.
//
// This is the write-time chokepoint (CreateRuleAction/UpdateRuleAction); the
// ruleapi.create handler also calls it up front to reject bad embedded actions
// before the parent rule is persisted (create() is not transactional).
func ValidateActionExecutable(templateID *uuid.UUID, config json.RawMessage) error {
	if templateID != nil {
		return nil
	}
	if ConfigActionType(config) != "" {
		return nil
	}
	return errs.NewFieldsError("template_id",
		errors.New(`action requires a template_id or an "action_type" (or legacy "type") field in action_config`))
}

// ConfigActionType extracts an inline action type from an action_config JSON
// document, preferring "action_type" and falling back to the legacy "type"
// key. Returns "" when the document is absent, malformed, or carries no
// (non-empty) type.
//
// This is the single canonical resolver: the write-time guard
// (ValidateActionExecutable), the Temporal executor (edgedb.toActionNode), and
// the API read path (ruleapi.resolveActionType) all call it, so an action's
// type resolves identically at creation, dispatch, and read.
func ConfigActionType(config json.RawMessage) string {
	var cfg struct {
		ActionType string `json:"action_type"`
		Type       string `json:"type"`
	}
	if err := json.Unmarshal(config, &cfg); err != nil {
		return ""
	}
	if cfg.ActionType != "" {
		return cfg.ActionType
	}
	return cfg.Type
}
