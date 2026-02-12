package tools

import (
	"context"
	"encoding/json"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/timmaaaz/ichor/mcp/internal/client"
)

// RegisterWorkflowWriteTools adds workflow mutation tools to the MCP server.
func RegisterWorkflowWriteTools(s *mcp.Server, c *client.Client) {
	// validate_workflow — dry-run validation without committing.
	type ValidateWorkflowArgs struct {
		Workflow json.RawMessage `json:"workflow"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "validate_workflow",
		Description: "Validate a workflow definition without saving it. Returns validation results including whether the graph is valid, any errors, and action/edge counts. Use this before create_workflow to catch errors early.",
		Annotations: &mcp.ToolAnnotations{
			// Read-only in effect since dry-run doesn't write.
		},
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"workflow": {Type: "object", Description: "Full workflow save payload (rule + actions + edges) to validate"},
			},
			Required: []string{"workflow"},
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ValidateWorkflowArgs) (*mcp.CallToolResult, any, error) {
		data, err := c.ValidateWorkflow(ctx, args.Workflow)
		if err != nil {
			return errorResult("Validation request failed: " + err.Error()), nil, nil
		}
		return jsonResult(data), nil, nil
	})

	// create_workflow — create a new workflow rule with actions and edges.
	type CreateWorkflowArgs struct {
		Workflow json.RawMessage `json:"workflow"`
		Validate *bool           `json:"validate,omitempty"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "create_workflow",
		Description: "Create a new workflow automation rule with its action graph. By default, validates the workflow first using dry-run and only creates if validation passes. Set validate=false to skip pre-validation.",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"workflow": {Type: "object", Description: "Full workflow save payload (rule + actions + edges)"},
				"validate": {Type: "boolean", Description: "If true, run dry-run validation first and abort on errors (default: true)"},
			},
			Required: []string{"workflow"},
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, args CreateWorkflowArgs) (*mcp.CallToolResult, any, error) {
		shouldValidate := args.Validate == nil || *args.Validate

		if shouldValidate {
			valResult, err := c.ValidateWorkflow(ctx, args.Workflow)
			if err != nil {
				return errorResult("Pre-validation failed: " + err.Error()), nil, nil
			}

			var result struct {
				Valid  bool     `json:"valid"`
				Errors []string `json:"errors"`
			}
			if err := json.Unmarshal(valResult, &result); err == nil && !result.Valid {
				resp := map[string]any{
					"created":           false,
					"validation_errors": result.Errors,
					"message":           "Workflow validation failed. Fix the errors and try again.",
				}
				data, _ := json.Marshal(resp)
				return jsonResult(data), nil, nil
			}
		}

		data, err := c.CreateWorkflow(ctx, args.Workflow)
		if err != nil {
			return errorResult("Failed to create workflow: " + err.Error()), nil, nil
		}
		return jsonResult(data), nil, nil
	})

	// update_workflow — update an existing workflow rule.
	type UpdateWorkflowArgs struct {
		ID       string          `json:"id"`
		Workflow json.RawMessage `json:"workflow"`
		Validate *bool           `json:"validate,omitempty"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "update_workflow",
		Description: "Update an existing workflow rule and its action graph. By default, validates first using dry-run.",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"id":       {Type: "string", Description: "UUID of the workflow rule to update"},
				"workflow": {Type: "object", Description: "Full workflow save payload (rule + actions + edges)"},
				"validate": {Type: "boolean", Description: "If true, run dry-run validation first (default: true)"},
			},
			Required: []string{"id", "workflow"},
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, args UpdateWorkflowArgs) (*mcp.CallToolResult, any, error) {
		if args.ID == "" {
			return errorResult("id is required"), nil, nil
		}

		shouldValidate := args.Validate == nil || *args.Validate

		if shouldValidate {
			valResult, err := c.ValidateWorkflow(ctx, args.Workflow)
			if err != nil {
				return errorResult("Pre-validation failed: " + err.Error()), nil, nil
			}

			var result struct {
				Valid  bool     `json:"valid"`
				Errors []string `json:"errors"`
			}
			if err := json.Unmarshal(valResult, &result); err == nil && !result.Valid {
				resp := map[string]any{
					"updated":           false,
					"validation_errors": result.Errors,
					"message":           "Workflow validation failed. Fix the errors and try again.",
				}
				data, _ := json.Marshal(resp)
				return jsonResult(data), nil, nil
			}
		}

		data, err := c.UpdateWorkflow(ctx, args.ID, args.Workflow)
		if err != nil {
			return errorResult("Failed to update workflow: " + err.Error()), nil, nil
		}
		return jsonResult(data), nil, nil
	})
}
