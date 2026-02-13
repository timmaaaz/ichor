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
		Description: "Get a workflow rule by ID including its full action graph (actions, edges, and output ports). Returns the complete workflow definition with a summary containing node count, branch count, action types used, and a human-readable flow outline. Use this to understand what a workflow does before drilling into specific nodes with explain_workflow_node.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args GetWorkflowArgs) (*mcp.CallToolResult, any, error) {
		if args.ID == "" {
			return errorResult("id is required"), nil, nil
		}

		// Fetch rule, actions, and edges.
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

		// Parse graph and compute summary.
		graph, graphErr := parseWorkflowGraph(actions, edges)

		result := map[string]any{
			"rule":    json.RawMessage(rule),
			"actions": json.RawMessage(actions),
			"edges":   json.RawMessage(edges),
		}

		if graphErr == nil {
			result["summary"] = graph.computeSummary()
		}

		data, err := json.Marshal(result)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to marshal result: %v", err)), nil, nil
		}

		return jsonResult(data), nil, nil
	})

	// explain_workflow_node — explain a specific action/node in a workflow.
	type ExplainWorkflowNodeArgs struct {
		RuleID     string `json:"rule_id" jsonschema:"UUID of the workflow rule,required"`
		Identifier string `json:"identifier" jsonschema:"Action name or UUID to look up,required"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "explain_workflow_node",
		Description: "Explain a specific action/node in a workflow by name or ID. Returns the action details, incoming edges (what feeds into it), outgoing edges (what it feeds with output port labels), depth from the start node, and action type metadata. Use after get_workflow to drill into individual nodes.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ExplainWorkflowNodeArgs) (*mcp.CallToolResult, any, error) {
		if args.RuleID == "" || args.Identifier == "" {
			return errorResult("rule_id and identifier are required"), nil, nil
		}

		// Fetch actions and edges for the rule.
		actions, err := c.GetWorkflowRuleActions(ctx, args.RuleID)
		if err != nil {
			return errorResult("Failed to fetch actions: " + err.Error()), nil, nil
		}

		edges, err := c.GetWorkflowRuleEdges(ctx, args.RuleID)
		if err != nil {
			return errorResult("Failed to fetch edges: " + err.Error()), nil, nil
		}

		// Parse graph.
		graph, err := parseWorkflowGraph(actions, edges)
		if err != nil {
			return errorResult("Failed to parse workflow graph: " + err.Error()), nil, nil
		}

		// Find action by name or ID.
		action := graph.findAction(args.Identifier)
		if action == nil {
			return errorResult(fmt.Sprintf("Action not found: %s", args.Identifier)), nil, nil
		}

		// Build explanation.
		explanation := graph.explainNode(action)

		// Fetch action type metadata if available.
		if action.ActionType != "" {
			if typeInfo, fetchErr := c.GetActionTypeSchema(ctx, action.ActionType); fetchErr == nil {
				explanation.ActionTypeInfo = typeInfo
			}
		}

		data, err := json.Marshal(explanation)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to marshal result: %v", err)), nil, nil
		}

		return jsonResult(data), nil, nil
	})

	// explain_workflow_path — trace a path through a workflow from a starting node.
	type ExplainWorkflowPathArgs struct {
		RuleID string `json:"rule_id" jsonschema:"UUID of the workflow rule,required"`
		From   string `json:"from,omitempty" jsonschema:"Action name or UUID to start tracing from (defaults to start nodes)"`
		Output string `json:"output,omitempty" jsonschema:"Output port to follow from the starting node (e.g. false, insufficient, rejected)"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "explain_workflow_path",
		Description: "Trace a path through a workflow graph starting from a given node. If 'output' is provided, only follows edges from 'from' whose output port matches (e.g. follow the 'false' branch of a condition). If 'from' is omitted, traces from the start nodes. Returns a structured path with each action and the edge label that leads to it, plus a text outline.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ExplainWorkflowPathArgs) (*mcp.CallToolResult, any, error) {
		if args.RuleID == "" {
			return errorResult("rule_id is required"), nil, nil
		}
		if args.Output != "" && args.From == "" {
			return errorResult("output requires from to be specified"), nil, nil
		}

		actions, err := c.GetWorkflowRuleActions(ctx, args.RuleID)
		if err != nil {
			return errorResult("Failed to fetch actions: " + err.Error()), nil, nil
		}

		edges, err := c.GetWorkflowRuleEdges(ctx, args.RuleID)
		if err != nil {
			return errorResult("Failed to fetch edges: " + err.Error()), nil, nil
		}

		graph, err := parseWorkflowGraph(actions, edges)
		if err != nil {
			return errorResult("Failed to parse workflow graph: " + err.Error()), nil, nil
		}

		var result pathResult

		if args.From == "" {
			// Trace from start nodes.
			result = graph.tracePath(graph.startNodes)
			result.StartingFrom = "(start)"
		} else {
			action := graph.findAction(args.From)
			if action == nil {
				return errorResult(fmt.Sprintf("Action not found: %s", args.From)), nil, nil
			}
			result.StartingFrom = action.Name

			if args.Output != "" {
				// Filter outgoing edges to only those matching the output port.
				result.OutputFollowed = args.Output
				var targetIDs []string
				for _, edge := range graph.outgoing[action.ID] {
					if edge.SourceOutput == args.Output {
						targetIDs = append(targetIDs, edge.TargetActionID)
					}
				}
				if len(targetIDs) == 0 {
					return errorResult(fmt.Sprintf("No edges from %q with output %q", action.Name, args.Output)), nil, nil
				}
				traced := graph.tracePath(targetIDs)
				result.Path = traced.Path
				result.TextOutline = traced.TextOutline
			} else {
				// Trace full subtree from the given action.
				traced := graph.tracePath([]string{action.ID})
				result.Path = traced.Path
				result.TextOutline = traced.TextOutline
			}
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
