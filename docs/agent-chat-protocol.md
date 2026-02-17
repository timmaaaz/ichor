# Agent Chat Protocol

This document describes the data packets exchanged in the Agent Chat system — the HTTP request/response format, SSE event stream, LLM provider interface, tool call lifecycle, and agentic loop mechanics.

## Overview

```
┌──────────┐   POST /v1/agent/chat   ┌──────────┐   StreamChat()   ┌──────────────┐
│  Browser  │ ──────────────────────→ │ chatapi  │ ──────────────→ │ LLM Provider │
│ (SSE)     │ ←── SSE event stream ── │ handler  │ ←── chan<Event> │ (Gemini/etc)  │
└──────────┘                          └────┬─────┘                 └──────────────┘
                                           │
                                           │ Execute()
                                           ▼
                                      ┌──────────┐   HTTP   ┌──────────┐
                                      │ agenttools│ ──────→ │ Ichor    │
                                      │ executor  │ ←────── │ REST API │
                                      └──────────┘          └──────────┘
```

The handler runs an **agentic loop** (up to 20 turns): stream LLM output to the client, execute any tool calls the LLM requests, feed tool results back to the LLM, repeat until the LLM produces a final text response.

---

## 1. HTTP Request

**Endpoint:** `POST /v1/agent/chat`
**Auth:** `Authorization: Bearer <JWT>` (forwarded verbatim to tool calls)
**Content-Type:** `application/json`
**Response:** `text/event-stream` (SSE)

### Request Body

```json
{
  "message": "What does this workflow do?",
  "context_type": "workflow",
  "context": { ... },
  "conversation_id": "optional-tracking-id"
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `message` | string | yes | The user's message |
| `context_type` | string | yes | `"workflow"` or `"tables"` — determines which tools and system prompt are used |
| `context` | object | no | Current UI state (workflow nodes/edges, rule metadata). Injected into the system prompt |
| `conversation_id` | string | no | Client-provided conversation tracking ID |

### Context Object (workflow)

When `context_type` is `"workflow"`, the context typically contains:

```json
{
  "workflow_id": "550e8400-e29b-41d4-a716-446655440000",
  "rule_name": "Low Stock Alert",
  "entity_schema": "inventory",
  "entity_name": "inventory_items",
  "trigger_type": "on_update",
  "nodes": [ ... ],
  "edges": [ ... ],
  "summary": "3 actions, 1 branch"
}
```

The `workflow_id` is extracted and auto-injected into tool calls where the LLM omits it (e.g. `explain_workflow_node`, `get_workflow_rule`, `list_alerts_for_rule`).

---

## 2. SSE Event Stream (Server → Client)

All events follow standard SSE format:
```
event: <event_type>
data: <json_payload>

```

### Event Types

#### `message_start`
Emitted at the beginning of each LLM turn.

```json
{"turn": 0}
```

| Field | Type | Description |
|---|---|---|
| `turn` | int | Zero-based turn index in the agentic loop |

#### `content_chunk`
Streamed text fragment from the LLM. Concatenate all chunks to form the full response.

```json
{"chunk": "The **Low Stock Alert** rule triggers when"}
```

| Field | Type | Description |
|---|---|---|
| `chunk` | string | Text fragment (may contain markdown) |

#### `tool_call_start`
The LLM is requesting a tool invocation.

```json
{"tool_use_id": "toolu_abc123", "name": "get_workflow_rule"}
```

| Field | Type | Description |
|---|---|---|
| `tool_use_id` | string | Unique ID for this tool invocation (from the LLM provider) |
| `name` | string | Tool name (see [Tool Catalog](#5-tool-catalog)) |

**Note:** The tool's JSON input is streamed incrementally from the LLM but is NOT sent to the client. Only the tool name and ID are emitted.

#### `tool_call_result`
A tool has finished executing.

```json
{"tool_use_id": "toolu_abc123", "name": "get_workflow_rule", "is_error": false}
```

| Field | Type | Description |
|---|---|---|
| `tool_use_id` | string | Matches the `tool_call_start` ID |
| `name` | string | Tool name |
| `is_error` | bool | Whether the tool returned an error |

**Note:** The actual tool result content is NOT sent to the client — it's fed back to the LLM internally. The client only knows a tool ran and whether it succeeded.

#### `workflow_preview`
A `preview_workflow` or `preview_draft` tool call succeeded validation. The client should render a visual preview for user approval.

```json
{
  "description": "Add low stock alert when inventory drops below threshold",
  "workflow": { "name": "Low Stock Alert", "actions": [...], "edges": [...] },
  "is_update": false,
  "workflow_id": "550e8400-..."
}
```

| Field | Type | Description |
|---|---|---|
| `description` | string | Human-readable description of changes (from the LLM) |
| `workflow` | object | The full workflow payload that was validated |
| `is_update` | bool | `true` if updating an existing workflow, `false` if creating new |
| `workflow_id` | string | Present only for updates — the existing workflow's UUID |

After this event, the LLM receives a synthetic result: `{"status":"preview_sent","message":"Preview sent to user for approval..."}`. The LLM should NOT follow up with `create_workflow` or `update_workflow`.

#### `error`
An unrecoverable error occurred.

```json
{"message": "LLM provider error: context deadline exceeded"}
```

| Field | Type | Description |
|---|---|---|
| `message` | string | Error description |

#### `message_complete`
The LLM finished its final response. No more events will follow.

```json
{}
```

### Full Event Sequence Examples

**Simple question (no tools):**
```
event: message_start      → {"turn": 0}
event: content_chunk      → {"chunk": "The "}
event: content_chunk      → {"chunk": "**Low Stock Alert**"}
event: content_chunk      → {"chunk": " rule triggers..."}
event: message_complete   → {}
```

**Question requiring a tool call (2 turns):**
```
event: message_start      → {"turn": 0}
event: content_chunk      → {"chunk": "Let me look up that workflow."}
event: tool_call_start    → {"tool_use_id": "t1", "name": "get_workflow_rule"}
event: tool_call_result   → {"tool_use_id": "t1", "name": "get_workflow_rule", "is_error": false}
event: message_start      → {"turn": 1}
event: content_chunk      → {"chunk": "The workflow has 3 actions..."}
event: message_complete   → {}
```

**Workflow creation with preview (3 turns):**
```
event: message_start      → {"turn": 0}
event: tool_call_start    → {"tool_use_id": "t1", "name": "discover_action_types"}
event: tool_call_result   → {"tool_use_id": "t1", ...}
event: message_start      → {"turn": 1}
event: content_chunk      → {"chunk": "I'll build that workflow..."}
event: tool_call_start    → {"tool_use_id": "t2", "name": "preview_workflow"}
event: tool_call_result   → {"tool_use_id": "t2", ...}
event: workflow_preview   → {"description": "...", "workflow": {...}, "is_update": false}
event: message_start      → {"turn": 2}
event: content_chunk      → {"chunk": "I've sent a preview for your review."}
event: message_complete   → {}
```

---

## 3. Agentic Loop Internals

The handler maintains a conversation history (`[]llm.Message`) that grows across turns:

```
Turn 0:
  → SystemPrompt + [user message]          (sent to LLM)
  ← assistant text + tool_calls            (streamed to client)
  → [tool results]                         (fed back to LLM)

Turn 1:
  → SystemPrompt + [user, assistant+tools, tool_results]
  ← assistant text (final)                 (streamed to client)
  → message_complete
```

### Message Types in Conversation History

```go
// User message (initial)
{Role: "user", Content: "What does this workflow do?"}

// Assistant message (with tool requests)
{Role: "assistant", Content: "Let me look that up.", ToolCalls: [{ID, Name, Input}]}

// Tool results (as a user message)
{Role: "user", ToolResults: [{ToolUseID, Content, IsError}]}

// Assistant message (final, no tools)
{Role: "assistant", Content: "The workflow has 3 actions..."}
```

### Chinese Response Retry

If the LLM responds in Chinese (detected by CJK character scan), the handler injects a retry message and continues the loop:
```
← assistant: "这个工作流程有三个步骤..."
→ user: "Please respond in English, not Chinese."
← assistant: "This workflow has three steps..."
```

The client sees `[Retrying in English...]` as a `content_chunk`.

### Auto-Injected workflow_id

For tools `explain_workflow_node`, `get_workflow_rule`, and `list_alerts_for_rule`: if the LLM omits `workflow_id` or provides a non-UUID value (like a rule name), the handler injects the `workflow_id` from the request context. Valid UUIDs provided by the LLM are preserved (the LLM might be referencing a different workflow).

---

## 4. LLM Provider Interface

The `llm.Provider` interface abstracts all LLM vendors:

```go
type Provider interface {
    StreamChat(ctx context.Context, req ChatRequest) (<-chan StreamEvent, error)
}
```

### ChatRequest (handler → provider)

```go
type ChatRequest struct {
    SystemPrompt string      // Assembled by buildSystemPrompt()
    Messages     []Message   // Conversation history (grows across turns)
    Tools        []ToolDef   // Filtered by context_type
    MaxTokens    int         // 4096
}
```

### StreamEvent (provider → handler)

Events flow through a channel. The handler drains it to build the SSE stream.

| EventType | Fields Used | Description |
|---|---|---|
| `message_start` | — | LLM started generating |
| `content_delta` | `Text` | Text fragment |
| `thinking_delta` | `ThinkingText` | Reasoning (logged server-side, not streamed) |
| `tool_use_start` | `ToolCallID`, `ToolCallName` | LLM wants to call a tool |
| `tool_use_input` | `PartialInput` | Incremental JSON for tool input |
| `message_complete` | `StopForToolUse` | LLM finished; if `StopForToolUse=true`, execute tools and loop |
| `error` | `Err` | Provider error |

### Available Providers

| Provider | Package | Notes |
|---|---|---|
| Gemini | `llm/gemini` | Default provider |
| Ollama | `llm/ollama` | Local models (qwen3:14b) |
| Claude | `llm/claude` | Anthropic API |
| OpenAI-compatible | `llm/openaicompat` | Any OpenAI-compatible API |

---

## 5. Tool Catalog

Tools are defined in `business/sdk/agenttools/definitions.go` and filtered by `context_type` using `business/sdk/toolcatalog/`.

### Tool Definition Format

```go
type ToolDef struct {
    Name        string          // e.g. "get_workflow_rule"
    Description string          // Sent to LLM as tool description
    InputSchema json.RawMessage // JSON Schema for the tool's parameters
}
```

### Workflow Tools (context_type="workflow")

| Tool | Category | Description |
|---|---|---|
| `discover_action_types` | Discovery | List action types with config field names and output ports (summarized) |
| `discover_trigger_types` | Discovery | List trigger types (on_create, on_update, on_delete, manual) |
| `discover_entities` | Discovery | List entity types (schema.table pairs) |
| `get_workflow_rule` | Read | Fetch a rule by name/ID — returns compact summary with flow outline |
| `explain_workflow_node` | Read | Get full detail on a specific action (config, edges, depth) |
| `list_workflow_rules` | Read | List all rules (summarized: id, name, entity, trigger, active) |
| `list_my_alerts` | Alerts | User's personal alert inbox (summarized) |
| `get_alert_detail` | Alerts | Full detail on a specific alert by UUID |
| `list_alerts_for_rule` | Alerts | Alerts fired by a specific rule (summarized) |
| `create_workflow` | Write | Create a new rule (schema ref → `preview_workflow`) |
| `update_workflow` | Write | Update an existing rule (schema ref → `preview_workflow`) |
| `validate_workflow` | Write | Dry-run validation (schema ref → `preview_workflow`) |
| `preview_workflow` | Preview | Validate + send visual preview for user approval (has full schema) |
| `start_draft` | Draft | Start building a workflow incrementally |
| `add_draft_action` | Draft | Add an action to a draft |
| `remove_draft_action` | Draft | Remove an action from a draft |
| `preview_draft` | Draft | Assemble, validate, and preview a draft |

### Tool Call Lifecycle

```
1. LLM emits tool_use_start + tool_use_input events
2. Handler assembles ToolCall{ID, Name, Input}
3. Handler may inject workflow_id from context
4. Executor.Execute() dispatches to the Ichor REST API
5. Result is summarized (action types, rule lists, alert lists)
6. For preview tools: handler intercepts, emits workflow_preview SSE event
7. Result is fed back to LLM as ToolResult{ToolUseID, Content, IsError}
8. Handler emits tool_call_result SSE event
```

### Tool Result Summarization

Some tool results are summarized before being sent to the LLM to reduce token usage:

| Tool | What's summarized |
|---|---|
| `discover_action_types` | Full config schemas → field names only; descriptions truncated to first sentence; ports as string list |
| `list_workflow_rules` | Full rule objects → id, name, entity_name, trigger_type, is_active only |
| `list_my_alerts` / `list_alerts_for_rule` | Full alert objects → id, title, message, severity, status, created_at, recipient names only |

### Name Resolution

Several tools support name-based references that the executor resolves to UUIDs:

| Input | Resolved to | Example |
|---|---|---|
| `entity: "schema.table"` | `entity_id: UUID` | `"inventory.inventory_items"` → UUID |
| `trigger_type: "on_update"` | `trigger_type_id: UUID` | `"on_update"` → UUID |
| `workflow_id: "Rule Name"` | `workflow_id: UUID` | `"Low Stock Alert"` → UUID |

Resolution results are cached per executor instance.

### Edge Shorthand ("after" field)

Actions can declare their predecessor inline instead of building an explicit edges array:

```json
{
  "actions": [
    {"name": "Check Stock", "action_type": "evaluate_condition", ...},
    {"name": "Send Alert", "action_type": "create_alert", "after": "Check Stock:output-true", ...}
  ]
}
```

The executor generates edges automatically:
- Action without `"after"` → start edge
- `"after": "ActionName:port"` → sequence edge from that port
- `"after": "ActionName"` → sequence edge from the action type's default port

---

## 6. System Prompt Structure

The system prompt is assembled by `buildSystemPrompt()` from constant blocks:

```
┌─────────────────────────────────────┐
│ roleBlock                           │  Identity, capabilities, preview-first workflow
├─────────────────────────────────────┤
│ toolGuidance                        │  Workflow concepts, tool selection table
├─────────────────────────────────────┤
│ responseGuidance                    │  Formatting rules (interpret, don't dump JSON)
├─────────────────────────────────────┤
│ contextPreamble + context JSON      │  Current workflow state (if provided)
│ **Current workflow ID: `UUID`**     │  Extracted for easy LLM access
│ ```json                             │
│ {compact context JSON}              │
│ ```                                 │
└─────────────────────────────────────┘
```

For `context_type="tables"`, `tablesRoleBlock` replaces `roleBlock` + `toolGuidance`.

---

## 7. Error Handling

| Error Source | Client Sees | LLM Sees |
|---|---|---|
| Bad request body | HTTP 400 | — |
| Auth failure | HTTP 401 | — |
| LLM provider error | `event: error` | — |
| Tool execution error | `event: tool_call_result` with `is_error: true` | `{"error": "message"}` in ToolResult |
| Max loops (20) | `event: error` "Maximum tool-call rounds exceeded" | — |
| Client disconnect | — (stream ends) | Context cancelled |

Tool errors are non-fatal: the LLM receives the error message and can retry or explain the failure to the user.

---

## 8. Connection Lifecycle

1. Client sends POST with `Accept: text/event-stream`
2. Server clears read/write deadlines (SSE is long-lived)
3. Server detaches from request context deadline (prevents 5s timeout killing LLM inference)
4. Server monitors original context for client disconnect
5. SSE headers set: `Content-Type: text/event-stream`, `Cache-Control: no-cache`, `X-Accel-Buffering: no`
6. Agentic loop runs until final response or error
7. `event: message_complete` signals end of stream

CORS is handled via middleware, with explicit OPTIONS preflight support for the `/v1/agent/chat` endpoint.
