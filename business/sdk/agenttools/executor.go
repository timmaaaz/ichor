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

	// Alerts
	case "list_my_alerts":
		return e.listMyAlerts(ctx, tc, token)
	case "get_alert_detail":
		var p struct {
			AlertID string `json:"alert_id"`
		}
		if err := json.Unmarshal(tc.Input, &p); err != nil {
			return nil, fmt.Errorf("bad params: %w", err)
		}
		if err := requireUUID(p.AlertID, "alert_id"); err != nil {
			return nil, err
		}
		return e.get(ctx, "/v1/workflow/alerts/"+p.AlertID, token)

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
		return e.getWorkflowSummary(ctx, p.RuleID, token)
	case "explain_workflow_node":
		var p struct {
			RuleID     string `json:"rule_id"`
			Identifier string `json:"identifier"`
		}
		if err := json.Unmarshal(tc.Input, &p); err != nil {
			return nil, fmt.Errorf("bad params: %w", err)
		}
		if err := requireUUID(p.RuleID, "rule_id"); err != nil {
			return nil, err
		}
		if p.Identifier == "" {
			return nil, fmt.Errorf("identifier is required")
		}
		return e.explainWorkflowNode(ctx, p.RuleID, p.Identifier, token)
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

// getWorkflowSummary fetches rule + actions + edges, then returns a compact
// summary with flow outline instead of raw JSON. This is much easier for
// smaller LLMs to interpret than the full action/edge payloads.
func (e *Executor) getWorkflowSummary(ctx context.Context, ruleID string, token string) (json.RawMessage, error) {
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

	result := map[string]any{
		"rule": json.RawMessage(rule),
	}

	// Parse graph and compute a compact summary with flow outline.
	graph, graphErr := parseWorkflowGraph(actions, edges)
	if graphErr == nil {
		// Enrich create_alert configs with human-readable recipient names
		// so the summary flow outline can include them inline.
		for i := range graph.actions {
			if graph.actions[i].ActionType == "create_alert" && len(graph.actions[i].Config) > 0 {
				graph.actions[i].Config = e.enrichCreateAlertConfig(ctx, graph.actions[i].Config, token)
			}
		}
		result["summary"] = graph.computeSummary()
	}

	// NOTE: Raw actions/edges are omitted to keep the response compact for
	// smaller LLMs. If a larger model is in use and needs full detail, the
	// following lines can be uncommented:
	//
	// result["actions"] = json.RawMessage(actions)
	// result["edges"] = json.RawMessage(edges)

	b, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("marshal workflow summary: %w", err)
	}
	return b, nil
}

// explainWorkflowNode fetches actions + edges for a rule, then returns
// detailed information about a specific action node identified by name or UUID.
func (e *Executor) explainWorkflowNode(ctx context.Context, ruleID, identifier, token string) (json.RawMessage, error) {
	actions, err := e.get(ctx, "/v1/workflow/rules/"+ruleID+"/actions", token)
	if err != nil {
		return nil, fmt.Errorf("get actions: %w", err)
	}
	edges, err := e.get(ctx, "/v1/workflow/rules/"+ruleID+"/edges", token)
	if err != nil {
		return nil, fmt.Errorf("get edges: %w", err)
	}

	graph, err := parseWorkflowGraph(actions, edges)
	if err != nil {
		return nil, fmt.Errorf("parse workflow graph: %w", err)
	}

	action := graph.findAction(identifier)
	if action == nil {
		return nil, fmt.Errorf("action not found: %s", identifier)
	}

	// For create_alert actions, enrich the config with human-readable
	// recipient names/emails so the LLM can describe them to the user.
	if action.ActionType == "create_alert" && len(action.Config) > 0 {
		action.Config = e.enrichCreateAlertConfig(ctx, action.Config, token)
	}

	explanation := graph.explainNode(action)

	b, err := json.Marshal(explanation)
	if err != nil {
		return nil, fmt.Errorf("marshal node explanation: %w", err)
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

// listMyAlerts fetches the current user's alerts with optional filters.
func (e *Executor) listMyAlerts(ctx context.Context, tc llm.ToolCall, token string) (json.RawMessage, error) {
	var p struct {
		Status   string `json:"status"`
		Severity string `json:"severity"`
		Page     string `json:"page"`
		Rows     string `json:"rows"`
	}
	if err := json.Unmarshal(tc.Input, &p); err != nil {
		return nil, fmt.Errorf("bad params: %w", err)
	}

	path := "/v1/workflow/alerts/mine?"
	if p.Status != "" {
		path += "status=" + p.Status + "&"
	}
	if p.Severity != "" {
		path += "severity=" + p.Severity + "&"
	}
	if p.Page != "" {
		path += "page=" + p.Page + "&"
	}
	if p.Rows != "" {
		path += "rows=" + p.Rows + "&"
	}

	return e.get(ctx, path, token)
}

// enrichCreateAlertConfig parses a create_alert action config and resolves
// recipient UUIDs to human-readable names/emails via the REST API.
func (e *Executor) enrichCreateAlertConfig(ctx context.Context, config json.RawMessage, token string) json.RawMessage {
	var cfg struct {
		AlertType  string `json:"alert_type"`
		Severity   string `json:"severity"`
		Title      string `json:"title"`
		Message    string `json:"message"`
		Recipients struct {
			Users []string `json:"users"`
			Roles []string `json:"roles"`
		} `json:"recipients"`
		Context      json.RawMessage `json:"context"`
		ResolvePrior bool            `json:"resolve_prior"`
	}
	if err := json.Unmarshal(config, &cfg); err != nil {
		return config
	}

	type enrichedRecipient struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email,omitempty"`
	}

	var enrichedUsers []enrichedRecipient
	for _, userID := range cfg.Recipients.Users {
		r := enrichedRecipient{ID: userID}
		userData, err := e.get(ctx, "/v1/core/users/"+userID, token)
		if err == nil {
			var user struct {
				FirstName string `json:"firstName"`
				LastName  string `json:"lastName"`
				Email     string `json:"email"`
			}
			if json.Unmarshal(userData, &user) == nil {
				r.Name = user.FirstName + " " + user.LastName
				r.Email = user.Email
			}
		}
		if r.Name == "" {
			r.Name = "Unknown User"
		}
		enrichedUsers = append(enrichedUsers, r)
	}

	var enrichedRoles []enrichedRecipient
	for _, roleID := range cfg.Recipients.Roles {
		r := enrichedRecipient{ID: roleID}
		roleData, err := e.get(ctx, "/v1/core/roles/"+roleID, token)
		if err == nil {
			var role struct {
				Name string `json:"name"`
			}
			if json.Unmarshal(roleData, &role) == nil {
				r.Name = role.Name
			}
		}
		if r.Name == "" {
			r.Name = "Unknown Role"
		}
		enrichedRoles = append(enrichedRoles, r)
	}

	enriched := map[string]any{
		"alert_type": cfg.AlertType,
		"severity":   cfg.Severity,
		"title":      cfg.Title,
		"message":    cfg.Message,
		"recipients": map[string]any{
			"users": enrichedUsers,
			"roles": enrichedRoles,
		},
		"resolve_prior": cfg.ResolvePrior,
	}
	if len(cfg.Context) > 0 {
		enriched["context"] = cfg.Context
	}

	b, err := json.Marshal(enriched)
	if err != nil {
		return config
	}
	return b
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
