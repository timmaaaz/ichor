// Package catalogapi provides a single discovery endpoint that tells agents
// what configuration surfaces are available in the system, with links to
// CRUD endpoints, discovery endpoints, and constraint summaries.
package catalogapi

import (
	"context"
	"net/http"

	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct{}

func newAPI() *api {
	return &api{}
}

// queryCatalog handles GET /v1/agent/catalog
func (a *api) queryCatalog(_ context.Context, _ *http.Request) web.Encoder {
	return Catalog(catalog)
}

// catalog is the static list of all configurable surfaces in the system.
var catalog = []ConfigSurface{
	// =========================================================================
	// UI Configuration
	// =========================================================================
	{
		Name:        "Page Configs",
		Description: "Page-level configuration containers that hold content blocks, actions, and layout settings. Each page config defines a named page in the application.",
		Category:    "ui",
		Endpoints: Endpoints{
			List:   "GET /v1/config/page-configs/all",
			Get:    "GET /v1/config/page-configs/id/{config_id}",
			Create: "POST /v1/config/page-configs",
			Update: "PUT /v1/config/page-configs/id/{config_id}",
			Delete: "DELETE /v1/config/page-configs/id/{config_id}",
		},
		DiscoveryURL: "GET /v1/config/page-configs/all",
		Constraints:  []string{"name must be unique", "supports import/export via POST /v1/config/page-configs/export and /import"},
	},
	{
		Name:        "Page Content",
		Description: "Content blocks within a page config. Each block has a content type (table, form, chart, tabs, container, text) and a layout configuration.",
		Category:    "ui",
		Endpoints: Endpoints{
			List:   "GET /v1/config/page-configs/content/{page_config_id}",
			Get:    "GET /v1/config/page-content/{content_id}",
			Create: "POST /v1/config/page-content",
			Update: "PUT /v1/config/page-content/{content_id}",
			Delete: "DELETE /v1/config/page-content/{content_id}",
		},
		DiscoveryURL: "GET /v1/config/schemas/content-types",
		Constraints: []string{
			"content_type must be one of: table, form, chart, tabs, container, text",
			"table and chart types require a table_config_id reference",
			"form type requires a form_id reference",
			"tabs and container types support nested children",
			"layout JSONB schema: GET /v1/config/schemas/layout",
		},
	},
	{
		Name:        "Page Actions",
		Description: "Action buttons, dropdowns, and separators attached to a page config. Supports buttons, dropdown menus with items, and visual separators.",
		Category:    "ui",
		Endpoints: Endpoints{
			List:   "GET /v1/config/page-configs/actions/{page_config_id}",
			Get:    "GET /v1/config/page-actions/{action_id}",
			Create: "POST /v1/config/page-actions/buttons",
			Update: "PUT /v1/config/page-actions/buttons/{action_id}",
			Delete: "DELETE /v1/config/page-actions/{action_id}",
		},
		DiscoveryURL: "GET /v1/config/schemas/page-action-types",
		Constraints: []string{
			"three action kinds: buttons (POST /buttons), dropdowns (POST /dropdowns), separators (POST /separators)",
			"update uses kind-specific routes: PUT /buttons/{id}, PUT /dropdowns/{id}, PUT /separators/{id}",
			"batch create via POST /v1/config/page-configs/actions/batch/{page_config_id}",
		},
	},
	{
		Name:        "Table Configs",
		Description: "Table/widget configurations stored as JSONB. Defines data sources, column selections, joins, filters, sorting, visual settings, and chart aggregation.",
		Category:    "ui",
		Endpoints: Endpoints{
			List:   "GET /v1/data/configs/all",
			Get:    "GET /v1/data/id/{table_config_id}",
			Create: "POST /v1/data",
			Update: "PUT /v1/data/{table_config_id}",
			Delete: "DELETE /v1/data/{table_config_id}",
		},
		DiscoveryURL: "GET /v1/config/schemas/table-config",
		Constraints: []string{
			"config JSONB schema: GET /v1/config/schemas/table-config",
			"validate config without saving: POST /v1/data/validate",
			"execute query: POST /v1/data/execute/{table_config_id}",
			"chart query: POST /v1/data/chart/{table_config_id}",
			"supports import/export via POST /v1/data/export and /import",
			"all columns in visual_settings must have an explicit type (string, number, datetime, boolean, uuid, status, computed, lookup)",
		},
	},
	{
		Name:        "Forms",
		Description: "Form definitions for data entry. Each form targets an entity and contains ordered form fields with field-type-specific configuration.",
		Category:    "ui",
		Endpoints: Endpoints{
			List:   "GET /v1/config/forms",
			Get:    "GET /v1/config/forms/{form_id}/full",
			Create: "POST /v1/config/forms",
			Update: "PUT /v1/config/forms/{form_id}",
			Delete: "DELETE /v1/config/forms/{form_id}",
		},
		DiscoveryURL: "GET /v1/config/form-field-types",
		Constraints: []string{
			"get form with fields: GET /v1/config/forms/{form_id}/full",
			"get form by name: GET /v1/config/forms/name/{form_name}/full",
			"field type schemas: GET /v1/config/form-field-types/{type}/schema",
			"supports import/export via POST /v1/config/forms/export and /import",
		},
	},
	{
		Name:        "Form Fields",
		Description: "Individual fields within a form. Each field has a type (text, number, dropdown, etc.) with type-specific configuration stored as JSONB.",
		Category:    "ui",
		Endpoints: Endpoints{
			List:   "GET /v1/config/forms/{form_id}/form-fields",
			Get:    "GET /v1/config/form-fields/{field_id}",
			Create: "POST /v1/config/form-fields",
			Update: "PUT /v1/config/form-fields/{field_id}",
			Delete: "DELETE /v1/config/form-fields/{field_id}",
		},
		DiscoveryURL: "GET /v1/config/form-field-types",
		Constraints: []string{
			"field_type must match a registered type (see GET /v1/config/form-field-types)",
			"config JSONB must conform to the field type schema (see GET /v1/config/form-field-types/{type}/schema)",
			"order_index controls display order within the form",
		},
	},

	// =========================================================================
	// Workflow Configuration
	// =========================================================================
	{
		Name:        "Workflow Rules",
		Description: "Automation rules that define trigger conditions and action graphs. Each rule has a trigger type, entity type, and a directed acyclic graph of actions connected by edges.",
		Category:    "workflow",
		Endpoints: Endpoints{
			List:   "GET /v1/workflow/rules",
			Get:    "GET /v1/workflow/rules/{id}",
			Create: "POST /v1/workflow/rules/full",
			Update: "PUT /v1/workflow/rules/{id}/full",
			Delete: "DELETE /v1/workflow/rules/{id}",
		},
		DiscoveryURL: "GET /v1/workflow/action-types",
		Constraints: []string{
			"full save (create/update with actions + edges): POST/PUT /v1/workflow/rules/full",
			"dry-run validation: POST /v1/workflow/rules/full?dry_run=true",
			"action graph must be a valid DAG with exactly one start edge",
			"all actions must have a valid action_type (see GET /v1/workflow/action-types)",
			"action config must match the type schema (see GET /v1/workflow/action-types/{type}/schema)",
			"edges must reference valid output ports on the source action",
			"trigger types: GET /v1/workflow/trigger-types",
			"entity types: GET /v1/workflow/entity-types",
			"available entities: GET /v1/workflow/entities",
		},
	},
	{
		Name:        "Action Templates",
		Description: "Reusable action configuration templates that can be referenced by rule actions. Templates define a base action type and default config.",
		Category:    "workflow",
		Endpoints: Endpoints{
			List: "GET /v1/workflow/templates",
		},
		DiscoveryURL: "GET /v1/workflow/templates/active",
		Constraints: []string{
			"templates are read-only through reference endpoints",
			"active templates: GET /v1/workflow/templates/active",
		},
	},
	{
		Name:        "Alerts",
		Description: "Workflow-generated alerts delivered to users. Supports acknowledgement and dismissal. Created by the create_alert action type.",
		Category:    "workflow",
		Endpoints: Endpoints{
			List: "GET /v1/workflow/alerts",
			Get:  "GET /v1/workflow/alerts/{id}",
		},
		Constraints: []string{
			"user alerts: GET /v1/workflow/alerts/mine",
			"acknowledge: POST /v1/workflow/alerts/{id}/acknowledge",
			"dismiss: POST /v1/workflow/alerts/{id}/dismiss",
			"batch acknowledge: POST /v1/workflow/alerts/acknowledge-selected",
			"batch dismiss: POST /v1/workflow/alerts/dismiss-selected",
		},
	},

	// =========================================================================
	// System Configuration
	// =========================================================================
	{
		Name:        "Action Permissions",
		Description: "Controls which roles can manually execute workflow actions. Managed through the action API.",
		Category:    "system",
		Endpoints: Endpoints{
			List: "GET /v1/workflow/actions",
		},
		Constraints: []string{
			"execute action: POST /v1/workflow/actions/{actionType}/execute",
			"permissions are checked per-role in the app layer",
		},
	},
	{
		Name:        "Enum Labels",
		Description: "Human-friendly display labels for PostgreSQL ENUM values. Used to show readable names in dropdowns and status badges instead of raw enum values.",
		Category:    "system",
		Endpoints: Endpoints{
			List: "GET /v1/config/enums/{schema}/{name}/options",
		},
		DiscoveryURL: "GET /v1/introspection/enums/{schema}",
		Constraints: []string{
			"list enum types in a schema: GET /v1/introspection/enums/{schema}",
			"get raw enum values: GET /v1/introspection/enums/{schema}/{name}",
			"get merged options with labels: GET /v1/config/enums/{schema}/{name}/options",
		},
	},
	{
		Name:        "Database Introspection",
		Description: "Read-only introspection of the PostgreSQL schema. Discover tables, columns, relationships, foreign keys, and enum types across all schemas.",
		Category:    "system",
		Endpoints: Endpoints{
			List: "GET /v1/introspection/schemas",
		},
		DiscoveryURL: "GET /v1/introspection/schemas",
		Constraints: []string{
			"tables in schema: GET /v1/introspection/schemas/{schema}/tables",
			"columns: GET /v1/introspection/tables/{schema}/{table}/columns",
			"relationships: GET /v1/introspection/tables/{schema}/{table}/relationships",
			"referencing tables: GET /v1/introspection/tables/{schema}/{table}/referencing-tables",
			"read-only (no mutations)",
		},
	},
}
