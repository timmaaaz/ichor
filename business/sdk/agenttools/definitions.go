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
			"description": "UUID of the target entity (e.g. '550e8400-e29b-41d4-a716-446655440000'). Use this OR 'entity' (name), not both.",
		},
		"entity": map[string]any{
			"type":        "string",
			"description": "Entity name in 'schema.table' format (e.g. 'inventory.inventory_items'). Resolved to entity_id automatically. Use this OR 'entity_id' (UUID), not both.",
		},
		"trigger_type_id": map[string]any{
			"type":        "string",
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
			"type":        "array",
			"description": "Nodes in the workflow graph. Each action is a step that executes when reached via an edge.",
			"items": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"id": map[string]any{
						"type":        "string",
						"description": "UUID for existing actions (when updating). Omit for new actions.",
					},
					"name": map[string]any{
						"type":        "string",
						"description": "Action name (1-255 chars). Used as the node label in the graph.",
					},
					"description": map[string]any{
						"type": "string",
					},
					"action_type": map[string]any{
						"type":        "string",
						"enum":        actionTypeEnum,
						"description": "Determines the action's behavior and available output ports (e.g. evaluate_condition has 'output-true'/'output-false').",
					},
					"action_config": map[string]any{
						"type":        "object",
						"description": "Config matching the action type's schema (see discover_action_types for schemas and output ports).",
					},
					"is_active": map[string]any{
						"type": "boolean",
					},
					"after": map[string]any{
						"type":        "string",
						"description": "Shorthand to declare the predecessor edge. Format: 'ActionName:port' (e.g. 'Check Stock:output-true') or 'ActionName' (uses default port). Omit for the first action — it becomes the start node. When actions use 'after', the edges array can be omitted.",
					},
				},
				"required": []string{"name", "action_type", "action_config", "is_active"},
			},
		},
		"edges": map[string]any{
			"type":        "array",
			"description": "Directed connections between action nodes. Defines the execution flow of the graph. Can be omitted when actions use the 'after' shorthand.",
			"items": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"source_action_id": map[string]any{
						"type":        "string",
						"description": "ID of the upstream action node. Leave empty for the start edge (the entry point of the graph).",
					},
					"target_action_id": map[string]any{
						"type":        "string",
						"description": "ID of the downstream action node. Use 'temp:N' to reference new actions by their array index (e.g. 'temp:0' for the first action).",
					},
					"edge_type": map[string]any{
						"type":        "string",
						"enum":        edgeTypeEnum,
						"description": "Edge type: 'start' = entry point to the first action, 'sequence' = conditional connection via an output port, 'always' = unconditional connection that always fires.",
					},
					"source_output": map[string]any{
						"type":        "string",
						"description": "The output port on the source action this edge connects from. Creates branching — e.g. 'output-true'/'output-false' for conditions, 'success'/'failure' for operations.",
					},
					"edge_order": map[string]any{
						"type":        "integer",
						"description": "Priority when multiple edges leave the same source port. Lower numbers execute first (0, 1, 2...). Controls branch execution order.",
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

// TableToolDefinitions returns the tools exposed to the LLM for table
// configuration operations.
func TableToolDefinitions() []llm.ToolDef {
	return []llm.ToolDef{
		// =================================================================
		// Discovery (local computation, no API call)
		// =================================================================
		{
			Name:        "discover_table_reference",
			Description: "Get all valid column types, format options, editable types, filter operators, join types, and date format tokens for building table configs. Returns everything needed to construct valid configurations in a single call — no API round-trip required.",
			InputSchema: schema(map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			}),
			ExampleQueries: []string{
				"what column types are available",
				"show me the valid format options for tables",
				"what filter operators can I use",
				"list the editable field types",
			},
		},

		// =================================================================
		// Schema search (shared with workflow)
		// =================================================================
		{
			Name:        "search_database_schema",
			Description: "Browse the database schema. Call with no args to list schemas, with schema to list tables, or with schema+table to get columns and relationships. Use this to verify column names and types before building table configs.",
			InputSchema: schema(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"schema": map[string]any{
						"type":        "string",
						"description": "PostgreSQL schema name (e.g. 'core', 'inventory'). If omitted, lists all schemas.",
					},
					"table": map[string]any{
						"type":        "string",
						"description": "Table name. If provided with schema, returns columns and relationships.",
					},
				},
			}),
			ExampleQueries: []string{
				"what schemas are in the database",
				"show me the columns for the users table",
				"what tables are in the inventory schema",
				"describe the products table structure",
			},
		},

		// =================================================================
		// Table config read
		// =================================================================
		{
			Name: "get_table_config",
			ExampleQueries: []string{
				"show me the users table config",
				"get the configuration for the inventory table",
				"what does the products table config look like",
			},
			Description: "Fetch a single table configuration by UUID or name. Returns the full config including data_source and visual_settings.",
			InputSchema: schema(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"id": map[string]any{
						"type":        "string",
						"description": "UUID of the table config. Use this OR 'name', not both.",
					},
					"name": map[string]any{
						"type":        "string",
						"description": "Name of the table config. Use this OR 'id', not both.",
					},
				},
			}),
		},
		{
			Name:        "list_table_configs",
			Description: "List all table configurations. Returns config metadata (id, name, description) for every table config in the system.",
			InputSchema: schema(map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			}),
			ExampleQueries: []string{
				"show me all table configs",
				"what table configurations exist",
				"list the tables that have been configured",
			},
		},

		// =================================================================
		// Table config validation
		// =================================================================
		{
			Name: "validate_table_config",
			ExampleQueries: []string{
				"check if this table config is valid",
				"validate my table configuration",
				"dry run the table config",
			},
			Description: "Validate a table configuration without saving it. Returns validation errors if the config is invalid. Use this as a dry-run before creating or updating.",
			InputSchema: schema(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"config": map[string]any{
						"type":        "object",
						"description": "The full table config JSON to validate.",
					},
				},
				"required": []string{"config"},
			}),
		},

		// =================================================================
		// Table config write
		// =================================================================
		{
			Name: "create_table_config",
			ExampleQueries: []string{
				"create a new table config",
				"set up a table for the orders data",
				"build a table configuration",
			},
			Description: "Create a new table configuration. Always validate first with validate_table_config, then show the user what will be created before calling this.",
			InputSchema: schema(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{
						"type":        "string",
						"description": "Display name for the table config.",
					},
					"description": map[string]any{
						"type":        "string",
						"description": "Optional description of the table config.",
					},
					"config": map[string]any{
						"type":        "object",
						"description": "The full table config JSON (data_source + visual_settings).",
					},
				},
				"required": []string{"name", "config"},
			}),
		},
		{
			Name: "update_table_config",
			ExampleQueries: []string{
				"update the users table config",
				"change the column settings for this table",
				"modify the table configuration",
			},
			Description: "Update an existing table configuration by UUID. Always fetch the current config first with get_table_config, then validate changes with validate_table_config before updating.",
			InputSchema: schema(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"id": map[string]any{
						"type":        "string",
						"description": "UUID of the table config to update.",
					},
					"config": map[string]any{
						"type":        "object",
						"description": "The updated table config JSON.",
					},
				},
				"required": []string{"id", "config"},
			}),
		},
	}
}

// AllToolDefinitions returns every tool across all contexts (workflow + tables).
// Used by the ToolIndex to build the embedding index at startup.
func AllToolDefinitions() []llm.ToolDef {
	all := make([]llm.ToolDef, 0, 30)
	all = append(all, ToolDefinitions()...)
	all = append(all, TableToolDefinitions()...)
	return all
}

// ToolDefinitions returns the fixed set of tools exposed to the LLM.
func ToolDefinitions() []llm.ToolDef {
	return []llm.ToolDef{
		// =================================================================
		// Discovery (read-only)
		// =================================================================
		{
			Name:        "discover",
			Description: "Discover available workflow building blocks: action types (with config schemas and output ports), trigger types (on_create, on_update, on_delete, manual), or entity types (schema.table pairs that can trigger workflows).",
			InputSchema: schema(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"category": map[string]any{
						"type":        "string",
						"enum":        []string{"action_types", "trigger_types", "entities"},
						"description": "What to discover: action_types (config schemas + output ports), trigger_types (when rules fire), or entities (what rules can trigger on).",
					},
				},
				"required": []string{"category"},
			}),
			ExampleQueries: []string{
				"what action types are available",
				"show me the trigger types",
				"what entities can I use in workflows",
				"what config fields does send_email need",
				"what output ports does evaluate_condition have",
			},
		},

		// =================================================================
		// Workflow read
		// =================================================================
		{
			Name: "get_workflow_rule",
			ExampleQueries: []string{
				"show me the Low Stock Alert workflow",
				"describe the inventory reorder rule",
				"what does this workflow do",
				"explain the automation rule",
				"get the workflow details",
			},
			Description: "Fetch a single automation rule by name or ID. Returns a compact summary: rule metadata (name, description, trigger, entity), node/edge/branch counts, action types used, and a human-readable flow outline showing the execution path. Does NOT return raw action configs — use explain_workflow_node for detail on a specific action.",
			InputSchema: schema(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"workflow_id": map[string]any{
						"type":        "string",
						"description": "Workflow UUID or name. Names are resolved automatically.",
					},
				},
				"required": []string{"workflow_id"},
			}),
		},
		{
			Name: "explain_workflow_node",
			ExampleQueries: []string{
				"who receives alerts from this workflow",
				"what does the condition check",
				"explain the send email action",
				"show me the config for this action",
				"what are the alert recipients",
			},
			Description: "Get details about a specific action in a workflow by name or UUID. Returns the action's full config (including alert recipients, email templates, conditions, etc.), incoming edges, outgoing edges, and depth from the start node. Use this to answer questions about action specifics like 'who receives alerts?' or 'what does this condition check?'. Use after get_workflow_rule to drill into a specific action.",
			InputSchema: schema(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"workflow_id": map[string]any{
						"type":        "string",
						"description": "Workflow UUID or name. Omit to use the current workflow from context.",
					},
					"node_name": map[string]any{
						"type":        "string",
						"description": "The action's name (e.g. 'Alert - Reservation Success') or UUID.",
					},
				},
				"required": []string{"node_name"},
			}),
		},
		{
			Name:        "list_workflow_rules",
			Description: "List all automation rules. Returns up to 500 rules. Returns rule metadata (id, name, entity, trigger type, active status). Response includes total count and has_more flag.",
			InputSchema: schema(map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			}),
			ExampleQueries: []string{
				"show me all workflows",
				"list the automation rules",
				"what workflow rules exist",
				"how many workflows are there",
			},
		},

		// =================================================================
		// Alerts
		// =================================================================
		{
			Name: "list_my_alerts",
			ExampleQueries: []string{
				"show me my alerts",
				"do I have any notifications",
				"check my inbox",
				"what alerts have I received",
			},
			Description: "List alerts in YOUR inbox (alerts where you are a recipient, either directly or via a role). This does NOT search all alerts in the system — only ones addressed to you. Returns up to 50 alerts per request (default). Response includes total count — if has_more is true, increment page to load more. Returns enriched recipient data (names, emails, role names). To see who a workflow is configured to alert, use explain_workflow_node instead.",
			InputSchema: schema(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"status": map[string]any{
						"type":        "string",
						"description": "Filter by status (defaults to 'active' if omitted).",
						"enum":        []string{"active", "acknowledged", "dismissed", "resolved"},
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
						"description": "Results per page (default: 50).",
					},
				},
			}),
		},
		{
			Name: "get_alert_detail",
			ExampleQueries: []string{
				"show me the details of that alert",
				"get info on this specific alert",
				"what happened with alert X",
			},
			Description: "Get detailed information about a specific fired alert by its UUID, including enriched recipient data (user names, emails, role names). Use this for alerts you already know the ID of. To see the configured recipients of a workflow's alert action, use explain_workflow_node instead.",
			InputSchema: schema(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"alert_id": map[string]any{
						"type":        "string",
						"description": "UUID of the alert (e.g. '550e8400-e29b-41d4-a716-446655440000').",
					},
				},
				"required": []string{"alert_id"},
			}),
		},
		{
			Name: "list_alerts_for_rule",
			ExampleQueries: []string{
				"has this workflow triggered any alerts",
				"show alerts fired by the Low Stock rule",
				"what alerts has this rule created",
				"check if the workflow is actually firing",
			},
			Description: "List alerts that were fired by a specific workflow rule. Shows all alerts created by the rule (not just yours). Returns up to 50 alerts per request (default). Response includes total count — if has_more is true, increment page to load more. Returns enriched recipient data. Use this to check if a workflow has actually triggered alerts and who received them.",
			InputSchema: schema(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"workflow_id": map[string]any{
						"type":        "string",
						"description": "Workflow UUID or name. Names are resolved automatically.",
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
						"description": "Results per page (default: 50).",
					},
				},
				"required": []string{"workflow_id"},
			}),
		},

		// =================================================================
		// Preview (validate + send to user for approval)
		// =================================================================
		{
			Name: "preview_workflow",
			ExampleQueries: []string{
				"preview the changes I made",
				"show me what the workflow will look like",
				"validate and preview this workflow",
				"update the workflow and let me review it",
				"change the condition and preview",
			},
			Description: "Validate a workflow and send a visual preview to the user for approval. ALWAYS use this instead of create_workflow or update_workflow. The user will see the proposed changes and can accept or reject them directly.",
			InputSchema: schema(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"workflow_id": map[string]any{
						"type":        "string",
						"description": "UUID of the workflow to update. Omit when creating a new workflow.",
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
			Name: "start_draft",
			ExampleQueries: []string{
				"create a new workflow",
				"build an automation rule",
				"set up a new workflow for inventory",
				"I want to make a new rule",
			},
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
			Name: "add_draft_action",
			ExampleQueries: []string{
				"add a step to the workflow",
				"add a send email action",
				"add a condition check after the lookup",
				"insert an alert action",
			},
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
						"type":        "boolean",
						"description": "Whether the action is active (defaults to true if omitted).",
					},
				},
				"required": []string{"draft_id", "name", "action_type", "action_config"},
			}),
		},
		{
			Name: "remove_draft_action",
			ExampleQueries: []string{
				"remove the email step",
				"delete that action from the draft",
				"take out the condition check",
			},
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
			Name: "preview_draft",
			ExampleQueries: []string{
				"preview the draft workflow",
				"show me what the draft looks like",
				"validate and preview the draft",
				"I'm done adding actions, preview it",
			},
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
