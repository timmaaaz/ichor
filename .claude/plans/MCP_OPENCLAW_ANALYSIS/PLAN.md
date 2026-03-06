# MCP OpenClaw Enterprise Analysis Plan

## Context

This plan analyzes the Ichor MCP server from the perspective of an **OpenClaw Enterprise client**.
OpenClaw clients connect their LLM agents (Claude Desktop, custom agents) to a running Ichor
deployment via the MCP server using a Bearer token. The analysis covers:

- What can they DO with the MCP today?
- What can they NOT do (gaps, missing tools, blocked domains)?
- What PROTECTIONS exist (auth, authz, validation, hardening)?
- What RISKS exist (attack surface, weak spots, missing controls)?

The MCP server exposes 33 tools, 7 resources, and 3 guided prompts across 3 context modes
(all / workflow / tables). It is a thin translation layer — all auth/authz decisions are
delegated to the Ichor REST API.

---

## Phase 1 — Capability Inventory: What Can Clients DO?

**Objective**: Produce a precise, categorized inventory of every operation an OpenClaw
enterprise client can perform today, mapped to business outcomes they can achieve.

### Sub-scope:

1. **Tool-by-tool functional audit**
   - Map all 33 tools to: inputs, outputs, side effects, target REST endpoint
   - Classify each as: Read / Write / Validate / Analyse
   - Identify what business intent each tool serves (e.g., "design a new approval workflow")

2. **Resource and Prompt audit**
   - Catalog all 7 resources (5 static + 2 templates) and what data they surface
   - Catalog all 3 prompts and what guided flows they enable

3. **Context mode mapping**
   - Document exactly which tools each context mode (`all`, `workflow`, `tables`) exposes
   - Identify the intended agent personas for each mode

4. **Business domain coverage matrix**
   - For each Ichor domain (workflow, pages, forms, tables, users, orders, inventory, etc.):
     assess what fraction of CRUD is exposed via MCP
   - Separate: configuration-plane access vs. data-plane access

5. **Achievable enterprise use-cases**
   - List concrete workflows an OpenClaw enterprise agent can execute end-to-end
     (e.g., "create a new inventory approval workflow", "build a form for a new page")
   - Note prerequisites (token scope, role permissions required)

**Depth target**: Every tool has a row in a capability matrix. No tool is undocumented.

---

## Phase 2 — Access Boundary Analysis: What Should and Shouldn't Clients Be Able to Do?

**Objective**: Evaluate access boundaries from two complementary angles:
(A) **Under-exposure** — operations that are absent but should exist (gaps),
and (B) **Over-exposure** — operations that are present but should not be, or should be
gated more tightly (least-privilege violations). Both are access boundary failures; one
leaves capability on the table, the other creates unnecessary risk.

### Part A — Under-Exposure: What Can Clients NOT Do (But Should)?

1. **Missing CRUD operations**
   - DELETE is absent for all resource types (workflows, pages, forms, tables)
   - UPDATE is missing for forms (only add_form_field exists)
   - Page actions have no write tools (read-only)
   - Action templates have no write tools (read-only)
   - Quantify: what percentage of full CRUD is exposed per domain?

2. **Missing data-plane access**
   - Actual business data (orders, inventory, products, users) is completely inaccessible
   - Workflow execution history, logs, and runtime state are inaccessible
   - Agents cannot verify if a workflow they created is actually running
   - Form submission data is inaccessible
   - Assess: what can OpenClaw NOT automate as a result?

3. **Missing configuration operations**
   - No tools to manage roles, permissions, or user accounts
   - No tools for system settings or environment configuration
   - No tools to manage alerts configuration directly

4. **Operational gaps**
   - Cannot pause, cancel, or manually trigger workflows
   - Cannot read audit logs for agent-initiated changes
   - Cannot manage workflow execution schedules

5. **Integration gaps**
   - No webhook or callback registration via MCP
   - No batch operations (bulk create/update)
   - No export/import of configurations

6. **Gap prioritization**
   - Rank gaps by: enterprise impact × implementation cost
   - Identify gaps that block core OpenClaw use-cases vs. nice-to-haves

---

### Part B — Over-Exposure: What Can Clients Do (But Shouldn't, or Not by Default)?

**The principle of least privilege asks not just "can agents do what they need?" but
"can agents do things they should never need?"** Some MCP capabilities are legitimately
dangerous for an LLM agent to hold without additional controls.

1. **Full schema introspection (`search_database_schema`)**
   - Exposes complete table names, column names, types, and foreign key relationships
     across ALL PostgreSQL schemas to any authenticated agent
   - An agent (or a prompt-injected agent) can map the entire data model before
     attempting further exploitation
   - Question: should a workflow-building agent need to see the `core.users` table schema?
   - Assessment: currently gated by `auth.RuleAdminOnly` — but is that sufficient?
     Admin tokens are often what enterprise agents run with.

2. **Workflow creation without human review**
   - An agent can autonomously `create_workflow` with live trigger hooks and destructive
     actions (inventory allocation, email notifications, approvals) with zero human
     confirmation step
   - The validate-first pattern catches malformed graphs, not semantically dangerous ones
     (e.g., a valid workflow that sends 10,000 emails on every order update)
   - Question: should autonomous agents be able to deploy live automation without
     a human-in-the-loop approval gate?

3. **`validate=false` bypass on workflow writes**
   - Both `create_workflow` and `update_workflow` accept `validate: false` which skips
     the dry-run entirely — any authenticated agent can pass this flag
   - This exists as a developer escape hatch but is fully exposed to enterprise clients
   - Question: should a non-admin agent ever be permitted to bypass validation?

4. **Enum introspection as business vocabulary harvesting**
   - `search_enums` returns human-readable labels for all PostgreSQL enum types
   - Gated by `auth.RuleAny` (any authenticated user) — the lowest possible bar
   - Exposes business vocabulary: status names, category labels, condition codes, etc.
   - A compromised or over-shared token grants this to anyone
   - Question: is enum read access appropriate for all roles, or only specific ones?

5. **Context mode `all` as default**
   - Running `--context all` (the default) registers every tool — write tools for
     workflows AND UI components simultaneously
   - An agent scoped to "build approval workflows" has no need for `create_table_config`
     or `add_form_field`
   - Default-all violates least privilege; the safer default would be read-only
   - Question: should `--context all` require an explicit opt-in flag?

6. **Prompts pre-fetch sensitive context without scope checks**
   - The `build-workflow` prompt pre-fetches ALL action types and trigger types and
     injects them into the agent's context before the agent has done anything
   - This front-loads capability knowledge — including action types for destructive
     operations (inventory allocation, external webhooks) — into every prompt call
   - Question: should prompts pre-load only the action types relevant to the stated
     use-case, not the full catalog?

7. **No differentiation between read-only and read-write tokens**
   - A single Bearer token grants access to all tools registered for the context mode
   - There is no token scope that allows "read workflows but not write them"
   - An agent performing analysis (`explain_workflow_path`, `analyze_workflow`) holds
     the same credentials as one performing creation (`create_workflow`)
   - Question: should MCP support token scope declarations (read-only vs. read-write)?

---

### Part C — Intentional vs. Accidental Restrictions

For each identified under-exposure gap, classify it as:
- **By design**: deliberately excluded (e.g., user management is out of MCP scope)
- **Unplanned**: missing because it wasn't built yet, not because it shouldn't exist
- **Ambiguous**: unclear whether exclusion was intentional

This classification matters for OpenClaw clients: "by design" gaps should inform their
integration strategy; "unplanned" gaps are candidates for roadmap requests.

---

**Depth target**: Every item in Part A has a business impact statement and remediation
path. Every item in Part B has a risk statement, a "who is harmed if exploited?" answer,
and a recommended control (scope restriction, opt-in flag, human gate, or role-gating).

---

## Phase 3 — Protection Analysis: What Safeguards Exist?

**Objective**: Evaluate every security and reliability control present in the MCP stack,
from the stdio transport layer through to the PostgreSQL backend.

### Sub-scope:

1. **Authentication layer**
   - Bearer token mechanism: how tokens are injected, transmitted, stored
   - Token lifecycle: expiry, rotation, revocation
   - CLI flag vs. env var security comparison (`--token` appears in process args)
   - No token → server exits at startup (hard fail, good)

2. **Authorization layer**
   - How Ichor's table-based RBAC is enforced (PermissionsBus)
   - Which MCP tools hit `auth.RuleAdminOnly` endpoints (database introspection)
   - Which hit `auth.RuleAny` (enum lookups)
   - Which hit table-level permission checks (workflow, config, data)
   - Critical: MCP server has ZERO per-tool authorization — all enforced by REST API

3. **Input validation controls**
   - Required field checks at MCP tool level (id, name, etc.)
   - Workflow validation-first pattern (validate → create/update)
   - What is NOT validated at MCP level (JSON schema, UUID format, SQL injection)
   - Passthrough JSON strategy: implications for injection attacks

4. **Transport security**
   - stdio transport: no network exposure, inherently local
   - All traffic to Ichor REST API is HTTP (not HTTPS in dev setup)
   - 30-second timeout per request: DoS implications

5. **Reliability protections**
   - Validate-first pattern prevents broken workflow creation
   - `validate=false` bypass: present risk for create/update workflow tools
   - Error propagation: how backend errors surface to agents

6. **Audit and observability**
   - Does the MCP server log tool calls?
   - Can admins see which agent called which tool?
   - Is there a request ID / correlation strategy?

**Depth target**: Every protection is rated: Strong / Adequate / Weak / Absent,
with a rationale for each rating.

---

## Phase 4 — Risk & Attack Surface Analysis

**Objective**: Model the attack surface systematically — from both external threat actors
and from the LLM agent itself (prompt injection, capability misuse) — and identify the
highest-risk vulnerabilities for an enterprise OpenClaw deployment.

### Sub-scope:

1. **Threat model: who are the adversaries?**
   - External attacker with stolen Bearer token
   - Malicious prompt injection via user input reaching the agent
   - Over-privileged agent with a high-permission token
   - Insider threat using agent to exfiltrate schema/config data

2. **Attack surface: MCP server**
   - `validate=false` bypass: agent can create unvalidated workflows
   - Resource URI parsing: `config://db/{schema}/{table}` — path traversal / injection?
   - Unbounded graph traversal in `explain_workflow_path`: cycle detection?
   - Passthrough JSON to REST API: malformed/oversized payloads

3. **Attack surface: auth/authz**
   - Bearer token in `--token` CLI arg: visible in `ps aux` / process listing
   - Bearer token in MCP config files (Claude Desktop): file permission exposure
   - No CSRF / replay protection on REST API requests
   - No per-tool scope restrictions: one token grants all 33 tools

4. **Attack surface: data exposure**
   - `search_database_schema` exposes full schema to any authenticated agent
   - `search_enums` exposes business vocabulary (enum values, labels)
   - Workflow configs may contain sensitive business logic as plaintext

5. **Agent-specific risks (LLM misuse)**
   - Prompt injection: can malicious data in workflow names/configs redirect the agent?
   - Autonomous write access: agent can create/update workflows without human review
   - No human-in-the-loop controls: no approval step before workflow creation
   - Agent could create conflicting or duplicate workflows silently

6. **Risk scoring matrix**
   - Score each finding: Likelihood × Impact
   - Prioritize top 5 risks for immediate remediation recommendation

**Depth target**: A STRIDE or equivalent threat model with concrete attack scenarios per
surface, not just abstract categories.

---

## Phase 5 — Enterprise Hardening Recommendations

**Objective**: Translate Phase 3 and Phase 4 findings into a concrete, prioritized set of
improvements — both what OpenClaw clients should demand from a hardened deployment, and
what the Ichor MCP server itself should implement.

### Sub-scope:

1. **Immediate / low-effort hardening** (can be done today without API changes)
   - Token management best practices (short-lived tokens, env var over CLI flag)
   - Claude Desktop config file permissions
   - HTTPS enforcement for Ichor API in production
   - Context mode discipline (don't run `--context all` unless needed)

2. **MCP server-level improvements** (changes to `mcp/` module)
   - Remove `validate=false` bypass or gate behind explicit admin flag
   - Add depth limit to `explain_workflow_path` graph traversal
   - Add request ID logging to every tool call
   - Add per-tool audit log entries (tool name, user token hash, timestamp)
   - Validate resource URI components before forwarding to backend

3. **Backend API improvements** (changes to Ichor REST API)
   - Add dry-run endpoints for UI config writes (pages, forms, tables)
   - Add DELETE tools for workflows, pages, forms
   - Add UPDATE form tool (not just add_form_field)
   - Add page action write tools
   - Rate limiting middleware for agent-sourced requests (identifiable via X-Agent header)

4. **Authorization hardening**
   - Scoped tokens: separate read-only tokens from write tokens
   - Per-context-mode token scopes (workflow token cannot call UI write tools)
   - Human-in-the-loop option: flag writes as "pending approval" before commit

5. **Enterprise deployment checklist**
   - What an OpenClaw enterprise deployment of Ichor + MCP should verify before go-live
   - Monitoring and alerting recommendations for agent activity
   - Incident response: how to detect and respond to agent misuse

**Depth target**: Each recommendation includes: problem it solves, implementation sketch,
effort estimate (S/M/L), and which threat from Phase 4 it mitigates.

---

## Execution Notes

- Each phase should be done in depth — no phase is a bullet-point summary.
- Phases 1 and 2 are largely analytical (read existing code, produce matrices).
- Phases 3 and 4 require security-minded reasoning and threat modeling.
- Phase 5 produces actionable output for both OpenClaw clients and Ichor developers.
- Code changes are NOT part of this analysis plan — findings only.

## Output Artifacts per Phase

| Phase | Primary Output |
|-------|---------------|
| 1 | Capability matrix (tool → domain → CRUD → use-case) |
| 2 | Gap registry (gap → impact → remediation) |
| 3 | Protection assessment (control → strength rating → evidence) |
| 4 | Threat model (actor → surface → scenario → risk score) |
| 5 | Hardening roadmap (recommendation → effort → mitigated threat) |

## Key Files Referenced

| File | Relevance |
|------|-----------|
| `mcp/cmd/ichor-mcp/main.go` | Entry point, CLI flags, context mode routing |
| `mcp/internal/tools/register.go` | Tool registration per context mode |
| `mcp/internal/client/ichor.go` | HTTP client, Bearer auth, all REST calls |
| `mcp/internal/tools/write_workflow.go` | Validate-first pattern, validate=false risk |
| `mcp/internal/tools/search.go` | Schema introspection, resource URI parsing |
| `mcp/internal/tools/read_workflow.go` | Graph traversal (unbounded depth risk) |
| `mcp/internal/resources/resources.go` | Static resources, template URI parsing |
| `mcp/internal/prompts/prompts.go` | Guided prompts, pre-fetched context |
| `api/domain/http/introspectionapi/routes.go` | RuleAdminOnly enforcement |
| `api/cmd/services/ichor/build/all/all.go` | How all routes are wired |
