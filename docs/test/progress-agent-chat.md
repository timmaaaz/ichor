# Progress Summary: agent-chat.md

## Overview
Comprehensive architecture for Ichor's AI agent/chat system. Describes how the LLM, tool selection, and tool execution work together for workflow management and UI configuration tasks.

## State Machine (Request Flow)

The chat system runs a loop-based agent:

1. **User Message In**
2. **ToolIndex.Search** — RAG retrieval selects relevant tools (topK=6)
3. **Provider.StreamChat** — LLM generates response with tool definitions, SSE events stream to client
4. **Check StopForToolUse**
   - NO → Done (send `message_complete` event)
   - YES → Execute tool calls
5. **Executor.Execute** — run tool with LLM-provided input
6. **Append Results** — add tool results to message history
7. **Loop Check** — if `loop++ < maxAgentLoops` (20), repeat from step 2
8. **Max Exceeded** — error "max iterations reached"

## ChatAPI [api] — `api/domain/http/agentapi/chatapi/`

**Route:** `POST /v1/agent/chat`

### Key Technical Details
- Registered as **RawHandlerFunc** (bypasses OTEL middleware and HTTP timeouts)
- Middleware chain: CORS → Authentication
- Reason for RawHandlerFunc: SSE streams exceed the HTTP server's 10s WriteTimeout
- OTEL wrapper writes 200 OK after handler return, which breaks long-lived SSE connections

### Loop Constants
- `maxAgentLoops = 20` — max LLM ↔ tool round-trips per request
- `ragTopK = 6` — tools retrieved per message
- `ragMinScore = 0.0` — no cosine similarity threshold filtering

### SSE Lifecycle
1. **Detach from request context** — `context.WithoutCancel()` preserves user/trace values but drops HTTP deadline
2. **Clear deadlines** — `http.ResponseController` removes read/write timeouts before streaming
3. **sseWriter setup** — headers: `Content-Type: text/event-stream`, `Cache-Control: no-cache`, `Connection: keep-alive`, `X-Accel-Buffering: no`
4. **Send loop** — `send(event, data)` JSON-marshals, writes `"event: {e}\ndata: {json}\n\n"`, flushes after each

## Executor [sdk] — `business/sdk/agenttools/executor.go` (~2394 lines)

**Responsibility:** Execute tool calls against Ichor API.

### Struct
```go
type Executor struct {
    log              *logger.Logger
    baseURL          string          // Ichor API root
    http             *http.Client    // 30s timeout

    // Name → UUID resolution caches (lazy-populated per session token)
    entityCache      map[string]string
    triggerTypeCache map[string]string
    actionTypeCache  map[string]actionTypeInfo
    ruleCache        map[string]string
    cacheMu          sync.Mutex

    // Draft builder state (10 min TTL)
    drafts           map[string]*draftWorkflow
    draftMu          sync.Mutex
}

type draftWorkflow struct {
    lastAccess  time.Time
    name        string
    entity      string            // UUID or "schema.table"
    triggerType string
    description string
    triggerCond json.RawMessage   // optional
    actions     []draftAction
}
```

### Key Facts
- **53 tool handlers** — one method per tool name constant (verified 2026-03-09)
- **HTTP calls** — all tools call Ichor API via Bearer token from request context
- **Draft builder tools** — StartDraft, AddDraftAction, RemoveDraftAction, PreviewDraft maintain in-memory state per session
- **Draft state is ephemeral** — lost on server restart

## ToolIndex [sdk] — `business/sdk/toolindex/toolindex.go`

**Responsibility:** RAG-based tool selection via semantic similarity.

### Struct & Interfaces
```go
type ToolIndex struct {
    tools    []indexedTool
    embedder Embedder
    log      *logger.Logger
}

type Embedder interface {
    Embed(ctx context.Context, text string) ([]float32, error)
}

type BatchEmbedder interface {
    Embedder
    BatchEmbed(ctx context.Context, texts []string) ([][]float32, error)
}
```

### Methods
- `New(ctx, cfg Config, tools []llm.ToolDef) (*ToolIndex, error)` — initialize index
- `Search(ctx, message string, topK int, opts SearchOptions) ([]ToolMatch, time.Duration, error)` — retrieve tools

### Key Facts
- **SearchOptions.Allowlist** — restricts candidates before scoring
- **ToolMatch.Score** — cosine similarity in [-1, 1]
- **Embedding source** — tool Name + Description + ExampleQueries (ExampleQueries never sent to LLM, RAG-only)

## LLMProvider [sdk] — `business/sdk/llm/`

**Responsibility:** Abstract LLM streaming interface.

### Provider Interface
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
    InputSchema    json.RawMessage
    ExampleQueries []string        // RAG-only, never sent to LLM
}

type StreamEvent struct {
    Type           EventType  // message_start | content_delta | thinking_delta | tool_use_start | tool_use_input | message_complete | error
    Text           string     // content_delta
    ThinkingText   string     // thinking_delta (server-side only)
    ToolCallID     string     // tool_use_start
    ToolCallName   string     // tool_use_start
    PartialInput   string     // tool_use_input (partial JSON)
    StopForToolUse bool
    Err            error
}
```

### Implementations (verified 2026-03-09)
- `business/sdk/llm/gemini/gemini.go` — **ACTIVE** (Gemini Flash 2.5, hot-swappable)
- `business/sdk/llm/claude/claude.go` — available but not active
- `business/sdk/llm/ollama/ollama.go` — available but not active

### Key Facts
- **Active provider:** Gemini Flash 2.5 (NOT Claude)
- **Hot-swappable:** Only need new Provider implementation + injection in all.go

## ToolCatalog [sdk] — `business/sdk/toolcatalog/toolcatalog.go`

**Responsibility:** Registry of all available agent tools.

### Tool Organization
**53 tools total** (verified 2026-03-09), organized in two groups:

#### GroupWorkflow (discovery, read, write, draft, alerts)
- Discover, DiscoverActionTypes, DiscoverTriggerTypes, DiscoverEntityTypes, DiscoverEntities
- GetWorkflow, GetWorkflowRule, ExplainWorkflowNode, ExplainWorkflowPath
- ListWorkflows, ListWorkflowRules, ListActionTemplates
- ValidateWorkflow, CreateWorkflow, UpdateWorkflow, PreviewWorkflow
- AnalyzeWorkflow, SuggestTemplates, ShowCascade
- StartDraft, AddDraftAction, RemoveDraftAction, PreviewDraft
- ListMyAlerts, GetAlertDetail, ListAlertsForRule
- SearchDatabaseSchema, SearchEnums

#### GroupTables (UI config, pages, forms, table configs)
- DiscoverConfigSurfaces, DiscoverFieldTypes, DiscoverContentTypes, DiscoverTableReference
- GetPageConfig, GetPageContent, GetTableConfig, GetFormDef
- ListPages, ListForms, ListTableCfgs
- CreatePageConfig, UpdatePageConfig, CreatePageContent, UpdatePageContent
- CreateForm, AddFormField, CreateTableConfig, UpdateTableConfig
- ValidateTableConfig, PreviewTableConfig
- ApplyColumnChange, ApplyFilterChange, ApplyJoinChange, ApplySortChange

### Methods
- `InGroup(toolName string, group ToolGroup) bool`
- `ToolsForGroup(group ToolGroup) []string`
- `AllTools() []string`

## Change Patterns

### ⚠ Adding a New Agent Tool
Affects 6 areas:
1. `business/sdk/toolcatalog/toolcatalog.go` — add constant + assign to group(s)
2. `business/sdk/agenttools/executor.go` — add Execute case + handler method
3. `business/sdk/llm/` — tune ToolDef.ExampleQueries for RAG recall
4. `mcp/tools/` — add corresponding MCP tool if needed (separate module)
5. `api/cmd/services/ichor/tests/agentapi/` — integration test for tool call
6. **Verify:** `documentSymbol(toolcatalog.go)` — confirm constant in correct group, total count matches 53

### ⚠ Changing the LLM Provider
Affects 3 areas:
1. `business/sdk/llm/` — implement new Provider interface
2. `api/cmd/services/ichor/build/all/` — inject new provider at startup
3. System prompt in `chatapi/` — may need re-tuning for new model's strengths
4. ToolDef.ExampleQueries — re-tune for new model's embedding space

### ⚠ Changing SSE Streaming Behavior
Affects 1 file (with care):
1. `api/domain/http/agentapi/chatapi/chatapi.go` — modify context.WithoutCancel, http.ResponseController, sseWriter
2. **Critical:** Route MUST remain RawHandlerFunc — switching to standard HandlerFunc breaks streaming (OTEL + WriteTimeout conflict)

## Critical Points
- Draft state is **ephemeral** — rebuilds with server restart
- ExampleQueries improve RAG but are **never sent to the LLM**
- Thinking text is **server-side only** — not forwarded to client
- SSE must be RawHandlerFunc to avoid OTEL/timeout conflicts
- Tool cache is session-keyed and lazy-populated
- Provider is hot-swappable via dependency injection

## Notes for Future Development
This is a sophisticated agent loop with careful streaming handling. Most changes should be tool additions (straightforward), not core loop modifications (risky).
