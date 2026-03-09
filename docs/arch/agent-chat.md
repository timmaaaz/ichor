# agent-chat

[bus]=business [app]=application [api]=HTTP [db]=store [sdk]=shared [rag]=vector-search
→=depends on ⊕=writes ⊗=reads ⚡=external [tx]=transaction [cache]=cached

---

## StateMachine

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

---

## ChatAPI [api]

file: api/domain/http/agentapi/chatapi/
key facts:
  - route: POST /v1/agent/chat
  - registered as RawHandlerFunc (bypasses OTEL middleware and HTTP server timeouts)
  - middleware: CORS → Authentication
  - SSE streams exceed the HTTP server's 10s WriteTimeout — must remain RawHandlerFunc
  - OTEL wrapper writes 200 OK after handler return, conflicting with long-lived connection

loop constants:
  maxAgentLoops = 20    // max LLM ↔ tool round-trips per request
  ragTopK       = 6     // tools retrieved from embedding index per message
  ragMinScore   = 0.0   // cosine similarity threshold (no filter)

SSE lifecycle:
  1. Detach from request context via context.WithoutCancel() — preserves user/trace values, drops HTTP deadline
  2. Clear read/write deadlines via http.ResponseController before streaming
  3. newSSEWriter sets headers: Content-Type: text/event-stream, Cache-Control: no-cache,
     Connection: keep-alive, X-Accel-Buffering: no
  4. sseWriter.send(event, data) JSON-marshals payload, writes "event: {e}\ndata: {json}\n\n", flushes after each event

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

key facts:
  - All 53 tool handlers live in executor.go — one method per tool name constant (verified 2026-03-09)
  - Calls Ichor REST API via http.Client with Bearer token from request context
  - Draft builder tools (StartDraft, AddDraftAction, RemoveDraftAction, PreviewDraft) maintain in-memory state per session
  - Draft state is lost on server restart

---

## ToolIndex [sdk]

file: business/sdk/toolindex/toolindex.go

```go
type ToolIndex struct {
    tools    []indexedTool    // tool + precomputed embedding vector
    embedder Embedder
    log      *logger.Logger
}
```

<!-- lsp:hover:13:6 -->
```go
type Embedder interface {
    // Embed returns a normalised embedding vector for the given text.
    Embed(ctx context.Context, text string) ([]float32, error)
}
```
<!-- lsp:refs:13:6 --> count=2 (gemini, ollama; excl. test mocks)

```go
// optional batch extension
type BatchEmbedder interface {
    Embedder
    BatchEmbed(ctx context.Context, texts []string) ([][]float32, error)
}
```

  New(ctx, cfg Config, tools []llm.ToolDef) (*ToolIndex, error)
  Search(ctx, message string, topK int, opts SearchOptions) ([]ToolMatch, time.Duration, error)

key facts:
  - SearchOptions.Allowlist restricts candidates to a named subset before scoring
  - ToolMatch.Score is cosine similarity in [-1, 1]
  - Embedding source per tool: Name + Description + ExampleQueries
  - ExampleQueries are never sent to the LLM; they only improve retrieval accuracy

---

## LLMProvider [sdk]

file: business/sdk/llm/

<!-- lsp:hover:12:6 -->
```go
type Provider interface {
    // StreamChat sends a request and returns a channel of streaming events.
    // The channel is closed when the response is complete or an error occurs.
    // The caller must drain the channel.
    StreamChat(ctx context.Context, req ChatRequest) (<-chan StreamEvent, error)
}
```
<!-- lsp:refs:12:6 --> count=3 (gemini/active, claude, ollama; excl. test mocks)

```go
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

Implementors (verified 2026-03-09):
  business/sdk/llm/gemini/gemini.go   ← ACTIVE (Gemini Flash 2.5)
  business/sdk/llm/claude/claude.go   ← available (not active)
  business/sdk/llm/ollama/ollama.go   ← available (not active)

key facts:
  - Active provider: Gemini Flash 2.5 (not Claude)
  - Provider is injected at startup; swapping requires only a new Provider implementation

---

## ToolCatalog [sdk]

file: business/sdk/toolcatalog/toolcatalog.go
key facts:
  <!-- lsp:refs:12:1 --> count=53 (exported constants, documentSymbol)
  - 53 tool name constants organized in two groups (verified 2026-03-09)

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
  verify: documentSymbol(business/sdk/toolcatalog/toolcatalog.go) — confirm constant is in correct group (GroupWorkflow vs GroupTables) and total count updated

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

