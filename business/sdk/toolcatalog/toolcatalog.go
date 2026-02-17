// Package toolcatalog provides a single source of truth for tool names and
// their group membership. Both the MCP server and the Agent Chat system use
// this catalog to decide which tools are available in a given context.
package toolcatalog

import "slices"

// ToolGroup identifies a logical grouping of tools.
type ToolGroup string

const (
	GroupWorkflow ToolGroup = "workflow"
	GroupTables   ToolGroup = "tables"
)

// Tool name constants — every tool registered in either the MCP server or the
// Agent Chat system should have an entry here.
const (
	// Discovery — workflow
	Discover             = "discover"
	DiscoverActionTypes  = "discover_action_types"
	DiscoverTriggerTypes = "discover_trigger_types"
	DiscoverEntityTypes  = "discover_entity_types"
	DiscoverEntities     = "discover_entities"

	// Discovery — tables
	DiscoverConfigSurfaces  = "discover_config_surfaces"
	DiscoverFieldTypes      = "discover_field_types"
	DiscoverContentTypes    = "discover_content_types"
	DiscoverTableReference  = "discover_table_reference"

	// Workflow read
	GetWorkflow         = "get_workflow"
	GetWorkflowRule     = "get_workflow_rule"
	ExplainWorkflowNode = "explain_workflow_node"
	ExplainWorkflowPath = "explain_workflow_path"
	ListWorkflows       = "list_workflows"
	ListWorkflowRules   = "list_workflow_rules"
	ListActionTemplates = "list_action_templates"

	// Workflow write
	ValidateWorkflow = "validate_workflow"
	CreateWorkflow   = "create_workflow"
	UpdateWorkflow   = "update_workflow"
	PreviewWorkflow  = "preview_workflow"

	// Workflow analysis
	AnalyzeWorkflow  = "analyze_workflow"
	SuggestTemplates = "suggest_templates"
	ShowCascade      = "show_cascade"

	// Draft builder — workflow
	StartDraft        = "start_draft"
	AddDraftAction    = "add_draft_action"
	RemoveDraftAction = "remove_draft_action"
	PreviewDraft      = "preview_draft"

	// Alerts (workflow)
	ListMyAlerts      = "list_my_alerts"
	GetAlertDetail    = "get_alert_detail"
	ListAlertsForRule = "list_alerts_for_rule"

	// Search (shared)
	SearchDatabaseSchema = "search_database_schema"
	SearchEnums          = "search_enums"

	// Tables — read
	GetPageConfig   = "get_page_config"
	GetPageContent  = "get_page_content"
	GetTableConfig  = "get_table_config"
	GetFormDef      = "get_form_definition"
	ListPages       = "list_pages"
	ListForms       = "list_forms"
	ListTableCfgs   = "list_table_configs"

	// Tables — write
	CreatePageConfig  = "create_page_config"
	UpdatePageConfig  = "update_page_config"
	CreatePageContent = "create_page_content"
	UpdatePageContent = "update_page_content"
	CreateForm        = "create_form"
	AddFormField      = "add_form_field"
	CreateTableConfig = "create_table_config"
	UpdateTableConfig = "update_table_config"

	// Tables — validation
	ValidateTableConfig = "validate_table_config"
)

// groupMembers maps each tool to the groups it belongs to. A tool can belong
// to one or both groups. Tools not listed here are implicitly excluded from
// all groups.
var groupMembers = map[string][]ToolGroup{
	// Workflow-only
	Discover:             {GroupWorkflow},
	DiscoverActionTypes:  {GroupWorkflow},
	DiscoverTriggerTypes: {GroupWorkflow},
	DiscoverEntityTypes:  {GroupWorkflow},
	DiscoverEntities:     {GroupWorkflow},
	GetWorkflow:          {GroupWorkflow},
	GetWorkflowRule:      {GroupWorkflow},
	ExplainWorkflowNode:  {GroupWorkflow},
	ExplainWorkflowPath:  {GroupWorkflow},
	ListWorkflows:        {GroupWorkflow},
	ListWorkflowRules:    {GroupWorkflow},
	ListActionTemplates:  {GroupWorkflow},
	ValidateWorkflow:     {GroupWorkflow},
	PreviewWorkflow:      {GroupWorkflow},
	AnalyzeWorkflow:      {GroupWorkflow},
	SuggestTemplates:     {GroupWorkflow},
	ShowCascade:          {GroupWorkflow},
	ListMyAlerts:         {GroupWorkflow},
	GetAlertDetail:       {GroupWorkflow},
	ListAlertsForRule:    {GroupWorkflow},
	StartDraft:           {GroupWorkflow},
	AddDraftAction:       {GroupWorkflow},
	RemoveDraftAction:    {GroupWorkflow},
	PreviewDraft:         {GroupWorkflow},

	// Tables-only
	DiscoverConfigSurfaces:  {GroupTables},
	DiscoverFieldTypes:      {GroupTables},
	DiscoverContentTypes:    {GroupTables},
	DiscoverTableReference:  {GroupTables},
	GetPageConfig:          {GroupTables},
	GetPageContent:         {GroupTables},
	GetTableConfig:         {GroupTables},
	GetFormDef:             {GroupTables},
	ListPages:              {GroupTables},
	ListForms:              {GroupTables},
	ListTableCfgs:          {GroupTables},
	CreatePageConfig:       {GroupTables},
	UpdatePageConfig:       {GroupTables},
	CreatePageContent:      {GroupTables},
	UpdatePageContent:      {GroupTables},
	CreateForm:             {GroupTables},
	AddFormField:           {GroupTables},
	CreateTableConfig:      {GroupTables},
	UpdateTableConfig:      {GroupTables},
	ValidateTableConfig:    {GroupTables},

	// Both groups
	SearchDatabaseSchema: {GroupWorkflow, GroupTables},
	SearchEnums:          {GroupWorkflow, GroupTables},
}

// InGroup reports whether the named tool belongs to the given group.
func InGroup(toolName string, group ToolGroup) bool {
	return slices.Contains(groupMembers[toolName], group)
}

// ToolsForGroup returns the set of tool names that belong to the given group.
func ToolsForGroup(group ToolGroup) []string {
	var out []string
	for name, groups := range groupMembers {
		if slices.Contains(groups, group) {
			out = append(out, name)
		}
	}
	return out
}

// AllTools returns every tool name known to the catalog.
func AllTools() []string {
	out := make([]string, 0, len(groupMembers))
	for name := range groupMembers {
		out = append(out, name)
	}
	return out
}
