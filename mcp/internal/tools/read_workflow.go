package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/timmaaaz/ichor/mcp/internal/client"
)

// RegisterWorkflowReadTools adds read-only workflow tools to the MCP server.
func RegisterWorkflowReadTools(s *mcp.Server, c *client.Client) {
	// get_workflow — get a workflow rule with its actions and edges.
	type GetWorkflowArgs struct {
		ID string `json:"id" jsonschema:"UUID of the workflow rule,required"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_workflow",
		Description: "Get a workflow rule by ID including its full action graph (actions, edges, and output ports). Returns the complete workflow definition.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args GetWorkflowArgs) (*mcp.CallToolResult, any, error) {
		if args.ID == "" {
			return errorResult("id is required"), nil, nil
		}

		// Fetch rule, actions, and edges in parallel concept but sequential for simplicity.
		rule, err := c.GetWorkflowRule(ctx, args.ID)
		if err != nil {
			return errorResult("Failed to fetch workflow rule: " + err.Error()), nil, nil
		}

		actions, err := c.GetWorkflowRuleActions(ctx, args.ID)
		if err != nil {
			return errorResult("Failed to fetch workflow actions: " + err.Error()), nil, nil
		}

		edges, err := c.GetWorkflowRuleEdges(ctx, args.ID)
		if err != nil {
			return errorResult("Failed to fetch workflow edges: " + err.Error()), nil, nil
		}

		// Merge into a single response.
		result := map[string]json.RawMessage{
			"rule":    rule,
			"actions": actions,
			"edges":   edges,
		}

		data, err := json.Marshal(result)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to marshal result: %v", err)), nil, nil
		}

		return jsonResult(data), nil, nil
	})

	// list_workflows — list all workflow rules.
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_workflows",
		Description: "List all workflow automation rules in the system with their trigger types, entity types, and active status.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		data, err := c.GetWorkflowRules(ctx)
		if err != nil {
			return errorResult("Failed to list workflows: " + err.Error()), nil, nil
		}
		return jsonResult(data), nil, nil
	})

	// list_action_templates — list workflow action templates.
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_action_templates",
		Description: "List all workflow action templates. Templates provide reusable action configurations that can be referenced by rule actions.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		data, err := c.GetTemplates(ctx)
		if err != nil {
			return errorResult("Failed to list templates: " + err.Error()), nil, nil
		}
		return jsonResult(data), nil, nil
	})
}
