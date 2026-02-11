package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/timmaaaz/ichor/mcp/internal/client"
)

// RegisterAnalysisTools adds workflow analysis and advisory tools to the MCP server.
func RegisterAnalysisTools(s *mcp.Server, c *client.Client) {
	// analyze_workflow — analyze a workflow for complexity and potential issues.
	type AnalyzeWorkflowArgs struct {
		ID string `json:"id" jsonschema:"UUID of the workflow rule to analyze,required"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "analyze_workflow",
		Description: "Analyze a workflow for complexity, coverage gaps, and optimization opportunities. Returns action count, edge count, branching complexity, and suggestions.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args AnalyzeWorkflowArgs) (*mcp.CallToolResult, any, error) {
		if args.ID == "" {
			return errorResult("id is required"), nil, nil
		}

		// Fetch the full workflow.
		rule, err := c.GetWorkflowRule(ctx, args.ID)
		if err != nil {
			return errorResult("Failed to fetch rule: " + err.Error()), nil, nil
		}

		actions, err := c.GetWorkflowRuleActions(ctx, args.ID)
		if err != nil {
			return errorResult("Failed to fetch actions: " + err.Error()), nil, nil
		}

		edges, err := c.GetWorkflowRuleEdges(ctx, args.ID)
		if err != nil {
			return errorResult("Failed to fetch edges: " + err.Error()), nil, nil
		}

		// Parse to count structures.
		var actionList []json.RawMessage
		var edgeList []json.RawMessage
		json.Unmarshal(actions, &actionList)
		json.Unmarshal(edges, &edgeList)

		// Analyze action types.
		actionTypeCounts := make(map[string]int)
		for _, a := range actionList {
			var action struct {
				ActionType string `json:"action_type"`
			}
			json.Unmarshal(a, &action)
			actionTypeCounts[action.ActionType]++
		}

		// Count edges per source (branching factor).
		edgesPerSource := make(map[string]int)
		for _, e := range edgeList {
			var edge struct {
				SourceActionID string `json:"source_action_id"`
			}
			json.Unmarshal(e, &edge)
			edgesPerSource[edge.SourceActionID]++
		}

		maxBranching := 0
		for _, count := range edgesPerSource {
			if count > maxBranching {
				maxBranching = count
			}
		}

		// Build suggestions.
		suggestions := []string{}
		if len(actionList) > 10 {
			suggestions = append(suggestions, "Consider splitting into multiple workflows — large graphs are harder to maintain")
		}
		if maxBranching > 3 {
			suggestions = append(suggestions, fmt.Sprintf("High branching factor (%d) detected — consider simplifying conditional logic", maxBranching))
		}
		if _, hasCondition := actionTypeCounts["evaluate_condition"]; !hasCondition && len(actionList) > 3 {
			suggestions = append(suggestions, "No conditional branching — consider adding evaluate_condition for error handling")
		}
		if _, hasAlert := actionTypeCounts["create_alert"]; !hasAlert {
			suggestions = append(suggestions, "No alerts — consider adding create_alert for failure notifications")
		}
		if _, hasAudit := actionTypeCounts["log_audit_entry"]; !hasAudit {
			suggestions = append(suggestions, "No audit logging — consider adding log_audit_entry for traceability")
		}

		analysis := map[string]any{
			"rule":               rule,
			"action_count":       len(actionList),
			"edge_count":         len(edgeList),
			"action_type_counts": actionTypeCounts,
			"max_branching":      maxBranching,
			"suggestions":        suggestions,
		}

		data, _ := json.Marshal(analysis)
		return jsonResult(data), nil, nil
	})

	// suggest_templates — suggest action templates for a use case.
	type SuggestTemplatesArgs struct {
		UseCase string `json:"use_case" jsonschema:"Description of what the workflow should accomplish (e.g. 'notify on low inventory', 'approve orders over $1000'),required"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "suggest_templates",
		Description: "Given a use case description, suggest relevant workflow action templates and action types that could be used to implement it.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args SuggestTemplatesArgs) (*mcp.CallToolResult, any, error) {
		if args.UseCase == "" {
			return errorResult("use_case is required"), nil, nil
		}

		// Fetch available templates and action types.
		templates, _ := c.GetActiveTemplates(ctx)
		actionTypes, _ := c.GetActionTypes(ctx)

		result := map[string]any{
			"use_case":              args.UseCase,
			"available_templates":   json.RawMessage(templates),
			"available_action_types": json.RawMessage(actionTypes),
			"guidance": "Based on the use case, select action types from the list above. " +
				"Check if any existing templates match your needs. " +
				"Common patterns: " +
				"notification workflows use send_email + create_alert; " +
				"approval workflows use seek_approval + evaluate_condition; " +
				"inventory workflows use check_inventory + reserve_inventory + commit_allocation; " +
				"data sync workflows use lookup_entity + update_field or create_entity.",
		}

		data, _ := json.Marshal(result)
		return jsonResult(data), nil, nil
	})

	// show_cascade — show what downstream workflows would trigger from an entity change.
	type ShowCascadeArgs struct {
		Entity string `json:"entity" jsonschema:"Entity type to check for cascade triggers (e.g. 'orders', 'products'),required"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "show_cascade",
		Description: "Show which workflow rules are triggered by changes to a given entity. Helps understand downstream automation effects before making changes.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ShowCascadeArgs) (*mcp.CallToolResult, any, error) {
		if args.Entity == "" {
			return errorResult("entity is required"), nil, nil
		}

		// Fetch all workflow rules.
		rules, err := c.GetWorkflowRules(ctx)
		if err != nil {
			return errorResult("Failed to fetch rules: " + err.Error()), nil, nil
		}

		// Parse and filter rules matching this entity.
		var allRules []json.RawMessage
		json.Unmarshal(rules, &allRules)

		var matching []json.RawMessage
		for _, r := range allRules {
			var rule struct {
				EntityType string `json:"entity_type"`
				IsActive   bool   `json:"is_active"`
			}
			json.Unmarshal(r, &rule)
			if rule.EntityType == args.Entity {
				matching = append(matching, r)
			}
		}

		result := map[string]any{
			"entity":          args.Entity,
			"triggered_rules": matching,
			"rule_count":      len(matching),
		}
		if len(matching) == 0 {
			result["message"] = fmt.Sprintf("No workflows are triggered by changes to %s", args.Entity)
		} else {
			result["message"] = fmt.Sprintf("%d workflow(s) will trigger when %s is created/updated/deleted", len(matching), args.Entity)
		}

		data, _ := json.Marshal(result)
		return jsonResult(data), nil, nil
	})
}
