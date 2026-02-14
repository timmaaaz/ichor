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

// actionTypeEnum lists every valid action_type value for workflow actions.
// Source of truth: workflowsaveapp.SaveActionRequest validate:"oneof=..."
var actionTypeEnum = []string{
	"allocate_inventory",
	"check_inventory",
	"check_reorder_point",
	"commit_allocation",
	"create_alert",
	"create_entity",
	"delay",
	"evaluate_condition",
	"log_audit_entry",
	"lookup_entity",
	"release_reservation",
	"reserve_inventory",
	"seek_approval",
	"send_email",
	"send_notification",
	"transition_status",
	"update_field",
}

// edgeTypeEnum lists every valid edge_type value for workflow edges.
// Source of truth: workflowsaveapp.SaveEdgeRequest validate:"oneof=..."
var edgeTypeEnum = []string{"start", "sequence", "always"}

// workflowPayloadSchema is the structured JSON Schema for the workflow object
// accepted by create_workflow, update_workflow, validate_workflow, and
// preview_workflow. It mirrors workflowsaveapp.SaveWorkflowRequest.
var workflowPayloadSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"name": map[string]any{
			"type":        "string",
			"description": "Rule name (1-255 chars).",
		},
		"description": map[string]any{
			"type":        "string",
			"description": "Rule description (max 1000 chars).",
		},
		"is_active": map[string]any{
			"type": "boolean",
		},
		"entity_id": map[string]any{
			"type":        "string",
			"format":      "uuid",
			"description": "UUID of the target entity. Use this OR 'entity' (name), not both.",
		},
		"entity": map[string]any{
			"type":        "string",
			"description": "Entity name in 'schema.table' format (e.g. 'inventory.inventory_items'). Resolved to entity_id automatically. Use this OR 'entity_id' (UUID), not both.",
		},
		"trigger_type_id": map[string]any{
			"type":        "string",
			"format":      "uuid",
			"description": "UUID of the trigger type. Use this OR 'trigger_type' (name), not both.",
		},
		"trigger_type": map[string]any{
			"type":        "string",
			"description": "Trigger type name (e.g. 'on_update', 'on_create', 'on_delete', 'manual'). Resolved to trigger_type_id automatically. Use this OR 'trigger_type_id' (UUID), not both.",
		},
		"trigger_conditions": map[string]any{
			"type":        "object",
			"description": "Optional trigger filter conditions.",
		},
		"actions": map[string]any{
			"type": "array",
			"items": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"id": map[string]any{
						"type":        "string",
						"description": "UUID for existing actions, omit for new.",
					},
					"name": map[string]any{
						"type":        "string",
						"description": "Action name (1-255 chars).",
					},
					"description": map[string]any{
						"type": "string",
					},
					"action_type": map[string]any{
						"type":        "string",
						"enum":        actionTypeEnum,
						"description": "The action type.",
					},
					"action_config": map[string]any{
						"type":        "object",
						"description": "Config matching the action type's schema (see discover_action_types).",
					},
					"is_active": map[string]any{
						"type": "boolean",
					},
					"after": map[string]any{
						"type":        "string",
						"description": "Predecessor action. Format: 'ActionName:port' or 'ActionName' (uses default port). Omit for the first action. When actions use 'after', the edges array can be omitted.",
					},
				},
				"required": []string{"name", "action_type", "action_config", "is_active"},
			},
		},
		"edges": map[string]any{
			"type": "array",
			"items": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"source_action_id": map[string]any{
						"type":        "string",
						"description": "Source action ID or empty for start edges.",
					},
					"target_action_id": map[string]any{
						"type":        "string",
						"description": "Target action ID. Use 'temp:N' for new actions by array index.",
					},
					"edge_type": map[string]any{
						"type":        "string",
						"enum":        edgeTypeEnum,
						"description": "Edge type.",
					},
					"source_output": map[string]any{
						"type":        "string",
						"description": "Output port name (e.g. 'success', 'output-true').",
					},
					"edge_order": map[string]any{
						"type":        "integer",
						"minimum":     0,
						"description": "Execution order for edges sharing the same source.",
					},
				},
				"required": []string{"target_action_id", "edge_type"},
			},
		},
		"canvas_layout": map[string]any{
			"type":        "object",
			"description": "Optional canvas layout for the UI.",
		},
	},
	"required": []string{"name", "is_active", "actions"},
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
			Description: "Get details about a specific action in a workflow by name or UUID. Returns the action's full config (including alert recipients, email templates, conditions, etc.), incoming edges, outgoing edges, and depth from the start node. Use this to answer questions about action specifics like 'who receives alerts?' or 'what does this condition check?'. Use after get_workflow_rule to drill into a specific action.",
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
					"workflow": workflowPayloadSchema,
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
					"workflow": workflowPayloadSchema,
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
					"workflow": workflowPayloadSchema,
				},
				"required": []string{"workflow"},
			}),
		},

		// =================================================================
		// Alerts
		// =================================================================
		{
			Name:        "list_my_alerts",
			Description: "List alerts in YOUR inbox (alerts where you are a recipient, either directly or via a role). This does NOT search all alerts in the system — only ones addressed to you. Returns enriched recipient data (names, emails, role names). To see who a workflow is configured to alert, use explain_workflow_node instead.",
			InputSchema: schema(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"status": map[string]any{
						"type":        "string",
						"description": "Filter by status.",
						"enum":        []string{"active", "acknowledged", "dismissed", "resolved"},
						"default":     "active",
					},
					"severity": map[string]any{
						"type":        "string",
						"description": "Filter by severity.",
						"enum":        []string{"low", "medium", "high", "critical"},
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
			Description: "Get detailed information about a specific fired alert by its UUID, including enriched recipient data (user names, emails, role names). Use this for alerts you already know the ID of. To see the configured recipients of a workflow's alert action, use explain_workflow_node instead.",
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
		{
			Name:        "list_alerts_for_rule",
			Description: "List alerts that were fired by a specific workflow rule. Shows all alerts created by the rule (not just yours). Returns enriched recipient data. Use this to check if a workflow has actually triggered alerts and who received them.",
			InputSchema: schema(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"rule_id": map[string]any{
						"type":        "string",
						"format":      "uuid",
						"description": "Full UUID of the workflow rule.",
						"pattern":     "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$",
					},
					"status": map[string]any{
						"type":        "string",
						"description": "Filter by alert status.",
						"enum":        []string{"active", "acknowledged", "dismissed", "resolved"},
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
				"required": []string{"rule_id"},
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
					"workflow": workflowPayloadSchema,
					"description": map[string]any{
						"type":        "string",
						"description": "Brief human-readable description of what changes are being made.",
					},
				},
				"required": []string{"workflow", "description"},
			}),
		},

		// =================================================================
		// Draft builder (incremental workflow creation)
		// =================================================================
		{
			Name:        "start_draft",
			Description: "Start building a new workflow incrementally. Returns a draft_id to use with add_draft_action and preview_draft. Accepts entity/trigger names (no UUID lookup needed).",
			InputSchema: schema(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{
						"type":        "string",
						"description": "Workflow rule name (1-255 chars).",
					},
					"entity": map[string]any{
						"type":        "string",
						"description": "Entity in 'schema.table' format (e.g. 'inventory.inventory_items') or UUID.",
					},
					"trigger_type": map[string]any{
						"type":        "string",
						"description": "Trigger type name (e.g. 'on_update') or UUID.",
					},
					"description": map[string]any{
						"type":        "string",
						"description": "Optional workflow description.",
					},
					"trigger_conditions": map[string]any{
						"type":        "object",
						"description": "Optional trigger filter conditions.",
					},
				},
				"required": []string{"name", "entity", "trigger_type"},
			}),
		},
		{
			Name:        "add_draft_action",
			Description: "Add an action to a draft workflow. Use 'after' to declare which action precedes this one. Omit 'after' for the first action (it becomes the start node). Returns the action's available output ports.",
			InputSchema: schema(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"draft_id": map[string]any{
						"type":        "string",
						"description": "Draft ID from start_draft.",
					},
					"name": map[string]any{
						"type":        "string",
						"description": "Action name (1-255 chars).",
					},
					"action_type": map[string]any{
						"type":        "string",
						"enum":        actionTypeEnum,
						"description": "The action type.",
					},
					"action_config": map[string]any{
						"type":        "object",
						"description": "Config matching the action type's schema (see discover_action_types).",
					},
					"after": map[string]any{
						"type":        "string",
						"description": "Predecessor: 'ActionName:port' or 'ActionName' (uses default port). Omit for first action.",
					},
					"is_active": map[string]any{
						"type":    "boolean",
						"default": true,
					},
				},
				"required": []string{"draft_id", "name", "action_type", "action_config"},
			}),
		},
		{
			Name:        "remove_draft_action",
			Description: "Remove an action from a draft workflow by name. Also removes any 'after' references to this action from other actions.",
			InputSchema: schema(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"draft_id": map[string]any{
						"type":        "string",
						"description": "Draft ID from start_draft.",
					},
					"action_name": map[string]any{
						"type":        "string",
						"description": "Name of the action to remove.",
					},
				},
				"required": []string{"draft_id", "action_name"},
			}),
		},
		{
			Name:        "preview_draft",
			Description: "Assemble the draft into a complete workflow, validate it, and send a visual preview to the user for approval. The user will accept or reject the preview directly.",
			InputSchema: schema(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"draft_id": map[string]any{
						"type":        "string",
						"description": "Draft ID from start_draft.",
					},
					"description": map[string]any{
						"type":        "string",
						"description": "Brief human-readable description of the workflow for user review.",
					},
				},
				"required": []string{"draft_id", "description"},
			}),
		},
	}
}
