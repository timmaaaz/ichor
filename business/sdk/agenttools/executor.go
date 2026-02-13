package agenttools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/llm"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Executor runs tool calls by forwarding them to Ichor's own REST API,
// using the caller's JWT token so all requests honour existing auth/perms.
type Executor struct {
	log     *logger.Logger
	baseURL string
	http    *http.Client
}

// NewExecutor creates a tool executor. baseURL is the Ichor API root
// (e.g. "http://localhost:8080").
func NewExecutor(log *logger.Logger, baseURL string) *Executor {
	return &Executor{
		log:     log,
		baseURL: baseURL,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

// Execute dispatches a single tool call and returns the result.
// authToken is the raw Authorization header value (e.g. "Bearer eyJ…").
func (e *Executor) Execute(ctx context.Context, tc llm.ToolCall, authToken string) llm.ToolResult {
	e.log.Info(ctx, "AGENT-CHAT: executing tool",
		"tool", tc.Name,
		"tool_use_id", tc.ID)

	result, err := e.dispatch(ctx, tc, authToken)
	if err != nil {
		e.log.Error(ctx, "AGENT-CHAT: tool execution failed",
			"tool", tc.Name,
			"error", err)
		errJSON, _ := json.Marshal(map[string]string{"error": err.Error()})
		return llm.ToolResult{
			ToolUseID: tc.ID,
			Content:   string(errJSON),
			IsError:   true,
		}
	}

	return llm.ToolResult{
		ToolUseID: tc.ID,
		Content:   string(result),
		IsError:   false,
	}
}

func (e *Executor) dispatch(ctx context.Context, tc llm.ToolCall, token string) (json.RawMessage, error) {
	switch tc.Name {
	// Discovery
	case "discover_action_types":
		return e.get(ctx, "/v1/workflow/action-types", token)
	case "discover_trigger_types":
		return e.get(ctx, "/v1/workflow/trigger-types", token)
	case "discover_entities":
		return e.get(ctx, "/v1/workflow/entities", token)

	// Read
	case "get_workflow_rule":
		var p struct {
			RuleID string `json:"rule_id"`
		}
		if err := json.Unmarshal(tc.Input, &p); err != nil {
			return nil, fmt.Errorf("bad params: %w", err)
		}
		if err := requireUUID(p.RuleID, "rule_id"); err != nil {
			return nil, err
		}
		return e.getWorkflowFull(ctx, p.RuleID, token)
	case "list_workflow_rules":
		return e.get(ctx, "/v1/workflow/rules", token)

	// Write
	case "create_workflow":
		var p struct {
			Workflow json.RawMessage `json:"workflow"`
		}
		if err := json.Unmarshal(tc.Input, &p); err != nil {
			return nil, fmt.Errorf("bad params: %w", err)
		}
		return e.post(ctx, "/v1/workflow/rules/full", p.Workflow, token)
	case "update_workflow":
		var p struct {
			RuleID   string          `json:"rule_id"`
			Workflow json.RawMessage `json:"workflow"`
		}
		if err := json.Unmarshal(tc.Input, &p); err != nil {
			return nil, fmt.Errorf("bad params: %w", err)
		}
		if err := requireUUID(p.RuleID, "rule_id"); err != nil {
			return nil, err
		}
		return e.put(ctx, "/v1/workflow/rules/"+p.RuleID+"/full", p.Workflow, token)
	case "validate_workflow", "preview_workflow":
		var p struct {
			Workflow json.RawMessage `json:"workflow"`
		}
		if err := json.Unmarshal(tc.Input, &p); err != nil {
			return nil, fmt.Errorf("bad params: %w", err)
		}
		return e.post(ctx, "/v1/workflow/rules/full?dry_run=true", p.Workflow, token)

	default:
		return nil, fmt.Errorf("unknown tool: %s", tc.Name)
	}
}

// getWorkflowFull merges rule + actions + edges into one response,
// mirroring what the MCP get_workflow tool does.
func (e *Executor) getWorkflowFull(ctx context.Context, ruleID string, token string) (json.RawMessage, error) {
	rule, err := e.get(ctx, "/v1/workflow/rules/"+ruleID, token)
	if err != nil {
		return nil, fmt.Errorf("get rule: %w", err)
	}
	actions, err := e.get(ctx, "/v1/workflow/rules/"+ruleID+"/actions", token)
	if err != nil {
		return nil, fmt.Errorf("get actions: %w", err)
	}
	edges, err := e.get(ctx, "/v1/workflow/rules/"+ruleID+"/edges", token)
	if err != nil {
		return nil, fmt.Errorf("get edges: %w", err)
	}

	merged := map[string]json.RawMessage{
		"rule":    rule,
		"actions": actions,
		"edges":   edges,
	}
	b, err := json.Marshal(merged)
	if err != nil {
		return nil, fmt.Errorf("marshal merged workflow: %w", err)
	}
	return b, nil
}

// =========================================================================
// HTTP helpers
// =========================================================================

func (e *Executor) get(ctx context.Context, path, token string) (json.RawMessage, error) {
	return e.do(ctx, http.MethodGet, path, nil, token)
}

func (e *Executor) post(ctx context.Context, path string, body json.RawMessage, token string) (json.RawMessage, error) {
	return e.do(ctx, http.MethodPost, path, body, token)
}

func (e *Executor) put(ctx context.Context, path string, body json.RawMessage, token string) (json.RawMessage, error) {
	return e.do(ctx, http.MethodPut, path, body, token)
}

func (e *Executor) do(ctx context.Context, method, path string, body json.RawMessage, token string) (json.RawMessage, error) {
	url := e.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", token)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := e.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return json.RawMessage(respBody), nil
}

// requireUUID validates that s is a well-formed UUID string (36 chars with
// hyphens). This catches truncated values from small LLMs before they hit
// the REST API, returning a descriptive error the LLM can self-correct from.
func requireUUID(s, field string) error {
	if _, err := uuid.Parse(s); err != nil {
		return fmt.Errorf("%s must be a full 36-character UUID (e.g. 35da6628-a96b-4bc4-a90f-8fa874ae48cc), got %q (length %d): check the workflow context for the correct value", field, s, len(s))
	}
	return nil
}
