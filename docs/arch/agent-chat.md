# agent-chat

[api]=API layer [sdk]=SDK/shared layer
→=depends on ⊕=writes ⊗=reads [rag]=vector search

---

## Overview

Conversational AI endpoint backed by Gemini Flash 2.5. Streams responses via SSE.
Agent loop: LLM ↔ tool execution, up to 20 rounds per request.
Tool selection uses RAG (embedding cosine similarity, top-6) to avoid sending all 55
tool definitions to the LLM.

---

## ChatAPI [api]

file: api/domain/http/agentapi/chatapi/

route:
  POST /v1/agent/chat
  registered as RawHandlerFunc (bypasses OTEL middleware and HTTP server timeouts)
  middleware: CORS → Authentication

Why RawHandlerFunc:
  - SSE streams exceed the HTTP server's 10s WriteTimeout
  - OTEL wrapper writes 200 OK after handler return, conflicting with long-lived connection
  - Same pattern used by WebSocket routes

SSE lifecycle:
  1. Detach from request context via context.WithoutCancel() — preserves user/trace values,
     drops HTTP deadline
  2. Clear read/write deadlines via http.ResponseController before streaming
  3. newSSEWriter sets headers: Content-Type: text/event-stream, Cache-Control: no-cache,
     Connection: keep-alive, X-Accel-Buffering: no
  4. sseWriter.send(event, data) JSON-marshals payload, writes "event: {e}\ndata: {json}\n\n",
     flushes after each event

loop constants:
  maxAgentLoops = 20    // max LLM ↔ tool round-trips per request
  ragTopK       = 6     // tools retrieved from embedding index per message
  ragMinScore   = 0.0   // cosine similarity threshold (no filter)

---

## Agent Loop (StateMachine)

```
Request
  │
  ▼
ToolIndex.Search(userMsg, topK=6)     ← select relevant tool subset via RAG
  │
  ▼
Provider.StreamChat(req)              ← call LLM with tool definitions
  │
  ├─ SSE events → client (text deltas, thinking deltas)
  │
  ▼
StopForToolUse = true?
  ├─ NO  → done (send message_complete event)
  └─ YES → execute tool calls
             │
             ▼
           Executor.Execute(toolName, input)
             │
             ▼
           Append ToolResults to Messages
             │
             ▼
           loop++ < maxAgentLoops?
             ├─ YES → ToolIndex.Search(lastMsg, topK=6) → Provider.StreamChat(...)
             └─ NO  → send error "max iterations reached"
```

---

## Executor [sdk]

file: business/sdk/agenttools/executor.go  (~2394 lines)

```go
type Executor struct {
    log              *logger.Logger
    baseURL          string          // Ichor API root (e.g. "http://localhost:8080")
    http             *http.Client    // timeout: 30s

    // name → UUID resolution caches (populated lazily per token)
    entityCache      map[string]string        // "schema.table" → entity UUID
    triggerTypeCache map[string]string        // trigger name → UUID
    actionTypeCache  map[string]actionTypeInfo
    ruleCache        map[string]string        // rule name → UUID
    cacheMu          sync.Mutex

    // draft state for incremental workflow building
    drafts           map[string]*draftWorkflow  // keyed by session token
    draftMu          sync.Mutex
}

type draftWorkflow struct {
    lastAccess  time.Time        // TTL: 10 min
    name        string
    entity      string           // "schema.table" or UUID
    triggerType string
    description string
    triggerCond json.RawMessage  // optional
    actions     []draftAction
}
```

All 55 tool handlers live in executor.go — one method per tool name constant.
Calls Ichor REST API via http.Client with Bearer token from request context.
Draft builder tools (StartDraft, AddDraftAction, RemoveDraftAction, PreviewDraft)
maintain in-memory state per session; state is lost on server restart.

---

## ToolIndex [sdk]

file: business/sdk/toolindex/toolindex.go

```go
type ToolIndex struct {
    tools    []indexedTool    // tool + precomputed embedding vector
    embedder Embedder
    log      *logger.Logger
}

type Embedder interface {
    Embed(ctx context.Context, text string) ([]float32, error)
}

// optional batch extension
type BatchEmbedder interface {
    Embedder
    BatchEmbed(ctx context.Context, texts []string) ([][]float32, error)
}
```

api:
  New(ctx, cfg Config, tools []llm.ToolDef) (*ToolIndex, error)
  Search(ctx, message string, topK int, opts SearchOptions) ([]ToolMatch, time.Duration, error)

SearchOptions.Allowlist restricts candidates to a named subset before scoring.
ToolMatch.Score is cosine similarity in [-1, 1].
Embedding source per tool: Name + Description + ExampleQueries (ExampleQueries are
never sent to the LLM; they only improve retrieval accuracy).

---

## LLM Provider [sdk]

file: business/sdk/llm/

```go
type Provider interface {
    StreamChat(ctx context.Context, req ChatRequest) (<-chan StreamEvent, error)
}

type ChatRequest struct {
    SystemPrompt string
    Messages     []Message
    Tools        []ToolDef
    MaxTokens    int
}

type ToolDef struct {
    Name           string
    Description    string
    InputSchema    json.RawMessage  // JSON Schema object
    ExampleQueries []string         // RAG only — never forwarded to LLM
}

type StreamEvent struct {
    Type           EventType   // message_start | content_delta | thinking_delta |
                               // tool_use_start | tool_use_input | message_complete | error
    Text           string      // content_delta
    ThinkingText   string      // thinking_delta (server-side only, not forwarded to client)
    ToolCallID     string      // tool_use_start
    ToolCallName   string      // tool_use_start
    PartialInput   string      // tool_use_input (partial JSON)
    StopForToolUse bool        // true = caller should execute tools and loop
    Err            error       // error
}
```

Active provider: Gemini Flash 2.5 (not Claude).
Provider is injected at startup; swapping requires only a new Provider implementation.

---

## ToolCatalog [sdk]

file: business/sdk/toolcatalog/toolcatalog.go

55 tool name constants organized in two groups:

GroupWorkflow tools (workflow discovery, read, write, draft, alerts):
  Discover, DiscoverActionTypes, DiscoverTriggerTypes, DiscoverEntityTypes, DiscoverEntities
  GetWorkflow, GetWorkflowRule, ExplainWorkflowNode, ExplainWorkflowPath
  ListWorkflows, ListWorkflowRules, ListActionTemplates
  ValidateWorkflow, CreateWorkflow, UpdateWorkflow, PreviewWorkflow
  AnalyzeWorkflow, SuggestTemplates, ShowCascade
  StartDraft, AddDraftAction, RemoveDraftAction, PreviewDraft
  ListMyAlerts, GetAlertDetail, ListAlertsForRule
  SearchDatabaseSchema, SearchEnums

GroupTables tools (UI config, page content, forms, table configs):
  DiscoverConfigSurfaces, DiscoverFieldTypes, DiscoverContentTypes, DiscoverTableReference
  GetPageConfig, GetPageContent, GetTableConfig, GetFormDef
  ListPages, ListForms, ListTableCfgs
  CreatePageConfig, UpdatePageConfig, CreatePageContent, UpdatePageContent
  CreateForm, AddFormField, CreateTableConfig, UpdateTableConfig
  ValidateTableConfig, PreviewTableConfig
  ApplyColumnChange, ApplyFilterChange, ApplyJoinChange, ApplySortChange

api:
  InGroup(toolName string, group ToolGroup) bool
  ToolsForGroup(group ToolGroup) []string
  AllTools() []string

---

## ⚠ Adding a new agent tool

  business/sdk/toolcatalog/toolcatalog.go   (add constant + assign to group(s))
  business/sdk/agenttools/executor.go        (add Execute case + handler method)
  business/sdk/llm/                          (ToolDef.ExampleQueries — improve RAG recall)
  mcp/tools/                                 (add corresponding MCP tool if needed — separate module)
  api/cmd/services/ichor/tests/agentapi/     (integration test for new tool call)

## ⚠ Changing the LLM provider

  business/sdk/llm/                  (implement new Provider interface)
  api/cmd/services/ichor/build/all/  (inject new provider at startup)
  system prompt in chatapi/          (may need re-tuning for new model's strengths)
  ToolDef.ExampleQueries             (re-tune for new model's embedding space if provider
                                      also changes the embedder)

## ⚠ Changing SSE streaming behavior

  api/domain/http/agentapi/chatapi/chatapi.go
    (context.WithoutCancel, http.ResponseController deadline clearing, sseWriter)
  Note: route MUST remain RawHandlerFunc — switching to standard HandlerFunc
  will break streaming (OTEL wrapper + WriteTimeout conflict)
