package toolcatalog

import (
	"sort"
	"testing"
)

func TestInGroup_WorkflowTools(t *testing.T) {
	workflowOnly := []string{
		DiscoverActionTypes, DiscoverTriggerTypes, DiscoverEntityTypes, DiscoverEntities,
		GetWorkflow, GetWorkflowRule, ExplainWorkflowNode, ExplainWorkflowPath,
		ListWorkflows, ListWorkflowRules, ListActionTemplates,
		ValidateWorkflow, CreateWorkflow, UpdateWorkflow, PreviewWorkflow,
		AnalyzeWorkflow, SuggestTemplates, ShowCascade,
		ListMyAlerts, GetAlertDetail,
	}
	for _, name := range workflowOnly {
		if !InGroup(name, GroupWorkflow) {
			t.Errorf("expected %q to be in GroupWorkflow", name)
		}
		if InGroup(name, GroupTables) {
			t.Errorf("expected %q NOT to be in GroupTables", name)
		}
	}
}

func TestInGroup_TablesTools(t *testing.T) {
	tablesOnly := []string{
		DiscoverConfigSurfaces, DiscoverFieldTypes, DiscoverContentTypes,
		GetPageConfig, GetPageContent, GetTableConfig, GetFormDef,
		ListPages, ListForms, ListTableCfgs,
		CreatePageConfig, UpdatePageConfig, CreatePageContent, UpdatePageContent,
		CreateForm, AddFormField, CreateTableConfig, UpdateTableConfig,
		ValidateTableConfig,
	}
	for _, name := range tablesOnly {
		if !InGroup(name, GroupTables) {
			t.Errorf("expected %q to be in GroupTables", name)
		}
		if InGroup(name, GroupWorkflow) {
			t.Errorf("expected %q NOT to be in GroupWorkflow", name)
		}
	}
}

func TestInGroup_SharedTools(t *testing.T) {
	shared := []string{SearchDatabaseSchema, SearchEnums}
	for _, name := range shared {
		if !InGroup(name, GroupWorkflow) {
			t.Errorf("expected %q to be in GroupWorkflow", name)
		}
		if !InGroup(name, GroupTables) {
			t.Errorf("expected %q to be in GroupTables", name)
		}
	}
}

func TestInGroup_UnknownTool(t *testing.T) {
	if InGroup("nonexistent_tool", GroupWorkflow) {
		t.Error("unknown tool should not be in any group")
	}
	if InGroup("nonexistent_tool", GroupTables) {
		t.Error("unknown tool should not be in any group")
	}
}

func TestToolsForGroup_Workflow(t *testing.T) {
	tools := ToolsForGroup(GroupWorkflow)
	if len(tools) == 0 {
		t.Fatal("expected workflow tools, got none")
	}

	// Should include workflow-specific and shared tools.
	set := make(map[string]bool, len(tools))
	for _, name := range tools {
		set[name] = true
	}
	if !set[DiscoverActionTypes] {
		t.Error("workflow group should include discover_action_types")
	}
	if !set[SearchDatabaseSchema] {
		t.Error("workflow group should include shared tool search_database_schema")
	}
	if set[GetPageConfig] {
		t.Error("workflow group should NOT include tables-only tool get_page_config")
	}
}

func TestToolsForGroup_Tables(t *testing.T) {
	tools := ToolsForGroup(GroupTables)
	if len(tools) == 0 {
		t.Fatal("expected tables tools, got none")
	}

	set := make(map[string]bool, len(tools))
	for _, name := range tools {
		set[name] = true
	}
	if !set[GetPageConfig] {
		t.Error("tables group should include get_page_config")
	}
	if !set[SearchEnums] {
		t.Error("tables group should include shared tool search_enums")
	}
	if set[CreateWorkflow] {
		t.Error("tables group should NOT include workflow-only tool create_workflow")
	}
}

func TestAllTools_Count(t *testing.T) {
	all := AllTools()
	// 21 workflow-only + 19 tables-only + 2 shared = 42
	if len(all) != 42 {
		names := make([]string, len(all))
		copy(names, all)
		sort.Strings(names)
		t.Errorf("expected 42 tools, got %d: %v", len(all), names)
	}
}

func TestToolsForGroup_NoDuplicates(t *testing.T) {
	for _, group := range []ToolGroup{GroupWorkflow, GroupTables} {
		tools := ToolsForGroup(group)
		seen := make(map[string]bool, len(tools))
		for _, name := range tools {
			if seen[name] {
				t.Errorf("duplicate tool %q in group %s", name, group)
			}
			seen[name] = true
		}
	}
}
