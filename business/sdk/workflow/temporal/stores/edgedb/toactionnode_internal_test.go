package edgedb

import (
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// Test_toActionNode_TemplatelessActionType verifies that an action with no
// linked template (template_action_type NULL) resolves its ActionType from an
// inline "action_type" in action_config, instead of dispatching with an empty
// ActionType that fails ActionActivityInput.Validate() at runtime.
func Test_toActionNode_TemplatelessActionType(t *testing.T) {
	t.Run("inline action_type resolves when template is NULL", func(t *testing.T) {
		dba := dbAction{
			ID:                 uuid.New().String(),
			Name:               "send welcome email",
			ActionConfig:       json.RawMessage(`{"action_type":"send_email","to":"x@y.z"}`),
			IsActive:           true,
			TemplateActionType: sql.NullString{Valid: false},
		}

		node := toActionNode(dba)

		if node.ActionType != "send_email" {
			t.Errorf("ActionType = %q, want %q (inline action_type should resolve when no template)", node.ActionType, "send_email")
		}
	})

	t.Run("empty when neither template nor inline action_type present", func(t *testing.T) {
		dba := dbAction{
			ID:                 uuid.New().String(),
			Name:               "broken action",
			ActionConfig:       json.RawMessage(`{"to":"x@y.z"}`),
			IsActive:           true,
			TemplateActionType: sql.NullString{Valid: false},
		}

		node := toActionNode(dba)

		if node.ActionType != "" {
			t.Errorf("ActionType = %q, want empty string (no template, no inline type)", node.ActionType)
		}
	})

	t.Run("legacy inline \"type\" key resolves when template is NULL", func(t *testing.T) {
		dba := dbAction{
			ID:                 uuid.New().String(),
			Name:               "legacy typed action",
			ActionConfig:       json.RawMessage(`{"type":"send_notification"}`),
			IsActive:           true,
			TemplateActionType: sql.NullString{Valid: false},
		}

		node := toActionNode(dba)

		if node.ActionType != "send_notification" {
			t.Errorf("ActionType = %q, want %q (legacy \"type\" key should resolve when no template)", node.ActionType, "send_notification")
		}
	})

	t.Run("action_type wins over legacy type when both present", func(t *testing.T) {
		dba := dbAction{
			ID:                 uuid.New().String(),
			Name:               "both keys",
			ActionConfig:       json.RawMessage(`{"action_type":"send_email","type":"send_notification"}`),
			IsActive:           true,
			TemplateActionType: sql.NullString{Valid: false},
		}

		node := toActionNode(dba)

		if node.ActionType != "send_email" {
			t.Errorf("ActionType = %q, want %q (action_type must win over legacy type)", node.ActionType, "send_email")
		}
	})

	t.Run("template action_type takes precedence over inline", func(t *testing.T) {
		dba := dbAction{
			ID:                 uuid.New().String(),
			Name:               "templated action",
			ActionConfig:       json.RawMessage(`{"action_type":"send_email"}`),
			IsActive:           true,
			TemplateActionType: sql.NullString{String: "create_alert", Valid: true},
		}

		node := toActionNode(dba)

		if node.ActionType != "create_alert" {
			t.Errorf("ActionType = %q, want %q (template must win over inline)", node.ActionType, "create_alert")
		}
	})
}
