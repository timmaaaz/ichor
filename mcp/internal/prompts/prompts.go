// Package prompts provides MCP prompts for guided workflow, page, and form building.
package prompts

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/timmaaaz/ichor/mcp/internal/client"
)

// RegisterPrompts adds guided prompts to the MCP server.
func RegisterPrompts(s *mcp.Server, c *client.Client) {
	// build-workflow prompt
	s.AddPrompt(&mcp.Prompt{
		Name:        "build-workflow",
		Description: "Guide the user through building a workflow automation rule for a specific trigger and entity.",
		Arguments: []*mcp.PromptArgument{
			{Name: "trigger", Description: "Trigger type (e.g., on_create, on_update, on_delete)", Required: true},
			{Name: "entity", Description: "Entity type to trigger on (e.g., orders, products)", Required: true},
		},
	}, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		trigger := req.Params.Arguments["trigger"]
		entity := req.Params.Arguments["entity"]

		// Fetch context: action types and entity schema.
		actionTypes, _ := c.GetActionTypes(ctx)
		triggerTypes, _ := c.GetTriggerTypes(ctx)

		contextText := fmt.Sprintf(`## Workflow Builder Context

**Target**: Build a workflow for trigger=%s on entity=%s

### Available Trigger Types
%s

### Available Action Types (with output ports and config schemas)
%s

### Instructions
1. Start by confirming the trigger type and entity are valid
2. Design the action graph (what should happen when the trigger fires)
3. For each action, choose an appropriate action type and configure it
4. Define edges between actions using output ports for conditional routing
5. Use validate_workflow to check the graph before creating it
6. Use create_workflow to save the final workflow

### Key Rules
- Every workflow needs exactly one start edge (from the trigger to the first action)
- Actions are connected by edges that reference output ports
- Each action type has specific config requirements (see schemas above)
- Use evaluate_condition for branching logic
- Use delay for time-based pauses
- The graph must be a valid DAG (no cycles)`, trigger, entity, string(triggerTypes), string(actionTypes))

		return &mcp.GetPromptResult{
			Description: fmt.Sprintf("Build a %s workflow for %s", trigger, entity),
			Messages: []*mcp.PromptMessage{
				{
					Role:    "user",
					Content: &mcp.TextContent{Text: contextText},
				},
			},
		}, nil
	})

	// configure-page prompt
	s.AddPrompt(&mcp.Prompt{
		Name:        "configure-page",
		Description: "Guide the user through configuring a page layout for an entity.",
		Arguments: []*mcp.PromptArgument{
			{Name: "entity", Description: "Entity to build a page for (e.g., orders, products)", Required: true},
		},
	}, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		entity := req.Params.Arguments["entity"]

		contentTypes, _ := c.GetContentTypes(ctx)
		fieldTypes, _ := c.GetFieldTypes(ctx)

		contextText := fmt.Sprintf(`## Page Configuration Context

**Target**: Configure a page layout for entity=%s

### Available Content Types
%s

### Available Form Field Types
%s

### Instructions
1. Create a page config with a descriptive name
2. Add content blocks: tables for data display, forms for data entry, charts for visualization
3. Use containers with grid layout for responsive multi-column layouts
4. Use tabs to organize related content
5. Configure layout JSONB for responsive breakpoints (default, sm, md, lg, xl, 2xl)
6. Reference existing table configs or create new ones for data tables
7. Reference existing forms or create new ones for data entry

### Layout Tips
- Use colSpan with ResponsiveValue for responsive widths
- gridCols on containers defines child grid
- gap controls spacing (use Tailwind classes like "gap-4")
- containerType: "grid-12" for 12-column grid, "stack" for vertical, "tab" for tabbed`, entity, string(contentTypes), string(fieldTypes))

		return &mcp.GetPromptResult{
			Description: fmt.Sprintf("Configure a page for %s", entity),
			Messages: []*mcp.PromptMessage{
				{
					Role:    "user",
					Content: &mcp.TextContent{Text: contextText},
				},
			},
		}, nil
	})

	// design-form prompt
	s.AddPrompt(&mcp.Prompt{
		Name:        "design-form",
		Description: "Guide the user through designing a data entry form for an entity.",
		Arguments: []*mcp.PromptArgument{
			{Name: "entity", Description: "Entity to build a form for (e.g., orders, products)", Required: true},
		},
	}, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		entity := req.Params.Arguments["entity"]

		fieldTypes, _ := c.GetFieldTypes(ctx)

		contextText := fmt.Sprintf(`## Form Design Context

**Target**: Design a data entry form for entity=%s

### Available Field Types (with config schemas)
%s

### Instructions
1. First, explore the entity's database schema using search_database_schema to understand its columns and relationships
2. Create a form definition targeting this entity
3. For each column that needs user input, add a form field:
   - text/textarea for string columns
   - number/currency/percent for numeric columns
   - date/datetime/time for temporal columns
   - dropdown/smart-combobox for foreign key references
   - enum for PostgreSQL enum columns
   - boolean for true/false columns
   - hidden for auto-populated fields (parent IDs, etc.)
4. Configure each field's validation rules and default values
5. Use the field type schema to understand all available config options
6. Set appropriate order_index values for field ordering

### Tips
- dropdown fields need entity, labelColumn, valueColumn config
- smart-combobox adds search/autocomplete to dropdown
- enum fields need enumName matching a PostgreSQL enum type
- lineitems creates inline child entity collections
- Use hidden fields for parent_id values that should be auto-set`, entity, string(fieldTypes))

		return &mcp.GetPromptResult{
			Description: fmt.Sprintf("Design a form for %s", entity),
			Messages: []*mcp.PromptMessage{
				{
					Role:    "user",
					Content: &mcp.TextContent{Text: contextText},
				},
			},
		}, nil
	})
}
