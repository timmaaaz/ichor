# Chat RAG Architecture

## PURPOSE
This document is structured for LLM-driven Excalidraw diagram generation.
Sections are labeled by diagram element type: NODES, EDGES, GROUPS, ANNOTATIONS.

---

## DIAGRAM: Tool Selection Pipeline

### NODES (rectangles unless noted)

```
ID: user_message
LABEL: "User Message"
TYPE: rectangle

ID: context_filter
LABEL: "Stage 1: Context Filter\n(filterToolsByContext)"
TYPE: rectangle
FILE: chatapi.go:543

ID: rag_search
LABEL: "Stage 2: RAG Semantic Search\n(toolIndex.Search)"
TYPE: rectangle
FILE: toolindex/toolindex.go

ID: core_merge
LABEL: "Stage 3: Core Tool Merge\n(mergeTools)"
TYPE: rectangle
FILE: chatapi.go:507

ID: llm_call
LABEL: "LLM Call\n(Gemini Flash 2.5)\n~8-9 tools"
TYPE: rectangle

ID: context_filter_output_workflow
LABEL: "~27 workflow tools"
TYPE: text (annotation)

ID: context_filter_output_tables
LABEL: "~23 table tools"
TYPE: text (annotation)

ID: rag_output
LABEL: "top 6 by cosine similarity"
TYPE: text (annotation)

ID: core_output
LABEL: "+ 2-3 core tools\n(always included)"
TYPE: text (annotation)
```

### EDGES (arrows)

```
user_message → context_filter
  LABEL: "context_type: workflow | tables"

context_filter → rag_search
  LABEL: "allowlisted subset"

rag_search → core_merge
  LABEL: "top ragTopK=6 matches"

core_merge → llm_call
  LABEL: "final ~8-9 tools"
```

### GROUPS

```
GROUP: pipeline
MEMBERS: user_message, context_filter, rag_search, core_merge, llm_call
```

---

## DIAGRAM: RAG Internals

### NODES

```
ID: tool_def
LABEL: "ToolDef\n- Name\n- Description\n- InputSchema\n- ExampleQueries[]"
TYPE: rectangle
FILE: agenttools/definitions.go

ID: embed_text
LABEL: "Embed Text Construction\n{name} — {desc} | {ex1} | {ex2}"
TYPE: rectangle

ID: embedder
LABEL: "Embedder Interface\n(Embed, BatchEmbed)"
TYPE: rectangle (interface)
FILE: toolindex/embedder.go

ID: gemini_embedder
LABEL: "GeminiEmbedder\nmodel: gemini-embedding-001\nbatch: yes\ntimeout: 30s"
TYPE: rectangle
FILE: toolindex/gemini.go

ID: ollama_embedder
LABEL: "OllamaEmbedder\nmodel: nomic-embed-text\nbatch: yes\ntimeout: 30s"
TYPE: rectangle
FILE: toolindex/ollama.go

ID: tool_index
LABEL: "ToolIndex\n- []indexedTool (precomputed vectors)\n- Embedder\n- L2 normalized"
TYPE: rectangle
FILE: toolindex/toolindex.go

ID: query_embed
LABEL: "Embed user message\n(same embedder)"
TYPE: rectangle

ID: dot_product
LABEL: "Dot product\n(= cosine sim on\nnormalized vectors)"
TYPE: diamond

ID: sort_truncate
LABEL: "Sort desc → top K=6"
TYPE: rectangle

ID: search_result
LABEL: "[]ToolMatch\n{name, score}"
TYPE: rectangle
```

### EDGES

```
tool_def → embed_text  LABEL: "ExampleQueries concatenated"
embed_text → embedder  LABEL: "at startup"
embedder → tool_index  LABEL: "store vectors"

gemini_embedder → embedder  LABEL: "implements"
ollama_embedder → embedder  LABEL: "implements"

query_embed → dot_product  LABEL: "normalized query vector"
tool_index → dot_product   LABEL: "normalized tool vectors\n(allowlist filtered)"
dot_product → sort_truncate
sort_truncate → search_result
```

### GROUPS

```
GROUP: startup_indexing
MEMBERS: tool_def, embed_text, embedder, gemini_embedder, ollama_embedder, tool_index
LABEL: "Startup (once)"

GROUP: runtime_search
MEMBERS: query_embed, dot_product, sort_truncate, search_result
LABEL: "Per Request"
```

---

## DIAGRAM: Full Request Flow (Sequence)

### ACTORS (left to right)

```
frontend
chatapi
prompt_builder
tool_pipeline
llm_provider
tool_executor
ichor_api
```

### SEQUENCE STEPS

```
1. frontend → chatapi: POST /v1/agent/chat\n{message, context_type, context, history}
2. chatapi → prompt_builder: build system prompt
3. prompt_builder → chatapi: role block + context JSON
4. chatapi → tool_pipeline: filterToolsByContext(context_type)
5. chatapi → tool_pipeline: toolIndex.Search(message, 6, allowlist)
6. chatapi → tool_pipeline: mergeTools(core, rag_results)
7. tool_pipeline → chatapi: ~8-9 tools
8. chatapi → llm_provider: StreamChat(system_prompt, history, tools)
9. llm_provider → chatapi: StreamEvent channel (SSE)
10. chatapi → frontend: SSE: content_chunk / tool_call_start
11. chatapi → tool_executor: Execute(tool_name, input)
12. tool_executor → chatapi: inject context IDs if UUID missing
13. tool_executor → ichor_api: HTTP call (REST endpoint)
14. ichor_api → tool_executor: result
15. [IF preview tool]: chatapi → frontend: SSE: workflow_preview | table_config_preview
16. tool_executor → chatapi: ToolResult
17. chatapi → llm_provider: continue loop with tool result
18. [LOOP until no more tool calls]
19. chatapi → frontend: SSE: message_complete
```

---

## DIAGRAM: Degradation / Fallback Tree

### NODES

```
ID: check_embedder
LABEL: "Embedder configured?"
TYPE: diamond

ID: check_index_build
LABEL: "Index build success?"
TYPE: diamond

ID: check_search
LABEL: "Search success at runtime?"
TYPE: diamond

ID: rag_enabled
LABEL: "RAG enabled\n(top 6 + core)"
TYPE: rectangle

ID: fallback_context
LABEL: "Fallback: all\ncontext-filtered tools"
TYPE: rectangle
STYLE: dashed

ID: fallback_all
LABEL: "Fallback: all tools"
TYPE: rectangle
STYLE: dashed
```

### EDGES

```
check_embedder → check_index_build  LABEL: "yes"
check_embedder → fallback_context   LABEL: "no"
check_index_build → rag_enabled     LABEL: "yes"
check_index_build → fallback_context LABEL: "no"
check_search → rag_enabled          LABEL: "success"
check_search → fallback_context     LABEL: "error (logged)"
```

---

## DIAGRAM: Core Tool Guarantee

### DATA TABLE (for annotation)

```
CONTEXT         | CORE TOOLS (always included)
----------------|-------------------------------
workflow        | list_workflow_rules, discover
tables          | list_table_configs, discover_table_reference
new_workflow    | start_draft, add_draft_action, remove_draft_action, preview_draft
```

---

## DIAGRAM: Preview Interception

### NODES

```
ID: tool_executes
LABEL: "Tool executes\n(dry_run via API)"
TYPE: rectangle

ID: is_preview
LABEL: "Preview interceptor?"
TYPE: diamond

ID: emit_sse
LABEL: "Emit SSE event\nto frontend\n(workflow_preview |\ntable_config_preview)"
TYPE: rectangle

ID: synthetic_result
LABEL: "Return synthetic result\n'preview_sent'\nto LLM"
TYPE: rectangle

ID: normal_result
LABEL: "Return full result\nto LLM"
TYPE: rectangle
```

### EDGES

```
tool_executes → is_preview
is_preview → emit_sse        LABEL: "yes"
emit_sse → synthetic_result
is_preview → normal_result   LABEL: "no"
```

### INTERCEPTOR MAP (annotation)

```
preview_workflow     → SSE: workflow_preview
preview_draft        → SSE: workflow_preview
preview_table_config → SSE: table_config_preview
```

---

## KEY CONSTANTS (for diagram annotations)

```
ragTopK     = 6     (chatapi.go:38)
ragMinScore = 0     (chatapi.go:42)
max_tokens  = 4096
```

## KEY FILE INDEX (for diagram tooltips)

```
chatapi.go                          api/domain/http/agentapi/chatapi/chatapi.go
prompt.go                           api/domain/http/agentapi/chatapi/prompt.go
events.go                           api/domain/http/agentapi/chatapi/events.go
toolindex/toolindex.go              business/sdk/toolindex/toolindex.go
toolindex/embedder.go               business/sdk/toolindex/embedder.go
toolindex/gemini.go                 business/sdk/toolindex/gemini.go
toolindex/ollama.go                 business/sdk/toolindex/ollama.go
agenttools/definitions.go           business/sdk/agenttools/definitions.go
agenttools/executor.go              business/sdk/agenttools/executor.go
toolcatalog/toolcatalog.go          business/sdk/toolcatalog/toolcatalog.go
llm/provider.go                     business/sdk/llm/provider.go
all.go                              api/cmd/services/ichor/build/all/all.go
```
