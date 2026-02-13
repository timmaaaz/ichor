// Package agenttools defines the tool catalog and executor that lets
// an LLM agent interact with the Ichor REST API.
package agenttools

import (
	"encoding/json"

	"github.com/timmaaaz/ichor/business/sdk/llm"
)

// schema is a convenience for building JSON Schema objects inline.
func schema(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

// ToolDefinitions returns the fixed set of tools exposed to the LLM.
func ToolDefinitions() []llm.ToolDef {
	return []llm.ToolDef{
		// =================================================================
		// Discovery (read-only, zero-arg)
		// =================================================================
		{
			Name:        "discover_action_types",
			Description: "List every available workflow action type with its config schema and output ports.",
			InputSchema: schema(map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			}),
		},
		{
			Name:        "discover_trigger_types",
			Description: "List available trigger types for automation rules (on_create, on_update, on_delete, manual).",
			InputSchema: schema(map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			}),
		},
		{
			Name:        "discover_entities",
			Description: "List entity types that can be used as workflow triggers (schema.table pairs).",
			InputSchema: schema(map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			}),
		},

		// =================================================================
		// Workflow read
		// =================================================================
		{
			Name:        "get_workflow_rule",
			Description: "Fetch a single automation rule by ID. Returns a compact summary: rule metadata (name, description, trigger, entity), node/edge/branch counts, action types used, and a human-readable flow outline showing the execution path. Does NOT return raw action configs — use explain_workflow_node for detail on a specific action.",
			InputSchema: schema(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"rule_id": map[string]any{
						"type":        "string",
						"format":      "uuid",
						"description": "Full UUID of the rule (36 characters with hyphens, e.g. 35da6628-a96b-4bc4-a90f-8fa874ae48cc).",
						"pattern":     "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$",
					},
				},
				"required": []string{"rule_id"},
			}),
		},
		{
			Name:        "explain_workflow_node",
			Description: "Get details about a specific action in a workflow by name or UUID. Returns the action's config, incoming edges (what feeds into it), outgoing edges (where it goes with output port labels), and depth from the start node. Use after get_workflow_rule to drill into a specific action.",
			InputSchema: schema(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"rule_id": map[string]any{
						"type":        "string",
						"format":      "uuid",
						"description": "Full UUID of the rule (36 characters with hyphens).",
						"pattern":     "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$",
					},
					"identifier": map[string]any{
						"type":        "string",
						"description": "Action name or UUID to look up within the workflow.",
					},
				},
				"required": []string{"rule_id", "identifier"},
			}),
		},
		{
			Name:        "list_workflow_rules",
			Description: "List all automation rules. Returns rule metadata (id, name, entity, trigger type, active status).",
			InputSchema: schema(map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			}),
		},

		// =================================================================
		// Workflow write
		// =================================================================
		{
			Name:        "create_workflow",
			Description: "Create a new automation rule with its actions and edges. The rule, actions, and edges are saved in a single transaction.",
			InputSchema: schema(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"workflow": map[string]any{
						"type":        "object",
						"description": "The full workflow payload matching POST /v1/workflow/rules/full.",
					},
				},
				"required": []string{"workflow"},
			}),
		},
		{
			Name:        "update_workflow",
			Description: "Update an existing automation rule (actions, edges, metadata). Uses PUT /v1/workflow/rules/{id}/full.",
			InputSchema: schema(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"rule_id": map[string]any{
						"type":        "string",
						"format":      "uuid",
						"description": "Full UUID of the rule (36 characters with hyphens, e.g. 35da6628-a96b-4bc4-a90f-8fa874ae48cc).",
						"pattern":     "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$",
					},
					"workflow": map[string]any{
						"type":        "object",
						"description": "The full workflow payload matching PUT /v1/workflow/rules/{id}/full.",
					},
				},
				"required": []string{"rule_id", "workflow"},
			}),
		},
		{
			Name:        "validate_workflow",
			Description: "Dry-run validate a workflow without saving. Returns validation errors or success. Uses POST /v1/workflow/rules/full?dry_run=true.",
			InputSchema: schema(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"workflow": map[string]any{
						"type":        "object",
						"description": "The full workflow payload to validate.",
					},
				},
				"required": []string{"workflow"},
			}),
		},

		// =================================================================
		// Alerts
		// =================================================================
		{
			Name:        "list_my_alerts",
			Description: "List active alerts for the current user. Returns alerts with enriched recipient data (names, emails, role names instead of UUIDs). Supports filtering by status and severity.",
			InputSchema: schema(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"status": map[string]any{
						"type":        "string",
						"description": "Filter by status: active, acknowledged, dismissed, resolved. Default: active.",
					},
					"severity": map[string]any{
						"type":        "string",
						"description": "Filter by severity (comma-separated): low, medium, high, critical.",
					},
					"page": map[string]any{
						"type":        "string",
						"description": "Page number (default: 1).",
					},
					"rows": map[string]any{
						"type":        "string",
						"description": "Results per page (default: 10).",
					},
				},
			}),
		},
		{
			Name:        "get_alert_detail",
			Description: "Get detailed information about a specific alert including enriched recipient data (user names, emails, role names).",
			InputSchema: schema(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"alert_id": map[string]any{
						"type":        "string",
						"format":      "uuid",
						"description": "Full UUID of the alert (36 characters with hyphens).",
						"pattern":     "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$",
					},
				},
				"required": []string{"alert_id"},
			}),
		},

		// =================================================================
		// Preview (validate + send to user for approval)
		// =================================================================
		{
			Name:        "preview_workflow",
			Description: "Validate a workflow and send a visual preview to the user for approval. ALWAYS use this instead of create_workflow or update_workflow. The user will see the proposed changes and can accept or reject them directly.",
			InputSchema: schema(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"rule_id": map[string]any{
						"type":        "string",
						"description": "UUID of the rule to update. Omit when creating a new workflow.",
					},
					"workflow": map[string]any{
						"type":        "object",
						"description": "The full workflow payload matching POST /v1/workflow/rules/full.",
					},
					"description": map[string]any{
						"type":        "string",
						"description": "Brief human-readable description of what changes are being made.",
					},
				},
				"required": []string{"workflow", "description"},
			}),
		},
	}
}
