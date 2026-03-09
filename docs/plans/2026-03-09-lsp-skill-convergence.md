# LSP Skill Convergence Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Converge arch-go with arch-vue's LSP maturity by adding Mode D (lsp-enrich), migrating inline coordinate notation, and wiring LSP pre-steps into investigate, fix-issue, and spawn-agents.

**Architecture:** Edit 4 skill SKILL.md files under `~/.claude/skills/` and migrate 6 Ichor arch docs under `docs/arch/`. No code changes — pure doc/skill edits. Migration converts `## LSP Volatile Facts` bottom blocks into inline `<!-- lsp:hover:LINE:COL -->` and `<!-- lsp:refs:LINE:COL -->` annotations placed next to the relevant content.

**Tech Stack:** Markdown skill files, LSP tool (for migration tasks), git for commits.

**Design doc:** `docs/plans/2026-03-09-lsp-skill-convergence-design.md`

---

## Task 1: arch-go — Mode Detection + Mode D entry

**Files:**
- Modify: `~/.claude/skills/arch-go/SKILL.md`

**Step 1: Read the current file**

Read `~/.claude/skills/arch-go/SKILL.md` — focus on the Mode Detection section (lines ~18-29).

**Step 2: Add Mode D to Mode Detection**

Find this block:
```
**`/arch-go update <system>`** (e.g., `/arch-go update workflow`):
→ run **Mode C: Targeted Update** for that system
```

Add immediately after it:
```
**`/arch-go lsp-enrich [system]`** (e.g., `/arch-go lsp-enrich workflow-engine`):
→ run **Mode D: LSP Enrich** for that system (or all arch files if no arg)
```

**Step 3: Verify**

Read the Mode Detection section — confirm 4 mode entries are present (A, B, C, D).

**Step 4: Commit**

```bash
git -C ~/.claude commit -am "feat(arch-go): add Mode D to mode detection"
```

---

## Task 2: arch-go — Update Mode B Stage 2

**Files:**
- Modify: `~/.claude/skills/arch-go/SKILL.md`

**Step 1: Read current Stage 2 block**

Find the `**Stage 2 — LSP Semantic Check**` section in the file. It currently reads entries from a `## LSP Volatile Facts` block using the format `<label> : <operation> → <file:line:col> → <expected>`.

**Step 2: Replace Stage 2 with inline-first logic**

Replace the entire Stage 2 block (from `**Stage 2 — LSP Semantic Check**` through the closing `If all LSP facts pass` paragraph) with:

```markdown
**Stage 2 — LSP Semantic Check (runs after Stage 1 path check):**

For each arch file that references a changed package, check for LSP coordinates in priority order:

**Priority 1 — Inline coordinates** (present after lsp-enrich has run):
Scan the arch file for `<!-- lsp:hover:LINE:COL -->` and `<!-- lsp:refs:LINE:COL -->` annotations.
- For each `lsp:hover`: run LSP hover at that position, compare against the Go code block immediately below it.
  - Zero tolerance: any signature mismatch → flag stale
- For each `lsp:refs`: run LSP findReferences, count results excluding `_test.go` files, compare against `count=N` on the same line.
  - Allow ±15% drift: outside N×0.85 to N×1.15 → flag stale

**Priority 2 — Bottom block** (fallback for files not yet enriched):
If no inline coordinates found, check for a `## LSP Volatile Facts` block and parse entries using the existing format:
```
<label> : <operation> → <file:line:col> → <expected>
```
Apply same tolerance rules as Priority 1.

**Priority 3 — Plain text warning** (fallback when no LSP data):
If neither inline coords nor a bottom block found → emit original text warning:
```
⚠ Arch file may be stale:
  Changed:    business/sdk/workflow/temporal/trigger.go
  Tracked in: docs/arch/workflow.md (not updated in this commit)

Update the arch file now (include in this commit), or confirm it's still accurate.
[U] Update it   [S] Skip (it's still accurate)
```

**Stale report format (for Priority 1 and 2 failures):**
```
⚠ Arch file stale (LSP check):
  File:     docs/arch/workflow-engine.md:67
  Fact:     ActionHandler implementors
  Expected: count=20
  Actual:   count=23 (3 new handlers not documented)
  Action:   /arch-go update workflow-engine
```

If all LSP facts pass → proceed with ship, no action needed.
If any LSP fact fails → pause and report before staging.

**When to skip Stage 2:** If LSP server is not available or returns no results, skip and note:
```
ℹ Stage 2 LSP check skipped (LSP server unavailable) — verify arch facts manually
```
```

**Step 3: Verify**

Read the updated Stage 2 section — confirm it has Priority 1/2/3 fallback chain.

**Step 4: Commit**

```bash
git -C ~/.claude commit -am "feat(arch-go): update Mode B Stage 2 to check inline coords first"
```

---

## Task 3: arch-go — Update Mode C to refresh inline coordinates

**Files:**
- Modify: `~/.claude/skills/arch-go/SKILL.md`

**Step 1: Read Mode C**

Find the `## Mode C — Targeted Update` section. It has Steps 1–6.

**Step 2: Update Step 5**

Current Step 5 reads:
```
**Step 5:** Re-run all LSP operations in the `## LSP Volatile Facts` block (if present) and update the expected values with current counts/signatures. Append `(verified YYYY-MM-DD)` to any changed value.
```

Replace with:
```
**Step 5:** Refresh all LSP coordinates in the arch file:
- Scan for `<!-- lsp:hover:LINE:COL -->` annotations: re-run hover at each position. If the inlined code block below it changed, update it. If the line numbers shifted (due to file edits), update LINE:COL to match new positions.
- Scan for `<!-- lsp:refs:LINE:COL -->` annotations: re-run findReferences, update count=N if changed. Update LINE:COL if shifted.
- If a `## LSP Volatile Facts` block is still present (file not yet migrated): re-run those operations and update counts/signatures. Append `(verified YYYY-MM-DD)` to any changed value.
- Report all refreshed coordinates in the completion summary.
```

**Step 3: Verify**

Read Step 5 — confirm it handles both inline coords and bottom block fallback.

**Step 4: Commit**

```bash
git -C ~/.claude commit -am "feat(arch-go): update Mode C to refresh inline coordinates"
```

---

## Task 4: arch-go — Add Mode D section

**Files:**
- Modify: `~/.claude/skills/arch-go/SKILL.md`

**Step 1: Find insertion point**

Locate the `## Rules` section at the bottom of the file. Mode D goes immediately before it, after the existing `## Mode C` section.

**Step 2: Insert Mode D**

Add this entire section between Mode C and ## Rules:

```markdown
---

## Mode D — LSP Enrich

Run when user calls `/arch-go lsp-enrich [system]`. If no system arg → run across all `docs/arch/*.md`.

Goal: add compiler-verified type signatures and reference counts inline to eliminate future file reads. Only add what removes a future read.

**Step 1:** Read the target arch file(s).

**Step 2 — Interface enrichment:**
For each interface type appearing in a ⚠ callout or `key facts:` section (without an inlined definition):
- Run `documentSymbol` on the source file to find the interface declaration line:col
- Run `hover` at that position to get the full interface definition
- If the interface has 2+ methods → inline it as a Go code block
- Add `<!-- lsp:hover:LINE:COL -->` immediately above the code block

**Step 3 — Implementor counts:**
For each interface type just enriched (or already inlined):
- Run `findReferences` on the interface declaration position
- Count results, subtract those in `_test.go` files
- Add `<!-- lsp:refs:LINE:COL --> count=N (excl. test mocks)` on the line immediately after the closing code fence

**Step 4 — Key method signatures:**
For each method signature listed in a `key facts:` block that is not already hover-verified:
- Run `hover` on the method name position in its source file
- If the live signature differs from what's documented → update it inline
- Add `<!-- lsp:hover:LINE:COL -->` on the line immediately above the updated fact

**Step 5 — Cross-domain struct types:**
For each struct type listed that is consumed across domains (appears in a ⚠ callout):
- Run `hover` at the struct declaration to get its full field list
- If the struct has 3+ fields → inline it as a Go code block with `<!-- lsp:hover:LINE:COL -->`

**Step 6 — Migration (if `## LSP Volatile Facts` block present):**
Convert each bottom-block entry to inline placement:
1. Parse the entry: `label : operation → file:line:col → expected`
2. Re-run the stated LSP operation at the stated coordinates
3. Find the nearest ⚠ callout or `key facts:` line referencing that type/function
4. Insert inline annotation + code block immediately above that content
   - `findReferences` / `goToImplementation` entries → `<!-- lsp:refs:LINE:COL --> count=N`
   - `hover` entries → `<!-- lsp:hover:LINE:COL -->` + inlined block
   - `documentSymbol` count entries → `<!-- lsp:refs:LINE:COL --> count=N` (counts symbols, not refs)
5. After all entries migrated → remove the `## LSP Volatile Facts` block entirely

**Step 7 — Report:**
```
lsp-enrich complete: docs/arch/workflow-engine.md
  Added:   ActionHandler interface definition (2 methods)
  Added:   ActionHandler implementors count=20
  Added:   WorkflowInput struct (6 fields)
  Updated: Execute() signature (was stale)
  Migrated: 3 entries from ## LSP Volatile Facts → inline
  Skipped: WorkflowBus (1 exported method — not worth inlining)
  Missing: QueueClient not found via documentSymbol — manual review needed
```

**Skip conditions:**
- Interface with only 1 method and not in a ⚠ callout → skip
- Type already has an inlined block with a coordinate → skip (already enriched)
- LSP returns no result for a symbol → report as Missing, do not block
```

**Step 3: Verify**

Read the Mode D section — confirm Steps 1–7 are present and the migration step (Step 6) covers the `goToImplementation` → `lsp:refs` normalization.

**Step 4: Commit**

```bash
git -C ~/.claude commit -am "feat(arch-go): add Mode D (lsp-enrich)"
```

---

## Task 5: Update investigate/SKILL.md — LSP pre-step

**Files:**
- Modify: `~/.claude/skills/investigate/SKILL.md`

**Step 1: Read current step 2 + step 3**

Find `### 2. Identify the Entry Point` and `### 3. Spawn Parallel Investigation Agents`.

**Step 2: Insert LSP pre-step between steps 2 and 3**

After the step 2 block (which ends with `"The command or request that triggers it"`), insert a new step 2.5 before `### 3. Spawn Parallel Investigation Agents`:

```markdown
### 2.5. LSP Pre-Flight (Go projects only)

Before spawning agents, extract verified call chain data from LSP. This sharpens Agent 1's prompt with facts instead of letting it rediscover topology from file reads.

**If entry point is a specific known symbol (function or method name):**
```
1. Run outgoingCalls(entry_point) → what functions does this call?
2. Run incomingCalls(entry_point) → what calls this function?
```

**If the bug description mentions a specific broken symbol (missing method, wrong type):**
```
3. Run findReferences(broken_symbol) → exact list of call sites
```

Capture the results. Inject them into Agent 1's prompt as a `Known call chain (LSP-verified):` block (see Agent 1 template below).

**Guard:** If the entry point is vague ("something in the auth flow", "somewhere in the DB layer") — skip this step and proceed directly to spawning. Do NOT block on LSP when the symbol is ambiguous.

**If LSP is unavailable** → skip silently and proceed to spawning.
```

**Step 3: Update Agent 1 prompt template**

Find the Agent 1 prompt block (starting with `Prompt: "Starting from [ENTRY_POINT]`). Add a new section at the top of the prompt, before `For each step in the flow:`:

```
{{#if lsp_context}}
Known call chain (LSP-verified — use this as your starting map):
  [ENTRY_POINT] calls: [LSP_OUTGOING_CALLS]
  [ENTRY_POINT] is called by: [LSP_INCOMING_CALLS]
  {{#if broken_symbol}}[BROKEN_SYMBOL] referenced at: [LSP_REFERENCES]{{/if}}

Trace the data transformations through these verified call sites first.
{{/if}}
```

Replace `[LSP_OUTGOING_CALLS]`, `[LSP_INCOMING_CALLS]`, `[LSP_REFERENCES]` with actual LSP results when available.

**Step 4: Verify**

Read the updated skill — confirm step 2.5 is present between steps 2 and 3, and Agent 1's prompt has the LSP context block.

**Step 5: Commit**

```bash
git -C ~/.claude commit -am "feat(investigate): add LSP pre-flight before spawning agents"
```

---

## Task 6: Update fix-issue/SKILL.md — LSP pre-step

**Files:**
- Modify: `~/.claude/skills/fix-issue/SKILL.md`

**Step 1: Read current step 7**

Find `### 7. Parallel Investigation` — this is where agents are spawned. The LSP pre-step goes at the start of this section, before `Launch **all 3 agents simultaneously**`.

**Step 2: Insert LSP pre-step at the top of step 7**

After the `### 7. Parallel Investigation (use model="sonnet" for all agents)` heading and before `Launch **all 3 agents simultaneously**`, insert:

```markdown
**Before spawning — LSP Pre-Flight (Go projects only):**

If a specific entry point or broken symbol is known from the issue details, extract verified call chain data first:

```
If entry point is a specific known symbol:
  1. Run outgoingCalls(entry_point)
  2. Run incomingCalls(entry_point)

If issue mentions a specific broken symbol:
  3. Run findReferences(broken_symbol)
```

Inject results into Agent 1's prompt as `Known call chain (LSP-verified):` context. If entry point is vague or LSP unavailable → skip and spawn directly.

---
```

**Step 3: Update Agent 1 prompt in fix-issue**

Find the Agent 1 prompt block (starting with `Prompt: "Investigating issue: [ISSUE_TITLE]`). After the issue description lines and before `Starting from the likely entry point`, add:

```
{{#if lsp_context}}
Known call chain (LSP-verified — use this as your starting map):
  Entry point calls: [LSP_OUTGOING_CALLS]
  Entry point called by: [LSP_INCOMING_CALLS]
  {{#if lsp_refs}}Broken symbol referenced at: [LSP_REFERENCES]{{/if}}
{{/if}}

```

**Step 4: Verify**

Read the updated step 7 — confirm the LSP pre-flight block appears before the agent spawn instruction.

**Step 5: Commit**

```bash
git -C ~/.claude commit -am "feat(fix-issue): add LSP pre-flight before spawning agents"
```

---

## Task 7: Update spawn-agents/SKILL.md — LSP note

**Files:**
- Modify: `~/.claude/skills/spawn-agents/SKILL.md`

**Step 1: Read step 3 (model selection)**

Find `### 3. Select Model Per Agent` and the `#### Use 'haiku' when the subtask is:` block.

**Step 2: Add LSP rule to haiku list**

Find the last bullet in the haiku list (`- **Git operations** ...`). Add after it:

```markdown
- **Symbol lookup in Go projects** — if the task is finding all references or
  implementations of a *known* symbol, run LSP directly first (`findReferences`,
  `documentSymbol`, `incomingCalls`) before spawning. LSP answers in one operation
  what a haiku agent reconstructs from 5–10 file reads. If LSP returns results,
  use them directly and skip spawning. Only spawn if the symbol is vague or LSP
  returns nothing.
```

**Step 3: Add anti-pattern**

Find the `## Anti-Patterns — Do NOT Do These` section. Add after the last bullet (`- **Using general-purpose when Explore works**`):

```markdown
- **Spawning haiku to find references of a known Go symbol** — run
  `findReferences` or `incomingCalls` directly first; it's faster and exhaustive.
  Only spawn if LSP is unavailable or the symbol is ambiguous.
```

**Step 4: Verify**

Read the model selection section and anti-patterns section — confirm both additions are present.

**Step 5: Commit**

```bash
git -C ~/.claude commit -am "feat(spawn-agents): prefer LSP over haiku for known symbol lookups"
```

---

## Task 8: Migrate docs/arch/delegate.md

**Files:**
- Modify: `docs/arch/delegate.md`

**Context — current LSP Volatile Facts block:**
```
delegate.Call call sites : findReferences → business/sdk/delegate/delegate.go:48:21   → count=205
delegate.Call signature  : hover          → business/sdk/delegate/delegate.go:48:21   → sig="func (d *Delegate) Call(ctx context.Context, data Data) error"
```

**Step 1: Read the full file**

Read `docs/arch/delegate.md` — locate where `delegate.Call` appears in the `key facts:` section.

**Step 2: Run LSP operations**

```
1. LSP hover at business/sdk/delegate/delegate.go:48:21 → get live signature
2. LSP findReferences at business/sdk/delegate/delegate.go:48:21 → count results excl. _test.go
```

**Step 3: Insert inline annotations**

Find the `key facts:` line that mentions `Call(ctx, data)` or `delegate.Call`. Immediately above it, insert:

```markdown
<!-- lsp:hover:48:21 -->
```go
func (d *Delegate) Call(ctx context.Context, data Data) error
```
<!-- lsp:refs:48:21 --> count=N (excl. test mocks)
```

Replace `N` with the actual findReferences count from step 2. Update `48:21` if the live hover result came from a different line.

**Step 4: Remove bottom block**

Delete the entire `## LSP Volatile Facts (auto-checked by /arch-go staleness)` section and its fenced code block.

**Step 5: Verify**

Read the file — confirm inline coords present, bottom block gone.

**Step 6: Commit**

```bash
git add docs/arch/delegate.md
git commit -m "chore(arch): migrate delegate.md LSP facts to inline coordinates"
```

---

## Task 9: Migrate docs/arch/sqldb.md

**Files:**
- Modify: `docs/arch/sqldb.md`

**Context — current LSP Volatile Facts block:**
```
NamedQuerySlice call sites : findReferences → business/sdk/sqldb/sqldb.go:199:6 → count=129
```

**Step 1: Read the full file**

Read `docs/arch/sqldb.md` — locate the section mentioning `NamedQuerySlice`.

**Step 2: Run LSP operation**

```
LSP findReferences at business/sdk/sqldb/sqldb.go:199:6 → count results excl. _test.go
```

**Step 3: Insert inline annotation**

Find the `key facts:` line that mentions `NamedQuerySlice`. Immediately above it, insert:

```markdown
<!-- lsp:refs:199:6 --> count=N (excl. test mocks)
```

Replace `N` with the live count.

**Step 4: Remove bottom block**

Delete the `## LSP Volatile Facts` section.

**Step 5: Verify + Commit**

```bash
git add docs/arch/sqldb.md
git commit -m "chore(arch): migrate sqldb.md LSP facts to inline coordinates"
```

---

## Task 10: Migrate docs/arch/errs.md

**Files:**
- Modify: `docs/arch/errs.md`

**Context — current LSP Volatile Facts block:**
```
errs.New usages           : findReferences → app/sdk/errs/errs.go:61:6   → count=1065
errs.NewFieldsError usages: findReferences → app/sdk/errs/errs.go:129:6  → count=616
```

**Step 1: Read the full file**

Read `docs/arch/errs.md` — locate where `errs.New` and `errs.NewFieldsError` appear.

**Step 2: Run LSP operations**

```
1. LSP findReferences at app/sdk/errs/errs.go:61:6  → count excl. _test.go  (errs.New)
2. LSP findReferences at app/sdk/errs/errs.go:129:6 → count excl. _test.go  (errs.NewFieldsError)
```

**Step 3: Insert inline annotations**

Find the `key facts:` lines for `errs.New` and `errs.NewFieldsError`. Above each, insert:

```markdown
<!-- lsp:refs:61:6 --> count=N (excl. test mocks)
```
and
```markdown
<!-- lsp:refs:129:6 --> count=N (excl. test mocks)
```

Update line numbers if hover confirms different positions.

**Step 4: Remove bottom block + Commit**

```bash
git add docs/arch/errs.md
git commit -m "chore(arch): migrate errs.md LSP facts to inline coordinates"
```

---

## Task 11: Migrate docs/arch/auth.md

**Files:**
- Modify: `docs/arch/auth.md`

**Context — current LSP Volatile Facts block:**
```
permissionsbus.Business refs : findReferences → business/domain/core/permissionsbus/permissionsbus.go:35:6 → count=81
```

**Step 1: Read the full file**

Read `docs/arch/auth.md` — locate `permissionsbus.Business` in the content.

**Step 2: Run LSP operation**

```
LSP findReferences at business/domain/core/permissionsbus/permissionsbus.go:35:6 → count excl. _test.go
```

**Step 3: Insert inline annotation**

Find the line mentioning `permissionsbus.Business`. Above it, insert:

```markdown
<!-- lsp:refs:35:6 --> count=N (excl. test mocks)
```

**Step 4: Remove bottom block + Commit**

```bash
git add docs/arch/auth.md
git commit -m "chore(arch): migrate auth.md LSP facts to inline coordinates"
```

---

## Task 12: Migrate docs/arch/workflow-engine.md

**Files:**
- Modify: `docs/arch/workflow-engine.md`

**Context — current LSP Volatile Facts block:**
```
ActionHandler implementors   : goToImplementation → business/sdk/workflow/interfaces.go:39:6     → count=20 (production, excl. test mocks)
TriggerProcessor.Initialize  : hover              → business/sdk/workflow/trigger.go:88:29        → sig="func (tp *TriggerProcessor) Initialize(ctx context.Context) error"
EdgeType constants           : documentSymbol     → business/sdk/workflow/temporal/models.go      → names=[EdgeTypeStart,EdgeTypeSequence,EdgeTypeAlways] count=3
```

**Step 1: Read the full file**

Read `docs/arch/workflow-engine.md` — locate `ActionHandler`, `TriggerProcessor.Initialize`, and `EdgeType` in the content.

**Step 2: Run LSP operations**

```
1. LSP hover at business/sdk/workflow/interfaces.go:39:6         → ActionHandler interface definition
2. LSP findReferences at business/sdk/workflow/interfaces.go:39:6 → count implementors excl. _test.go
3. LSP hover at business/sdk/workflow/trigger.go:88:29           → TriggerProcessor.Initialize signature
4. LSP documentSymbol on business/sdk/workflow/temporal/models.go → find EdgeType constants
```

**Step 3: Insert inline annotations**

For `ActionHandler` — find its mention in the ⚠ callout or key facts. Insert:
```markdown
<!-- lsp:hover:39:6 -->
```go
type ActionHandler interface {
    [live interface definition from hover]
}
```
<!-- lsp:refs:39:6 --> count=N (excl. test mocks)
```

For `TriggerProcessor.Initialize` — find its `key facts:` line. Insert above:
```markdown
<!-- lsp:hover:88:29 -->
```go
func (tp *TriggerProcessor) Initialize(ctx context.Context) error
```
```

For `EdgeType constants` — find where EdgeTypeStart/EdgeTypeSequence/EdgeTypeAlways are listed. Insert:
```markdown
<!-- lsp:refs:LINE:COL --> names=[EdgeTypeStart,EdgeTypeSequence,EdgeTypeAlways] count=3
```
(Use the line:col of the EdgeType type declaration from documentSymbol.)

**Step 4: Remove bottom block + Commit**

```bash
git add docs/arch/workflow-engine.md
git commit -m "chore(arch): migrate workflow-engine.md LSP facts to inline coordinates"
```

---

## Task 13: Migrate docs/arch/agent-chat.md

**Files:**
- Modify: `docs/arch/agent-chat.md`

**Context — current LSP Volatile Facts block:**
```
toolcatalog constants : documentSymbol     → business/sdk/toolcatalog/toolcatalog.go              → count=53
Provider implementors : goToImplementation → business/sdk/llm/provider.go:12:6                    → count=3 (gemini/active, claude, ollama)
Embedder implementors : goToImplementation → business/sdk/toolindex/embedder.go:13:6              → count=2 (gemini, ollama; excl. test mocks)
```

**Step 1: Read the full file**

Read `docs/arch/agent-chat.md` — locate `toolcatalog`, `Provider`, and `Embedder` in the content.

**Step 2: Run LSP operations**

```
1. LSP documentSymbol on business/sdk/toolcatalog/toolcatalog.go → count exported constants
2. LSP hover at business/sdk/llm/provider.go:12:6         → Provider interface definition
3. LSP findReferences at business/sdk/llm/provider.go:12:6 → count implementors excl. _test.go
4. LSP hover at business/sdk/toolindex/embedder.go:13:6    → Embedder interface definition
5. LSP findReferences at business/sdk/toolindex/embedder.go:13:6 → count implementors excl. _test.go
```

**Step 3: Insert inline annotations**

For `toolcatalog constants` — find where toolcatalog constant count is mentioned. Insert:
```markdown
<!-- lsp:refs:LINE:COL --> count=N (exported constants, from documentSymbol)
```
(Use the line:col of the toolcatalog.go package declaration or first constant declaration.)

For `Provider` interface — find its ⚠ callout mention. Insert:
```markdown
<!-- lsp:hover:12:6 -->
```go
[live Provider interface definition]
```
<!-- lsp:refs:12:6 --> count=3 (gemini/active, claude, ollama; excl. test mocks)
```

For `Embedder` interface — find its ⚠ callout mention. Insert:
```markdown
<!-- lsp:hover:13:6 -->
```go
[live Embedder interface definition]
```
<!-- lsp:refs:13:6 --> count=2 (gemini, ollama; excl. test mocks)
```

**Step 4: Remove bottom block + Commit**

```bash
git add docs/arch/agent-chat.md
git commit -m "chore(arch): migrate agent-chat.md LSP facts to inline coordinates"
```

---

## Task 14: Final verification

**Step 1: Verify no bottom blocks remain**

```bash
grep -rl "LSP Volatile Facts" docs/arch/
```

Expected: empty output (all 6 files migrated).

**Step 2: Verify inline coords present in all migrated files**

```bash
grep -rl "lsp:hover\|lsp:refs" docs/arch/
```

Expected: all 6 migrated files listed.

**Step 3: Verify skill files updated**

```bash
grep -l "Mode D\|lsp-enrich" ~/.claude/skills/arch-go/SKILL.md
grep -l "LSP Pre-Flight" ~/.claude/skills/investigate/SKILL.md ~/.claude/skills/fix-issue/SKILL.md
grep -l "Symbol lookup in Go" ~/.claude/skills/spawn-agents/SKILL.md
```

Expected: each file listed.

**Step 4: Final commit**

```bash
git add docs/arch/
git commit -m "chore: verify LSP skill convergence complete — all 6 arch files migrated"
```
