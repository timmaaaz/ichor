package agenttools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/llm"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// draftWorkflow holds the in-memory state for an incrementally built workflow.
type draftWorkflow struct {
	lastAccess time.Time
	name       string
	entity     string // "schema.table" or UUID
	triggerType string // name or UUID
	description string
	triggerCond json.RawMessage // optional trigger_conditions
	actions     []draftAction
}

// draftAction is a single action within a draft, preserving the "after" field.
type draftAction struct {
	Name       string          `json:"name"`
	ActionType string          `json:"action_type"`
	Config     json.RawMessage `json:"action_config"`
	Desc       string          `json:"description,omitempty"`
	IsActive   bool            `json:"is_active"`
	After      string          `json:"after,omitempty"`
}

// Executor runs tool calls by forwarding them to Ichor's own REST API,
// using the caller's JWT token so all requests honour existing auth/perms.
type Executor struct {
	log     *logger.Logger
	baseURL string
	http    *http.Client

	// Caches for name→UUID resolution (populated lazily per token).
	entityCache      map[string]string // "schema.table" → entity UUID
	triggerTypeCache map[string]string // trigger name → trigger UUID
	actionTypeCache  map[string]actionTypeInfo // action type → info with ports
	ruleCache        map[string]string // rule name → rule UUID
	cacheMu          sync.Mutex

	// Draft state for incremental workflow building.
	drafts  map[string]*draftWorkflow
	draftMu sync.Mutex
}

// actionTypeInfo holds cached action type metadata for default port resolution.
type actionTypeInfo struct {
	DefaultPort string
}

const (
	draftTTL   = 10 * time.Minute
	maxDrafts  = 100 // per executor instance
)

// NewExecutor creates a tool executor. baseURL is the Ichor API root
// (e.g. "http://localhost:8080").
func NewExecutor(log *logger.Logger, baseURL string) *Executor {
	return &Executor{
		log:     log,
		baseURL: baseURL,
		http:    &http.Client{Timeout: 30 * time.Second},
		drafts:  make(map[string]*draftWorkflow),
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
	// Discovery (consolidated)
	case "discover":
		var p struct {
			Category string `json:"category"`
		}
		if err := json.Unmarshal(tc.Input, &p); err != nil {
			return nil, fmt.Errorf("bad params: %w", err)
		}
		switch p.Category {
		case "action_types":
			data, err := e.get(ctx, "/v1/workflow/action-types", token)
			if err != nil {
				return nil, err
			}
			return summarizeActionTypes(data), nil
		case "trigger_types":
			return e.get(ctx, "/v1/workflow/trigger-types", token)
		case "entities":
			return e.get(ctx, "/v1/workflow/entities", token)
		default:
			return nil, fmt.Errorf("unknown discover category %q — use action_types, trigger_types, or entities", p.Category)
		}

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
	case "list_alerts_for_rule":
		return e.listAlertsForRule(ctx, tc, token)

	// Read
	case "get_workflow_rule":
		var p struct {
			RuleID string `json:"workflow_id"`
		}
		if err := json.Unmarshal(tc.Input, &p); err != nil {
			return nil, fmt.Errorf("bad params: %w", err)
		}
		ruleID, err := e.resolveRuleID(ctx, p.RuleID, token)
		if err != nil {
			return nil, err
		}
		return e.getWorkflowSummary(ctx, ruleID, token)
	case "explain_workflow_node":
		var p struct {
			RuleID   string `json:"workflow_id"`
			NodeName string `json:"node_name"`
		}
		if err := json.Unmarshal(tc.Input, &p); err != nil {
			return nil, fmt.Errorf("bad params: %w", err)
		}
		if p.RuleID == "" {
			return nil, fmt.Errorf("workflow_id is required — provide it or ensure the workflow context is set")
		}
		ruleID, err := e.resolveRuleID(ctx, p.RuleID, token)
		if err != nil {
			return nil, err
		}
		if p.NodeName == "" {
			return nil, fmt.Errorf("node_name is required")
		}
		return e.explainWorkflowNode(ctx, ruleID, p.NodeName, token)
	case "list_workflow_rules":
		data, err := e.get(ctx, "/v1/workflow/rules?rows=500", token)
		if err != nil {
			return nil, err
		}
		return summarizeRuleList(formatPaginatedResponse(data)), nil

	// Preview (validate + send to user for approval)
	case "preview_workflow":
		var p struct {
			Workflow json.RawMessage `json:"workflow"`
		}
		if err := json.Unmarshal(tc.Input, &p); err != nil {
			return nil, fmt.Errorf("bad params: %w", err)
		}
		transformed, err := e.transformWorkflowPayload(ctx, p.Workflow, token)
		if err != nil {
			return nil, fmt.Errorf("transform workflow: %w", err)
		}
		return e.post(ctx, "/v1/workflow/rules/full?dry_run=true", transformed, token)

	// Draft builder
	case "start_draft":
		return e.handleStartDraft(ctx, tc, token)
	case "add_draft_action":
		return e.handleAddDraftAction(ctx, tc, token)
	case "remove_draft_action":
		return e.handleRemoveDraftAction(ctx, tc)
	case "preview_draft":
		return e.handlePreviewDraft(ctx, tc, token)

	default:
		return nil, fmt.Errorf("unknown tool: %s", tc.Name)
	}
}

// =========================================================================
// Workflow payload transformation (name resolution + edge shorthand)
// =========================================================================

// transformWorkflowPayload resolves name-based entity/trigger references to
// UUIDs and generates edges from "after" fields when no edges are provided.
func (e *Executor) transformWorkflowPayload(ctx context.Context, workflow json.RawMessage, token string) (json.RawMessage, error) {
	var w map[string]json.RawMessage
	if err := json.Unmarshal(workflow, &w); err != nil {
		return nil, fmt.Errorf("invalid workflow JSON: %w", err)
	}

	// --- Name-to-UUID resolution ---

	// Resolve "entity" → "entity_id"
	if raw, ok := w["entity"]; ok {
		var entity string
		if err := json.Unmarshal(raw, &entity); err == nil && entity != "" {
			resolved, err := e.resolveEntityID(ctx, entity, token)
			if err != nil {
				return nil, err
			}
			w["entity_id"] = mustMarshal(resolved)
			delete(w, "entity")
		}
	}

	// Resolve "trigger_type" → "trigger_type_id"
	if raw, ok := w["trigger_type"]; ok {
		var tt string
		if err := json.Unmarshal(raw, &tt); err == nil && tt != "" {
			resolved, err := e.resolveTriggerTypeID(ctx, tt, token)
			if err != nil {
				return nil, err
			}
			w["trigger_type_id"] = mustMarshal(resolved)
			delete(w, "trigger_type")
		}
	}

	// --- Edge shorthand ("after" field) ---

	actionsRaw, hasActions := w["actions"]
	edgesRaw, hasEdges := w["edges"]

	// Check if edges is empty/missing
	edgesEmpty := !hasEdges
	if hasEdges {
		var edges []json.RawMessage
		if json.Unmarshal(edgesRaw, &edges) == nil && len(edges) == 0 {
			edgesEmpty = true
		}
	}

	if hasActions && edgesEmpty {
		var actions []map[string]json.RawMessage
		if err := json.Unmarshal(actionsRaw, &actions); err != nil {
			return nil, fmt.Errorf("invalid actions array: %w", err)
		}

		// Check if any action has an "after" field.
		hasAfter := false
		for _, a := range actions {
			if _, ok := a["after"]; ok {
				hasAfter = true
				break
			}
		}

		if hasAfter {
			edges, err := e.generateEdgesFromAfter(ctx, actions, token)
			if err != nil {
				return nil, fmt.Errorf("generate edges from 'after': %w", err)
			}
			w["edges"] = mustMarshal(edges)

			// Strip "after" from actions before forwarding.
			for i := range actions {
				delete(actions[i], "after")
			}
			w["actions"] = mustMarshal(actions)
		}
	}

	return json.Marshal(w)
}

// generateEdgesFromAfter builds an edges array from action "after" fields.
// The first action without an "after" field gets a start edge.
func (e *Executor) generateEdgesFromAfter(ctx context.Context, actions []map[string]json.RawMessage, token string) ([]map[string]any, error) {
	// Build name→index map for temp ID resolution.
	nameToIdx := make(map[string]int, len(actions))
	for i, a := range actions {
		var name string
		if raw, ok := a["name"]; ok {
			if err := json.Unmarshal(raw, &name); err != nil {
				return nil, fmt.Errorf("action at index %d has invalid name: %w", i, err)
			}
		}
		if name == "" {
			return nil, fmt.Errorf("action at index %d has no name", i)
		}
		nameToIdx[name] = i
	}

	var edges []map[string]any
	startCount := 0
	edgeOrder := 0

	for i, a := range actions {
		afterRaw, hasAfter := a["after"]

		var afterStr string
		if hasAfter {
			if err := json.Unmarshal(afterRaw, &afterStr); err != nil {
				return nil, fmt.Errorf("action at index %d has invalid 'after' field: %w", i, err)
			}
		}

		if afterStr == "" {
			// This action has no predecessor — it's the start node.
			startCount++
			if startCount > 1 {
				return nil, fmt.Errorf("multiple actions without 'after' — only one start action is allowed")
			}
			edges = append(edges, map[string]any{
				"target_action_id": fmt.Sprintf("temp:%d", i),
				"edge_type":        "start",
				"edge_order":       edgeOrder,
			})
			edgeOrder++
			continue
		}

		// Parse "ActionName:port" or "ActionName"
		sourceName, port, err := parseAfterField(afterStr)
		if err != nil {
			return nil, err
		}

		sourceIdx, ok := nameToIdx[sourceName]
		if !ok {
			return nil, fmt.Errorf("action %q referenced in 'after' not found in actions list", sourceName)
		}

		// If no port specified, look up default port for the source action type.
		if port == "" {
			var actionType string
			if raw, ok := actions[sourceIdx]["action_type"]; ok {
				if err := json.Unmarshal(raw, &actionType); err != nil {
					return nil, fmt.Errorf("action %q has invalid action_type: %w", sourceName, err)
				}
			}
			if actionType == "" {
				return nil, fmt.Errorf("cannot determine default port: source action %q has no action_type", sourceName)
			}
			defaultPort, err := e.getDefaultOutputPort(ctx, actionType, token)
			if err != nil {
				return nil, fmt.Errorf("resolve default port for %q (%s): %w", sourceName, actionType, err)
			}
			port = defaultPort
		}

		edges = append(edges, map[string]any{
			"source_action_id": fmt.Sprintf("temp:%d", sourceIdx),
			"target_action_id": fmt.Sprintf("temp:%d", i),
			"edge_type":        "sequence",
			"source_output":    port,
			"edge_order":       edgeOrder,
		})
		edgeOrder++
	}

	if startCount == 0 {
		return nil, fmt.Errorf("no start action found — at least one action must omit the 'after' field")
	}

	return edges, nil
}

// parseAfterField splits "ActionName:port" into name and port.
// If no colon, port is empty string.
func parseAfterField(after string) (name, port string, err error) {
	after = strings.TrimSpace(after)
	if after == "" {
		return "", "", fmt.Errorf("empty 'after' value")
	}

	// Split on last colon to support action names with colons (unlikely but safe).
	idx := strings.LastIndex(after, ":")
	if idx == -1 {
		return after, "", nil
	}

	name = strings.TrimSpace(after[:idx])
	port = strings.TrimSpace(after[idx+1:])
	if name == "" {
		return "", "", fmt.Errorf("invalid 'after' format %q: action name is empty", after)
	}
	return name, port, nil
}

// =========================================================================
// Name-to-UUID resolution with caching
// =========================================================================

// resolveEntityID resolves a "schema.table" string or UUID to an entity UUID.
func (e *Executor) resolveEntityID(ctx context.Context, entity, token string) (string, error) {
	// If it's already a UUID, return as-is.
	if _, err := uuid.Parse(entity); err == nil {
		return entity, nil
	}

	// Validate format: must be "schema.table"
	if !strings.Contains(entity, ".") {
		return "", fmt.Errorf("entity must be in 'schema.table' format (e.g. 'inventory.inventory_items') or a UUID, got %q", entity)
	}

	e.cacheMu.Lock()
	defer e.cacheMu.Unlock()

	// Check cache.
	if e.entityCache != nil {
		if id, ok := e.entityCache[entity]; ok {
			return id, nil
		}
	}

	// Fetch all entities and populate cache.
	data, err := e.get(ctx, "/v1/workflow/entities", token)
	if err != nil {
		return "", fmt.Errorf("failed to fetch entities for name resolution: %w", err)
	}

	var entities []struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		SchemaName string `json:"schema_name"`
	}
	if err := json.Unmarshal(data, &entities); err != nil {
		return "", fmt.Errorf("failed to parse entities response: %w", err)
	}

	e.entityCache = make(map[string]string, len(entities))
	for _, ent := range entities {
		key := ent.SchemaName + "." + ent.Name
		e.entityCache[key] = ent.ID
	}

	if id, ok := e.entityCache[entity]; ok {
		return id, nil
	}
	return "", fmt.Errorf("entity %q not found — use discover_entities to list available entities", entity)
}

// resolveTriggerTypeID resolves a trigger type name or UUID to a trigger type UUID.
func (e *Executor) resolveTriggerTypeID(ctx context.Context, triggerType, token string) (string, error) {
	// If it's already a UUID, return as-is.
	if _, err := uuid.Parse(triggerType); err == nil {
		return triggerType, nil
	}

	e.cacheMu.Lock()
	defer e.cacheMu.Unlock()

	// Check cache.
	if e.triggerTypeCache != nil {
		if id, ok := e.triggerTypeCache[triggerType]; ok {
			return id, nil
		}
	}

	// Fetch all trigger types and populate cache.
	data, err := e.get(ctx, "/v1/workflow/trigger-types", token)
	if err != nil {
		return "", fmt.Errorf("failed to fetch trigger types for name resolution: %w", err)
	}

	var types []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(data, &types); err != nil {
		return "", fmt.Errorf("failed to parse trigger types response: %w", err)
	}

	e.triggerTypeCache = make(map[string]string, len(types))
	for _, t := range types {
		e.triggerTypeCache[t.Name] = t.ID
	}

	if id, ok := e.triggerTypeCache[triggerType]; ok {
		return id, nil
	}
	return "", fmt.Errorf("trigger type %q not found — use discover_trigger_types to list available types", triggerType)
}

// resolveRuleID resolves a rule name or UUID to a rule UUID.
func (e *Executor) resolveRuleID(ctx context.Context, ruleRef, token string) (string, error) {
	// If it's already a UUID, return as-is.
	if _, err := uuid.Parse(ruleRef); err == nil {
		return ruleRef, nil
	}

	e.cacheMu.Lock()
	defer e.cacheMu.Unlock()

	// Normalize: replace underscores/hyphens with spaces for flexible matching.
	normalized := strings.NewReplacer("_", " ", "-", " ").Replace(ruleRef)

	// Check cache.
	if e.ruleCache != nil {
		if id, ok := e.ruleCache[normalized]; ok {
			return id, nil
		}
	}

	// Fetch all rules and populate cache. The rules endpoint returns a
	// paginated response which unwrapPaginated strips to a bare array.
	data, err := e.get(ctx, "/v1/workflow/rules?rows=500", token)
	if err != nil {
		return "", fmt.Errorf("failed to fetch rules for name resolution: %w", err)
	}

	var rules []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(unwrapPaginated(data), &rules); err != nil {
		return "", fmt.Errorf("failed to parse rules response: %w", err)
	}

	e.ruleCache = make(map[string]string, len(rules))
	for _, r := range rules {
		e.ruleCache[r.Name] = r.ID
	}

	if id, ok := e.ruleCache[normalized]; ok {
		return id, nil
	}
	// Also try exact match against original input (in case the name has underscores).
	if id, ok := e.ruleCache[ruleRef]; ok {
		return id, nil
	}
	return "", fmt.Errorf("rule %q not found — use list_workflow_rules to see available rules", ruleRef)
}

// getDefaultOutputPort returns the default output port name for an action type.
func (e *Executor) getDefaultOutputPort(ctx context.Context, actionType, token string) (string, error) {
	e.cacheMu.Lock()
	defer e.cacheMu.Unlock()

	// Check cache.
	if e.actionTypeCache != nil {
		if info, ok := e.actionTypeCache[actionType]; ok {
			return info.DefaultPort, nil
		}
	}

	// Fetch all action types and populate cache.
	data, err := e.get(ctx, "/v1/workflow/action-types", token)
	if err != nil {
		return "", fmt.Errorf("failed to fetch action types for port resolution: %w", err)
	}

	var types []struct {
		Type        string `json:"type"`
		OutputPorts []struct {
			Name      string `json:"name"`
			IsDefault bool   `json:"is_default"`
		} `json:"output_ports"`
	}
	if err := json.Unmarshal(data, &types); err != nil {
		return "", fmt.Errorf("failed to parse action types response: %w", err)
	}

	e.actionTypeCache = make(map[string]actionTypeInfo, len(types))
	for _, t := range types {
		defaultPort := "success" // fallback
		for _, p := range t.OutputPorts {
			if p.IsDefault {
				defaultPort = p.Name
				break
			}
		}
		e.actionTypeCache[t.Type] = actionTypeInfo{DefaultPort: defaultPort}
	}

	if info, ok := e.actionTypeCache[actionType]; ok {
		return info.DefaultPort, nil
	}
	return "success", nil // safe fallback for unknown types
}

// =========================================================================
// Draft builder handlers
// =========================================================================

// handleStartDraft creates a new in-memory draft workflow.
func (e *Executor) handleStartDraft(ctx context.Context, tc llm.ToolCall, token string) (json.RawMessage, error) {
	var p struct {
		Name        string          `json:"name"`
		Entity      string          `json:"entity"`
		TriggerType string          `json:"trigger_type"`
		Description string          `json:"description"`
		TriggerCond json.RawMessage `json:"trigger_conditions"`
	}
	if err := json.Unmarshal(tc.Input, &p); err != nil {
		return nil, fmt.Errorf("bad params: %w", err)
	}
	if p.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if p.Entity == "" {
		return nil, fmt.Errorf("entity is required")
	}
	if p.TriggerType == "" {
		return nil, fmt.Errorf("trigger_type is required")
	}

	// Validate entity and trigger_type can be resolved (fail fast).
	if _, err := e.resolveEntityID(ctx, p.Entity, token); err != nil {
		return nil, err
	}
	if _, err := e.resolveTriggerTypeID(ctx, p.TriggerType, token); err != nil {
		return nil, err
	}

	draftID := uuid.New().String()

	e.draftMu.Lock()
	e.cleanExpiredDraftsLocked()
	if len(e.drafts) >= maxDrafts {
		e.draftMu.Unlock()
		return nil, fmt.Errorf("maximum draft limit reached (%d) — please preview or abandon existing drafts", maxDrafts)
	}
	e.drafts[draftID] = &draftWorkflow{
		lastAccess:  time.Now(),
		name:        p.Name,
		entity:      p.Entity,
		triggerType: p.TriggerType,
		description: p.Description,
		triggerCond: p.TriggerCond,
	}
	e.draftMu.Unlock()

	e.log.Info(ctx, "AGENT-CHAT: draft created", "draft_id", draftID, "name", p.Name)

	return json.Marshal(map[string]string{
		"draft_id": draftID,
		"status":   "draft_created",
		"message":  fmt.Sprintf("Draft %q created. Use add_draft_action to add actions, then preview_draft to validate and preview.", p.Name),
	})
}

// handleAddDraftAction appends an action to a draft workflow.
func (e *Executor) handleAddDraftAction(ctx context.Context, tc llm.ToolCall, token string) (json.RawMessage, error) {
	var p struct {
		DraftID    string          `json:"draft_id"`
		Name       string          `json:"name"`
		ActionType string          `json:"action_type"`
		Config     json.RawMessage `json:"action_config"`
		Desc       string          `json:"description"`
		IsActive   *bool           `json:"is_active"`
		After      string          `json:"after"`
	}
	if err := json.Unmarshal(tc.Input, &p); err != nil {
		return nil, fmt.Errorf("bad params: %w", err)
	}
	if p.DraftID == "" {
		return nil, fmt.Errorf("draft_id is required")
	}
	if p.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if p.ActionType == "" {
		return nil, fmt.Errorf("action_type is required")
	}

	isActive := true
	if p.IsActive != nil {
		isActive = *p.IsActive
	}

	e.draftMu.Lock()
	e.cleanExpiredDraftsLocked()
	draft, ok := e.drafts[p.DraftID]
	if !ok {
		e.draftMu.Unlock()
		return nil, fmt.Errorf("draft %q not found or expired", p.DraftID)
	}

	// Check for duplicate action name.
	for _, a := range draft.actions {
		if a.Name == p.Name {
			e.draftMu.Unlock()
			return nil, fmt.Errorf("action %q already exists in draft — use a different name or remove_draft_action first", p.Name)
		}
	}

	// Validate "after" reference if provided.
	if p.After != "" {
		refName, _, err := parseAfterField(p.After)
		if err != nil {
			e.draftMu.Unlock()
			return nil, fmt.Errorf("invalid 'after' format: %w", err)
		}
		found := false
		for _, existing := range draft.actions {
			if existing.Name == refName {
				found = true
				break
			}
		}
		if !found {
			e.draftMu.Unlock()
			return nil, fmt.Errorf("'after' references unknown action %q — add that action first or omit 'after' for the start action", refName)
		}
	}

	draft.actions = append(draft.actions, draftAction{
		Name:       p.Name,
		ActionType: p.ActionType,
		Config:     p.Config,
		Desc:       p.Desc,
		IsActive:   isActive,
		After:      p.After,
	})
	draft.lastAccess = time.Now()
	e.draftMu.Unlock()

	// Look up output ports for this action type so the LLM knows what's available.
	ports := e.getOutputPortNames(ctx, p.ActionType, token)

	return json.Marshal(map[string]any{
		"status":       "action_added",
		"action_name":  p.Name,
		"action_index": len(draft.actions) - 1,
		"output_ports": ports,
		"message":      fmt.Sprintf("Action %q (%s) added to draft. Available output ports: %s", p.Name, p.ActionType, strings.Join(ports, ", ")),
	})
}

// handleRemoveDraftAction removes an action from a draft by name.
func (e *Executor) handleRemoveDraftAction(_ context.Context, tc llm.ToolCall) (json.RawMessage, error) {
	var p struct {
		DraftID    string `json:"draft_id"`
		ActionName string `json:"action_name"`
	}
	if err := json.Unmarshal(tc.Input, &p); err != nil {
		return nil, fmt.Errorf("bad params: %w", err)
	}
	if p.DraftID == "" {
		return nil, fmt.Errorf("draft_id is required")
	}
	if p.ActionName == "" {
		return nil, fmt.Errorf("action_name is required")
	}

	e.draftMu.Lock()
	e.cleanExpiredDraftsLocked()
	draft, ok := e.drafts[p.DraftID]
	if !ok {
		e.draftMu.Unlock()
		return nil, fmt.Errorf("draft %q not found or expired", p.DraftID)
	}

	found := false
	filtered := make([]draftAction, 0, len(draft.actions))
	for _, a := range draft.actions {
		if a.Name == p.ActionName {
			found = true
			continue
		}
		// Clear "after" references to the removed action.
		if a.After != "" {
			refName, _, _ := parseAfterField(a.After)
			if refName == p.ActionName {
				a.After = ""
			}
		}
		filtered = append(filtered, a)
	}
	if !found {
		e.draftMu.Unlock()
		return nil, fmt.Errorf("action %q not found in draft", p.ActionName)
	}

	draft.actions = filtered
	draft.lastAccess = time.Now()
	e.draftMu.Unlock()

	return json.Marshal(map[string]string{
		"status":  "action_removed",
		"message": fmt.Sprintf("Action %q removed from draft. %d actions remaining.", p.ActionName, len(filtered)),
	})
}

// handlePreviewDraft assembles a draft into a complete workflow payload,
// transforms it (name resolution + edge generation), validates via dry-run,
// and returns the result. chatapi.go intercepts this to emit a preview SSE event.
func (e *Executor) handlePreviewDraft(ctx context.Context, tc llm.ToolCall, token string) (json.RawMessage, error) {
	var p struct {
		DraftID     string `json:"draft_id"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal(tc.Input, &p); err != nil {
		return nil, fmt.Errorf("bad params: %w", err)
	}
	if p.DraftID == "" {
		return nil, fmt.Errorf("draft_id is required")
	}
	if p.Description == "" {
		return nil, fmt.Errorf("description is required")
	}

	e.draftMu.Lock()
	e.cleanExpiredDraftsLocked()
	draft, ok := e.drafts[p.DraftID]
	if !ok {
		e.draftMu.Unlock()
		return nil, fmt.Errorf("draft %q not found or expired", p.DraftID)
	}
	if len(draft.actions) == 0 {
		e.draftMu.Unlock()
		return nil, fmt.Errorf("draft has no actions — use add_draft_action first")
	}
	// Copy draft data under lock, then release.
	draftCopy := *draft
	actionsCopy := make([]draftAction, len(draft.actions))
	copy(actionsCopy, draft.actions)
	draftCopy.actions = actionsCopy
	draft.lastAccess = time.Now()
	e.draftMu.Unlock()

	// Assemble into a full workflow payload.
	workflow := e.assembleDraftWorkflow(&draftCopy)

	// Transform: resolve names → UUIDs, generate edges from "after".
	transformed, err := e.transformWorkflowPayload(ctx, workflow, token)
	if err != nil {
		return nil, fmt.Errorf("transform draft: %w", err)
	}

	// Validate via dry-run.
	validationResult, err := e.post(ctx, "/v1/workflow/rules/full?dry_run=true", transformed, token)
	if err != nil {
		return nil, fmt.Errorf("validate draft: %w", err)
	}

	// Wrap result with the assembled workflow so chatapi can build the preview event.
	var validationMap map[string]json.RawMessage
	if err := json.Unmarshal(validationResult, &validationMap); err != nil {
		// If we can't parse, just return the raw validation result.
		return validationResult, nil
	}
	validationMap["workflow"] = transformed

	return json.Marshal(validationMap)
}

// assembleDraftWorkflow builds a workflow JSON payload from draft state.
func (e *Executor) assembleDraftWorkflow(draft *draftWorkflow) json.RawMessage {
	actions := make([]map[string]any, len(draft.actions))
	for i, a := range draft.actions {
		action := map[string]any{
			"name":          a.Name,
			"action_type":   a.ActionType,
			"action_config": a.Config,
			"is_active":     a.IsActive,
		}
		if a.Desc != "" {
			action["description"] = a.Desc
		}
		if a.After != "" {
			action["after"] = a.After
		}
		actions[i] = action
	}

	w := map[string]any{
		"name":      draft.name,
		"is_active": true,
		"entity":    draft.entity,
		"trigger_type": draft.triggerType,
		"actions":   actions,
		"edges":     []any{}, // Will be generated from "after" by transformWorkflowPayload
	}
	if draft.description != "" {
		w["description"] = draft.description
	}
	if len(draft.triggerCond) > 0 {
		w["trigger_conditions"] = draft.triggerCond
	}

	b, _ := json.Marshal(w)
	return b
}

// getOutputPortNames returns the port names for an action type (for LLM feedback).
func (e *Executor) getOutputPortNames(ctx context.Context, actionType, token string) []string {
	// Try to get from cached action type data.
	data, err := e.get(ctx, "/v1/workflow/action-types", token)
	if err != nil {
		return []string{"success", "failure"}
	}

	var types []struct {
		Type        string `json:"type"`
		OutputPorts []struct {
			Name string `json:"name"`
		} `json:"output_ports"`
	}
	if err := json.Unmarshal(data, &types); err != nil {
		return []string{"success", "failure"}
	}

	for _, t := range types {
		if t.Type == actionType {
			ports := make([]string, len(t.OutputPorts))
			for i, p := range t.OutputPorts {
				ports[i] = p.Name
			}
			if len(ports) > 0 {
				return ports
			}
			break
		}
	}
	return []string{"success", "failure"}
}

// cleanExpiredDraftsLocked removes drafts older than draftTTL.
// Must be called with draftMu held.
func (e *Executor) cleanExpiredDraftsLocked() {
	now := time.Now()
	for id, d := range e.drafts {
		if now.Sub(d.lastAccess) > draftTTL {
			delete(e.drafts, id)
		}
	}
}

// =========================================================================
// Workflow summary and node explanation
// =========================================================================

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
		"workflow_id": ruleID,
		"rule":    json.RawMessage(rule),
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

	// Check if the identifier matches multiple nodes by action_type
	// (e.g. "create_alert" matches 3 alert actions). If so, return all.
	typeMatches := graph.findActionsByType(identifier)
	if len(typeMatches) > 1 {
		explanations := make([]any, 0, len(typeMatches))
		for _, action := range typeMatches {
			if action.ActionType == "create_alert" && len(action.Config) > 0 {
				action.Config = e.enrichCreateAlertConfig(ctx, action.Config, token)
			}
			explanations = append(explanations, graph.explainNode(action))
		}
		result := map[string]any{
			"matched_by":   "action_type",
			"match_count":  len(typeMatches),
			"message":      fmt.Sprintf("Found %d actions of type %q in this workflow", len(typeMatches), identifier),
			"explanations": explanations,
		}
		b, err := json.Marshal(result)
		if err != nil {
			return nil, fmt.Errorf("marshal multi-node explanation: %w", err)
		}
		return b, nil
	}

	// Single match (by name, ID, or single action_type match).
	action := graph.findAction(identifier)
	if action == nil {
		// Build helpful error listing available action names.
		names := make([]string, len(graph.actions))
		for i, a := range graph.actions {
			names[i] = a.Name
		}
		return nil, fmt.Errorf("action %q not found in this workflow — available actions: %s", identifier, strings.Join(names, ", "))
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

// =========================================================================
// Alert helpers
// =========================================================================

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

	if p.Rows == "" {
		p.Rows = "50"
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
	path += "rows=" + p.Rows + "&"

	data, err := e.get(ctx, path, token)
	if err != nil {
		return nil, err
	}
	return summarizeAlertList(formatPaginatedResponse(data)), nil
}

// listAlertsForRule fetches alerts fired by a specific workflow rule.
func (e *Executor) listAlertsForRule(ctx context.Context, tc llm.ToolCall, token string) (json.RawMessage, error) {
	var p struct {
		RuleID string `json:"workflow_id"`
		Status string `json:"status"`
		Page   string `json:"page"`
		Rows   string `json:"rows"`
	}
	if err := json.Unmarshal(tc.Input, &p); err != nil {
		return nil, fmt.Errorf("bad params: %w", err)
	}
	ruleID, err := e.resolveRuleID(ctx, p.RuleID, token)
	if err != nil {
		return nil, err
	}

	if p.Rows == "" {
		p.Rows = "50"
	}

	path := "/v1/workflow/alerts?sourceRuleId=" + ruleID
	if p.Status != "" {
		path += "&status=" + p.Status
	}
	if p.Page != "" {
		path += "&page=" + p.Page
	}
	path += "&rows=" + p.Rows

	data, err := e.get(ctx, path, token)
	if err != nil {
		return nil, err
	}
	return summarizeAlertList(formatPaginatedResponse(data)), nil
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

// =========================================================================
// Utilities
// =========================================================================

// unwrapPaginated extracts the "items" array from a paginated API response
// of the form {"items": [...], "total": N, "page": N, "rows_per_page": N}.
// If the response is already a bare array or doesn't have an "items" key,
// the original data is returned unchanged.
// Used internally (e.g. resolveRuleID) where a bare array is needed.
func unwrapPaginated(data json.RawMessage) json.RawMessage {
	// Quick check: if it starts with '[', it's already a bare array.
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) > 0 && trimmed[0] == '[' {
		return data
	}

	var wrapper struct {
		Items json.RawMessage `json:"items"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil || wrapper.Items == nil {
		return data
	}
	return wrapper.Items
}

// formatPaginatedResponse transforms a paginated API response into an
// LLM-friendly format that preserves pagination metadata. Returns:
//
//	{"results": [...], "total": N, "showing": N, "has_more": bool}
//
// If the response is already a bare array or non-paginated, returns it unchanged.
func formatPaginatedResponse(data json.RawMessage) json.RawMessage {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) > 0 && trimmed[0] == '[' {
		return data
	}

	var wrapper struct {
		Items       json.RawMessage `json:"items"`
		Total       int             `json:"total"`
		Page        int             `json:"page"`
		RowsPerPage int             `json:"rowsPerPage"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil || wrapper.Items == nil {
		return data
	}

	// Count items in the array.
	var items []json.RawMessage
	if err := json.Unmarshal(wrapper.Items, &items); err != nil {
		return data
	}

	showing := len(items)
	hasMore := showing < wrapper.Total

	result, err := json.Marshal(map[string]any{
		"results":  wrapper.Items,
		"total":    wrapper.Total,
		"showing":  showing,
		"has_more": hasMore,
	})
	if err != nil {
		return data
	}
	return result
}

// =========================================================================
// Result summarization (reduce token usage)
// =========================================================================

// summarizeActionTypes strips full config schemas from action type results,
// keeping only the type name, description (first sentence), output port names,
// and config field names.
func summarizeActionTypes(data json.RawMessage) json.RawMessage {
	var types []struct {
		Type        string `json:"type"`
		Description string `json:"description"`
		OutputPorts []struct {
			Name      string `json:"name"`
			IsDefault bool   `json:"is_default"`
		} `json:"output_ports"`
		ConfigSchema json.RawMessage `json:"config_schema"`
	}
	if err := json.Unmarshal(data, &types); err != nil {
		return data
	}

	summaries := make([]map[string]any, len(types))
	for i, t := range types {
		// Truncate description to first sentence.
		desc := t.Description
		if idx := strings.Index(desc, ". "); idx > 0 {
			desc = desc[:idx+1]
		} else if len(desc) > 120 {
			desc = desc[:120] + "…"
		}

		// Extract just port names.
		ports := make([]string, len(t.OutputPorts))
		for j, p := range t.OutputPorts {
			label := p.Name
			if p.IsDefault {
				label += " (default)"
			}
			ports[j] = label
		}

		// Extract just config field names from JSON Schema.
		var configFields []string
		if len(t.ConfigSchema) > 0 {
			var s struct {
				Properties map[string]json.RawMessage `json:"properties"`
			}
			if json.Unmarshal(t.ConfigSchema, &s) == nil {
				configFields = make([]string, 0, len(s.Properties))
				for k := range s.Properties {
					configFields = append(configFields, k)
				}
			}
		}

		entry := map[string]any{
			"type":         t.Type,
			"description":  desc,
			"output_ports": ports,
		}
		if len(configFields) > 0 {
			entry["config_fields"] = configFields
		}
		summaries[i] = entry
	}

	b, err := json.Marshal(summaries)
	if err != nil {
		return data
	}
	return b
}

// summarizeRuleList strips rule objects to essential fields: id, name, entity,
// trigger_type, is_active. Preserves pagination metadata.
func summarizeRuleList(data json.RawMessage) json.RawMessage {
	var envelope struct {
		Results json.RawMessage `json:"results"`
		Total   int             `json:"total"`
		Showing int             `json:"showing"`
		HasMore bool            `json:"has_more"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil || envelope.Results == nil {
		return data
	}

	var rules []struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		EntityName  string `json:"entity_name"`
		TriggerType string `json:"trigger_type"`
		IsActive    bool   `json:"is_active"`
	}
	if err := json.Unmarshal(envelope.Results, &rules); err != nil {
		return data
	}

	summaries := make([]map[string]any, len(rules))
	for i, r := range rules {
		summaries[i] = map[string]any{
			"id":           r.ID,
			"name":         r.Name,
			"entity_name":  r.EntityName,
			"trigger_type": r.TriggerType,
			"is_active":    r.IsActive,
		}
	}

	result, err := json.Marshal(map[string]any{
		"results":  summaries,
		"total":    envelope.Total,
		"showing":  envelope.Showing,
		"has_more": envelope.HasMore,
	})
	if err != nil {
		return data
	}
	return result
}

// summarizeAlertList strips alert objects to essential fields: id, message,
// severity, status, created_at, and recipient names. Preserves pagination.
func summarizeAlertList(data json.RawMessage) json.RawMessage {
	var envelope struct {
		Results json.RawMessage `json:"results"`
		Total   int             `json:"total"`
		Showing int             `json:"showing"`
		HasMore bool            `json:"has_more"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil || envelope.Results == nil {
		return data
	}

	var alerts []struct {
		ID        string `json:"id"`
		Title     string `json:"title"`
		Message   string `json:"message"`
		Severity  string `json:"severity"`
		Status    string `json:"status"`
		CreatedAt string `json:"created_at"`
		Recipients struct {
			Users []struct {
				Name string `json:"name"`
			} `json:"users"`
			Roles []struct {
				Name string `json:"name"`
			} `json:"roles"`
		} `json:"recipients"`
	}
	if err := json.Unmarshal(envelope.Results, &alerts); err != nil {
		return data
	}

	summaries := make([]map[string]any, len(alerts))
	for i, a := range alerts {
		// Collect recipient names.
		var recipientNames []string
		for _, u := range a.Recipients.Users {
			if u.Name != "" {
				recipientNames = append(recipientNames, u.Name)
			}
		}
		for _, r := range a.Recipients.Roles {
			if r.Name != "" {
				recipientNames = append(recipientNames, "role:"+r.Name)
			}
		}

		entry := map[string]any{
			"id":       a.ID,
			"severity": a.Severity,
			"status":   a.Status,
		}
		if a.Title != "" {
			entry["title"] = a.Title
		}
		if a.Message != "" {
			entry["message"] = a.Message
		}
		if a.CreatedAt != "" {
			entry["created_at"] = a.CreatedAt
		}
		if len(recipientNames) > 0 {
			entry["recipients"] = recipientNames
		}
		summaries[i] = entry
	}

	result, err := json.Marshal(map[string]any{
		"results":  summaries,
		"total":    envelope.Total,
		"showing":  envelope.Showing,
		"has_more": envelope.HasMore,
	})
	if err != nil {
		return data
	}
	return result
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

// mustMarshal marshals v to JSON, panicking on error (should never fail for
// simple types like strings).
func mustMarshal(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic("mustMarshal: " + err.Error())
	}
	return b
}
