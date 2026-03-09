# LSP Skill Convergence Design

**Date:** 2026-03-09
**Scope:** arch-go, investigate, fix-issue, spawn-agents

## Problem

arch-vue is ahead of arch-go on LSP maturity. arch-vue has Mode D (lsp-enrich), inline
`<!-- lsp:hover:LINE:COL -->` coordinates, and a hover-based staleness check. arch-go has
a `## LSP Volatile Facts` bottom block and Stage 2 check — a rougher version of the same idea.
Additionally, investigate, fix-issue, and spawn-agents don't use LSP for symbol lookups,
falling back to file reads that reconstruct what LSP can answer in one call.

## Goals

1. Add Mode D to arch-go (matching arch-vue's lsp-enrich mode, Go-specific targets)
2. Replace arch-go's bottom `## LSP Volatile Facts` block with inline coordinates
3. Migrate existing Ichor arch files to inline coordinates
4. Add LSP pre-steps to investigate and fix-issue before Agent 1 spawns
5. Add LSP-first rule to spawn-agents model selection

## Design

### Section 1: arch-go Mode D (lsp-enrich)

Triggered by `/arch-go lsp-enrich [system]`. If no system arg, runs across all `docs/arch/*.md`.

**Enrichment targets (Go-specific):**

| Target | LSP operation | Output |
|--------|--------------|--------|
| Interface type in a ⚠ callout | `documentSymbol` → find declaration → `hover` | Inline Go code block + `<!-- lsp:hover:LINE:COL -->` |
| Interface implementor count | `findReferences` on the interface type | `<!-- lsp:refs:LINE:COL --> count=N (excl. test mocks)` |
| Key method signature in `key facts:` | `hover` on the method name | Updated signature + `<!-- lsp:hover:LINE:COL -->` |
| Struct types listed cross-domain | `hover` at struct declaration | Inline struct definition as Go block |

**Coordinate placement rule:** Comment goes immediately above the inlined code block.

```markdown
<!-- lsp:hover:42:6 -->
\`\`\`go
type ActionHandler interface {
    Execute(ctx context.Context, input ActionInput) (ActionOutput, error)
    Name() string
}
\`\`\`
<!-- lsp:refs:42:6 --> count=20 (excl. test mocks)
```

**Completion report format:**
```
lsp-enrich complete: docs/arch/workflow-engine.md
  Added:   ActionHandler interface definition (2 methods)
  Added:   ActionHandler implementors count=20
  Added:   WorkflowInput struct (6 fields)
  Updated: Execute() signature (was stale)
  Skipped: WorkflowBus (simple wrapper — 1 exported method, not worth inlining)
  Missing: QueueClient not found via documentSymbol — manual review needed
```

**Skip condition:** If a type has only 1 exported method or field and is not referenced in a ⚠
callout, skip it — not worth inlining.

---

### Section 2: Mode B Stage 2 Updates

Mode B Stage 2 checks inline coordinates first, falls back to the bottom block, then falls back
to plain text warning.

**Updated Stage 2 logic:**
```
For each arch file that references a changed package:
  1. Scan for inline <!-- lsp:hover:LINE:COL --> and <!-- lsp:refs:LINE:COL --> coordinates
  2. For each lsp:hover → run hover, compare inlined block below it (zero tolerance)
  3. For each lsp:refs  → run findReferences, count (excl. _test.go), compare to count=N (±15%)
  4. If no inline coordinates → fall back to ## LSP Volatile Facts block (existing Stage 2 logic)
  5. If neither → fall back to plain text "may be stale" warning
```

**Stale report (updated to show line number):**
```
⚠ Arch file stale (LSP check):
  File:       docs/arch/workflow-engine.md:67
  Fact:       ActionHandler implementors
  Expected:   count=20
  Actual:     count=23 (3 new handlers not documented)
  Action:     /arch-go update workflow-engine
```

---

### Section 3: Mode C Updates + Migration

**Mode C additions:**
- After diffing source files, refresh any inline coordinates pointing to shifted lines
- Run `hover`/`findReferences` at updated positions, update `LINE:COL` values + inlined content
- Report refreshed coordinates in the completion summary

**Migration (runs once as part of implementation):**
1. Find all `docs/arch/*.md` with a `## LSP Volatile Facts` block
2. For each entry in the block (format: `label : operation → file:line:col → expected`):
   a. Re-run the LSP operation at stated coordinates to get fresh results
   b. Find the nearest ⚠ callout or `key facts:` line that references that type
   c. Insert inline annotation + code block immediately above that content
3. Remove the `## LSP Volatile Facts` block after all entries are migrated

---

### Section 4: investigate + fix-issue LSP Pre-Steps

After identifying the entry point and *before* spawning agents, run:

```
If entry point is a known symbol (function/method):
  1. outgoingCalls(entry_point) → what does this function call?
  2. incomingCalls(entry_point) → who calls this function?

If there is a broken/missing symbol in the bug description:
  3. findReferences(broken_symbol) → exact caller list

Inject into Agent 1 prompt as verified context:
  "Known call chain (LSP-verified):
     entry_point calls: [A, B, C] at [file:line refs]
     entry_point is called by: [X, Y] at [file:line refs]
     broken_symbol is referenced at: [file:line refs]
  Use this as your starting map — trace the data transformations through these."
```

**Guard:** If the entry point is vague (not a specific symbol), skip the pre-step and spawn directly.

**Files to update:** `~/.claude/skills/investigate/SKILL.md`, `~/.claude/skills/fix-issue/SKILL.md`

---

### Section 5: spawn-agents LSP Note

**New rule under "Use `haiku` when the subtask is:":**

```
- **Symbol lookup in Go projects** — if the task is finding all references or
  implementations of a *known* symbol, run LSP directly first (findReferences,
  documentSymbol, incomingCalls) before spawning. LSP answers in one operation
  what a haiku agent reconstructs from 5-10 file reads. Only spawn if the
  symbol is vague or LSP returns nothing.
```

**New anti-pattern entry:**

```
- **Spawning haiku to find references of a known Go symbol** — run
  `findReferences` or `incomingCalls` directly; it's faster and exhaustive.
  Spawn only if LSP is unavailable or the symbol is ambiguous.
```

**File to update:** `~/.claude/skills/spawn-agents/SKILL.md`

---

## Implementation Order

1. Update `arch-go/SKILL.md` — add Mode D, update Mode B Stage 2, update Mode C
2. Update `investigate/SKILL.md` — add LSP pre-step to step 2/3
3. Update `fix-issue/SKILL.md` — add LSP pre-step to step 7
4. Update `spawn-agents/SKILL.md` — add LSP note to model selection + anti-patterns
5. Migrate existing Ichor `docs/arch/*.md` files — convert bottom blocks to inline coordinates

## Files Changed

| File | Change |
|------|--------|
| `~/.claude/skills/arch-go/SKILL.md` | Add Mode D, update Mode B Stage 2 + Mode C |
| `~/.claude/skills/investigate/SKILL.md` | Add LSP pre-step before spawning |
| `~/.claude/skills/fix-issue/SKILL.md` | Add LSP pre-step before spawning |
| `~/.claude/skills/spawn-agents/SKILL.md` | Add LSP note to model selection + anti-patterns |
| `docs/arch/*.md` (Ichor files with LSP Volatile Facts) | Migrate bottom blocks to inline coords |
