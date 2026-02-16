// Package chatapi provides the POST /v1/agent/chat SSE endpoint that lets
// an LLM agent interact with the Ichor workflow system on behalf of an
// authenticated user.
//
// # Architecture Note: SSE Route Registration
//
// This endpoint MUST use app.RawHandlerFunc() instead of app.HandlerFunc().
// Reasons:
//  1. SSE streams exceed the HTTP server's WriteTimeout (10 s).
//  2. The OTEL wrapper writes 200 OK after the handler returns, which
//     conflicts with the long-lived SSE connection.
//  3. RawHandlerFunc registers on a separate mux that bypasses OTEL
//     wrapping, identical to how WebSocket routes work.
package chatapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/timmaaaz/ichor/app/sdk/mid"
	"github.com/timmaaaz/ichor/business/sdk/agenttools"
	"github.com/timmaaaz/ichor/business/sdk/llm"
	"github.com/timmaaaz/ichor/business/sdk/toolcatalog"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// maxAgentLoops caps the number of LLM ↔ tool round-trips per request
// to prevent runaway loops.
const maxAgentLoops = 20

type api struct {
	log      *logger.Logger
	provider llm.Provider
	tools    []llm.ToolDef
	executor *agenttools.Executor
}

func newAPI(cfg Config) *api {
	return &api{
		log:      cfg.Log,
		provider: cfg.LLMProvider,
		tools:    agenttools.ToolDefinitions(),
		executor: cfg.ToolExecutor,
	}
}

// chat handles POST /v1/agent/chat.
// It is an http.HandlerFunc for use with app.RawHandlerFunc().
func (a *api) chat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Clear the HTTP server's read/write deadlines for this SSE connection
	// so the long-lived stream isn't killed by the 5 s ReadTimeout or 10 s
	// WriteTimeout.
	rc := http.NewResponseController(w)
	if err := rc.SetReadDeadline(time.Time{}); err != nil {
		a.log.Error(ctx, "AGENT-CHAT: failed to clear read deadline", "error", err)
	}
	if err := rc.SetWriteDeadline(time.Time{}); err != nil {
		a.log.Error(ctx, "AGENT-CHAT: failed to clear write deadline", "error", err)
	}

	// Detach from the request context's cancellation. If SetReadDeadline
	// failed above, r.Context() will fire after ReadTimeout (5 s), killing
	// the LLM request mid-inference. WithoutCancel preserves context values
	// (user ID, trace ID) without inheriting the deadline/cancellation.
	//
	// We manually cancel sseCtx when the handler returns (deferred) or when
	// we detect the client has disconnected (via the CloseNotify-style
	// goroutine below).
	sseCtx, sseCancel := context.WithCancel(context.WithoutCancel(ctx))
	defer sseCancel()

	// Monitor the original request context for client disconnect. When the
	// client drops the connection, r.Context() fires and we cancel sseCtx
	// so the LLM request + tool calls stop promptly.
	go func() {
		select {
		case <-ctx.Done():
			sseCancel()
		case <-sseCtx.Done():
			// Handler finished normally.
		}
	}()

	ctx = sseCtx

	// Extract user ID (set by mid.Authenticate).
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		a.log.Error(ctx, "AGENT-CHAT: failed to get user ID", "error", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Decode and validate request.
	var req ChatRequest
	if err := web.Decode(r, &req); err != nil {
		a.log.Error(ctx, "AGENT-CHAT: bad request body", "error", err)
		http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Authorization header is forwarded to tool calls verbatim.
	authToken := r.Header.Get("Authorization")

	a.log.Info(ctx, "AGENT-CHAT: new session",
		"user_id", userID,
		"context_type", req.ContextType)

	// Prepare SSE writer.
	sse := newSSEWriter(w)
	if sse == nil {
		a.log.Error(ctx, "AGENT-CHAT: flushing not supported")
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Filter tools by context type.
	filteredTools := filterToolsByContext(a.tools, req.ContextType)

	// Extract the workflow_id from the context (if present) so we can inject
	// it into tool calls where the LLM omits it.
	contextWorkflowID := extractContextWorkflowID(req.Context)

	// Build initial LLM request.
	llmReq := llm.ChatRequest{
		SystemPrompt: buildSystemPrompt(req.ContextType, req.Context),
		Messages: []llm.Message{
			{Role: "user", Content: req.Message},
		},
		Tools:     filteredTools,
		MaxTokens: 4096,
	}

	// =====================================================================
	// Agentic loop: stream LLM → execute tools → feed results → repeat
	// =====================================================================

	for turn := 0; turn < maxAgentLoops; turn++ {
		eventCh, err := a.provider.StreamChat(ctx, llmReq)
		if err != nil {
			a.log.Error(ctx, "AGENT-CHAT: provider error", "error", err)
			sse.sendError("LLM provider error: " + err.Error())
			return
		}

		// Accumulate assistant text + tool calls from this turn.
		var (
			assistantText  string
			thinkingText   string
			toolCalls      []llm.ToolCall
			currentToolID  string
			currentToolName string
			currentToolJSON string
			stopForTools   bool
		)

		for ev := range eventCh {
			if ctx.Err() != nil {
				return // client disconnected
			}

			switch ev.Type {
			case llm.EventMessageStart:
				sse.send("message_start", map[string]any{
					"turn": turn,
				})

			case llm.EventThinkingDelta:
				thinkingText += ev.ThinkingText
				// Thinking content is logged server-side only, not sent to client.

			case llm.EventContentDelta:
				assistantText += ev.Text
				sse.send("content_chunk", map[string]string{
					"chunk": ev.Text,
				})

			case llm.EventToolUseStart:
				currentToolID = ev.ToolCallID
				currentToolName = ev.ToolCallName
				currentToolJSON = ""
				sse.send("tool_call_start", map[string]string{
					"tool_use_id": ev.ToolCallID,
					"name":        ev.ToolCallName,
				})

			case llm.EventToolUseInput:
				currentToolJSON += ev.PartialInput

			case llm.EventMessageComplete:
				stopForTools = ev.StopForToolUse
				// Finalize any pending tool call.
				if currentToolID != "" {
					toolCalls = append(toolCalls, llm.ToolCall{
						ID:    currentToolID,
						Name:  currentToolName,
						Input: json.RawMessage(currentToolJSON),
					})
					currentToolID = ""
				}

			case llm.EventError:
				a.log.Error(ctx, "AGENT-CHAT: stream error", "error", ev.Err)
				sse.sendError(ev.Err.Error())
				return
			}
		}

		// Log thinking content (server-side only).
		if thinkingText != "" {
			a.log.Info(ctx, "AGENT-CHAT: thinking",
				"turn", turn,
				"content", truncateLog(thinkingText, 2000))
		}

		// Log assistant response.
		if assistantText != "" {
			a.log.Info(ctx, "AGENT-CHAT: assistant response",
				"turn", turn,
				"length", len(assistantText),
				"text", truncateLog(assistantText, 2000))
		}

		// If the LLM didn't request tools, check for Chinese and possibly retry.
		if !stopForTools || len(toolCalls) == 0 {
			// If the response contains Chinese, inject a correction message
			// and continue the loop so the LLM can retry in English.
			if containsChinese(assistantText) {
				a.log.Warn(ctx, "AGENT-CHAT: detected Chinese in response, requesting English retry",
					"turn", turn)

				// Build assistant message with the Chinese response.
				llmReq.Messages = append(llmReq.Messages, llm.Message{
					Role:    "assistant",
					Content: assistantText,
				})

				// Inject a user message asking for English.
				llmReq.Messages = append(llmReq.Messages, llm.Message{
					Role:    "user",
					Content: "Please respond in English, not Chinese. Repeat your previous response in English.",
				})

				sse.send("content_chunk", map[string]string{
					"chunk": "\n\n[Retrying in English...]\n\n",
				})
				continue
			}

			sse.send("message_complete", nil)
			a.log.Info(ctx, "AGENT-CHAT: session complete",
				"user_id", userID,
				"turns", turn+1)
			return
		}

		// Build the assistant message for conversation history.
		assistantMsg := llm.Message{
			Role:      "assistant",
			Content:   assistantText,
			ToolCalls: toolCalls,
		}
		llmReq.Messages = append(llmReq.Messages, assistantMsg)

		// Execute all tool calls and collect results.
		a.log.Info(ctx, "AGENT-CHAT: executing tools",
			"count", len(toolCalls),
			"turn", turn)

		toolResults := make([]llm.ToolResult, 0, len(toolCalls))
		for _, tc := range toolCalls {
			// Inject the context workflow_id into tool calls that accept it
			// but where the LLM omitted it.
			tc.Input = injectWorkflowID(tc, contextWorkflowID)
			a.log.Info(ctx, "AGENT-CHAT: tool input",
				"name", tc.Name,
				"tool_use_id", tc.ID,
				"input", truncateLog(string(tc.Input), 2000))

			toolStart := time.Now()
			result := a.executor.Execute(ctx, tc, authToken)
			a.log.Info(ctx, "AGENT-CHAT: tool executed",
				"name", tc.Name,
				"tool_use_id", tc.ID,
				"elapsed", time.Since(toolStart),
				"is_error", result.IsError,
				"result", truncateLog(result.Content, 2000))

			// Intercept preview_workflow / preview_draft: if validation passed,
			// emit a workflow_preview SSE event so the frontend can show the
			// proposed state for user approval.
			if (tc.Name == "preview_workflow" || tc.Name == "preview_draft") && !result.IsError {
				if preview := buildPreviewEvent(tc, result); preview != nil {
					sse.send("workflow_preview", preview)
					result.Content = `{"status":"preview_sent","message":"Preview sent to user for approval. Do not call create_workflow or update_workflow. The user will accept or reject the preview directly."}`
				}
			}

			toolResults = append(toolResults, result)

			sse.send("tool_call_result", map[string]any{
				"tool_use_id": result.ToolUseID,
				"name":        tc.Name,
				"is_error":    result.IsError,
			})
		}

		// Append tool results as a user message.
		llmReq.Messages = append(llmReq.Messages, llm.Message{
			Role:        "user",
			ToolResults: toolResults,
		})
	}

	// Safety: max loops reached.
	a.log.Warn(ctx, "AGENT-CHAT: max loops reached", "user_id", userID)
	sse.sendError("Maximum tool-call rounds exceeded")
}

// filterToolsByContext returns the subset of tools that belong to the given
// context type. An empty or unrecognised context type returns all tools.
func filterToolsByContext(tools []llm.ToolDef, contextType string) []llm.ToolDef {
	var group toolcatalog.ToolGroup
	switch contextType {
	case "workflow":
		group = toolcatalog.GroupWorkflow
	case "tables":
		group = toolcatalog.GroupTables
	default:
		return tools
	}

	filtered := make([]llm.ToolDef, 0, len(tools))
	for _, t := range tools {
		if toolcatalog.InGroup(t.Name, group) {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

// buildPreviewEvent constructs the workflow_preview SSE event payload from a
// successful preview_workflow or preview_draft tool call. Returns nil if the
// validation failed (the LLM should see the errors and retry).
func buildPreviewEvent(tc llm.ToolCall, result llm.ToolResult) map[string]any {
	// Only emit preview when validation passed.
	var validation struct {
		Valid    bool            `json:"valid"`
		Workflow json.RawMessage `json:"workflow"` // present in preview_draft responses
	}
	if err := json.Unmarshal([]byte(result.Content), &validation); err != nil || !validation.Valid {
		return nil
	}

	// Extract metadata from the tool input.
	var input struct {
		RuleID      string          `json:"workflow_id"`
		DraftID     string          `json:"draft_id"`
		Workflow    json.RawMessage `json:"workflow"`
		Description string          `json:"description"`
	}
	if err := json.Unmarshal(tc.Input, &input); err != nil {
		return nil
	}

	// For preview_draft, the assembled workflow comes from the validation
	// response (the executor embeds it after transforming the draft).
	// For preview_workflow, the workflow comes from the tool input.
	workflow := input.Workflow
	if input.DraftID != "" && len(validation.Workflow) > 0 {
		workflow = validation.Workflow
	}

	// Guard against malformed workflow JSON reaching the frontend.
	if !json.Valid(workflow) {
		return nil
	}

	event := map[string]any{
		"description": input.Description,
		"workflow":    json.RawMessage(workflow),
		"is_update":  input.RuleID != "",
	}
	if input.RuleID != "" {
		event["workflow_id"] = input.RuleID
	}

	return event
}

// truncateLog returns s truncated to maxLen characters with a suffix
// indicating the original length. Used for debug logging large payloads.
func truncateLog(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + fmt.Sprintf("... [truncated, total %d chars]", len(s))
}

// extractContextWorkflowID parses the request context JSON and returns the
// workflow_id field, or empty string if not present.
func extractContextWorkflowID(rawCtx json.RawMessage) string {
	if len(rawCtx) == 0 {
		return ""
	}
	var ctx struct {
		WorkflowID string `json:"workflow_id"`
	}
	if err := json.Unmarshal(rawCtx, &ctx); err != nil {
		return ""
	}
	return ctx.WorkflowID
}

// toolsNeedingWorkflowID lists tools where workflow_id can be auto-filled
// from the current workflow context.
var toolsNeedingWorkflowID = map[string]bool{
	"explain_workflow_node": true,
	"get_workflow_rule":     true,
	"list_alerts_for_rule":  true,
}

// injectWorkflowID adds or overrides workflow_id in the tool call input when
// we have a context value. If the LLM provided a valid UUID, we trust it
// (it might be referencing a different workflow). Otherwise we inject the
// context workflow_id — this catches both missing values and fabricated names
// like "Granular_Inventory_Pipeline" that won't resolve.
func injectWorkflowID(tc llm.ToolCall, contextID string) json.RawMessage {
	if contextID == "" || !toolsNeedingWorkflowID[tc.Name] {
		return tc.Input
	}

	var input map[string]json.RawMessage
	if err := json.Unmarshal(tc.Input, &input); err != nil {
		return tc.Input
	}

	// If the LLM provided a valid UUID, keep it (might be a different workflow).
	if raw, ok := input["workflow_id"]; ok {
		var val string
		if json.Unmarshal(raw, &val) == nil && isUUID(val) {
			return tc.Input
		}
	}

	// Missing, empty, or non-UUID → inject the context workflow_id.
	input["workflow_id"] = json.RawMessage(`"` + contextID + `"`)
	patched, err := json.Marshal(input)
	if err != nil {
		return tc.Input
	}
	return patched
}

// isUUID checks if a string looks like a valid UUID (8-4-4-4-12 hex format).
func isUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, c := range s {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return false
			}
		} else if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// containsChinese returns true if the string contains any CJK Unified
// Ideographs (Chinese characters). Used to detect when the LLM responds in
// Chinese so we can request a retry in English.
func containsChinese(s string) bool {
	for _, r := range s {
		// CJK Unified Ideographs: U+4E00 to U+9FFF
		if r >= 0x4E00 && r <= 0x9FFF {
			return true
		}
	}
	return false
}
