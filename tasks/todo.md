# LSP × Arch Docs Enrichment Plan

**Goal:** Make arch docs drift-resistant by (1) burning LSP-verified exact facts into the doc bodies,
(2) adding targeted LSP hints to high-stakes ⚠ callouts, and (3) upgrading the `/arch-go`
staleness skill's Mode B to use semantic LSP checks instead of grep-only matching.

**Decision from debate:** Volatile facts (counts, signatures, implementor lists) get LSP-verified
and date-stamped in the docs. LSP hints go *only* in ⚠ callouts where stale data causes wrong plans.
Doc bodies stay compact prose — no LSP instructions scattered through narrative text.

---

## Phase 1 — LSP Verification Sweep

Run LSP operations to collect exact facts. Results feed directly into Phase 2 edits.

### 1a. Reference counts (replace all "~X" approximations)

- [ ] `findReferences` on `delegate.go → Call` method — replaces "~198 call sites" in `delegate.md`
- [ ] `findReferences` on `sqldb.go → NamedQuerySlice` — replaces "185+ importers" in `sqldb.md`
- [ ] `findReferences` on `errs.go → NewFieldsError` — replaces "~203 files" in `errs.md`
- [ ] `findReferences` on `errs.go → New` — replaces "496+ importers" in `errs.md`
- [ ] `findReferences` on `permissionsbus.go → Business struct` — replaces "89 files" in `auth.md`
- [ ] `findReferences` on `executor.go → Execute method` — verify "55 tool handlers" in `agent-chat.md`

### 1b. Interface implementors (add verified lists to docs)

- [ ] `goToImplementation` on `interfaces.go → ActionHandler` — all 20 handler file paths for `workflow-engine.md`
- [ ] `goToImplementation` on `llm/ → Provider interface` — confirm implementors for `agent-chat.md`
- [ ] `goToImplementation` on `toolindex.go → Embedder interface` — confirm implementors for `agent-chat.md`
- [ ] `goToImplementation` on `{entity}bus → Storer interface` (use orderbus as reference) — confirm pattern for `domain-template.md`

### 1c. Critical signature verification (the "burned us" set from memory)

- [ ] `hover` on `delegate.go → Call` — verify exact signature (Call not Raise)
- [ ] `hover` on `trigger.go → TriggerProcessor → Initialize` — verify method name (Initialize not LoadRules)
- [ ] `documentSymbol` on `workflow/temporal/models.go` — verify edge type constants (only: start, sequence, always)
- [ ] `hover` on `graph_executor.go → Execute` — verify GraphExecutor.Execute signature
- [ ] `hover` on `workflowsaveapp/model.go` — verify allowed edge type validation list

---

## Phase 2 — Enrich Doc Bodies

Apply Phase 1 results to each arch file. Rules:
- Replace `~N` / `N+` with exact number: `N (verified YYYY-MM-DD)`
- Add `Implementors` subsection under interfaces that have concrete types worth listing
- Annotate 3 critical signatures with `// ✓ verified YYYY-MM-DD` comment

### Per-file edits

- [ ] `delegate.md` — update call site count; annotate `Call()` signature
- [ ] `sqldb.md` — update importer count for `NamedQuerySlice`
- [ ] `errs.md` — update both importer counts (`New` and `NewFieldsError`)
- [ ] `auth.md` — update permissionsbus importer count; annotate `bcryptCost = 12`
- [ ] `workflow-engine.md` — add `Implementors` subsection under ActionHandler interface with all 20 handler file paths; annotate `Initialize()` signature; confirm edge type constant list
- [ ] `agent-chat.md` — update tool handler count; add `Implementors` subsection under Provider and Embedder

---

## Phase 3 — Add LSP Hints to ⚠ Callouts

**Rule:** LSP hints go only where stale data directly causes a wrong implementation plan.
Four callouts qualify. Format: append a `verify:` line after the callout file list.

- [ ] `delegate.md → ⚠ Changing Data struct shape`
  ```
  verify: findReferences(business/sdk/delegate/delegate.go, Call) — confirm exact call site count before mass edit
  ```

- [ ] `errs.md → ⚠ Changing FieldError struct shape`
  ```
  verify: findReferences(app/sdk/errs/errs.go, NewFieldsError) — exact count of callers before mass edit
  ```

- [ ] `workflow-engine.md → ⚠ Adding a new ActionHandler`
  ```
  verify: goToImplementation(business/sdk/workflow/interfaces.go, ActionHandler) — confirm existing 20 implementors; register new one alongside them in all.go
  ```

- [ ] `agent-chat.md → ⚠ Adding a new agent tool`
  ```
  verify: documentSymbol(business/sdk/toolcatalog/toolcatalog.go) — confirm constant is in correct group (GroupWorkflow vs GroupTables) before wiring Executor
  ```

---

## Phase 4 — Upgrade `/arch-go` Staleness Skill (Mode B)

Current Mode B: grep changed file paths → check if referenced arch file was also changed.
Enhanced Mode B: grep PLUS semantic LSP spot-checks on volatile facts.

### 4a. Add a volatile facts index to each arch doc

Each arch doc gets a new fenced block at the bottom:

```
## LSP Volatile Facts (auto-checked by /arch-go staleness)
delegate.Call          : findReferences → business/sdk/delegate/delegate.go:21:1   → count=202
NamedQuerySlice        : findReferences → business/sdk/sqldb/sqldb.go:47:1         → count=187
ActionHandler impls    : goToImplementation → business/sdk/workflow/interfaces.go:12:1 → count=20
TriggerProcessor.Init  : hover → business/sdk/workflow/temporal/trigger.go:XX:1    → sig="Initialize() error"
```

Format: `label : operation → file:line:col → expected=value`

- [ ] Add volatile facts blocks to: `delegate.md`, `sqldb.md`, `errs.md`, `auth.md`, `workflow-engine.md`, `agent-chat.md`

### 4b. Upgrade Mode B in `~/.claude/skills/arch-go/SKILL.md`

Replace the current grep-only staleness steps with a two-stage check:

**Stage 1 (unchanged):** grep changed file paths → find candidate arch files

**Stage 2 (new — LSP semantic check):** For each candidate arch file that has a
`## LSP Volatile Facts` block, parse each entry and run the stated LSP operation.
Compare result against `expected=value`:
- Count within ±15%: pass
- Count outside ±15%: flag as stale with specific fact: "ActionHandler impls: doc says 20, LSP says 23 — arch file is stale"
- Signature mismatch: always flag (zero tolerance on wrong method names)

Report format:
```
⚠ Arch file stale (LSP check):
  File:     docs/arch/workflow-engine.md
  Fact:     ActionHandler impls
  Expected: count=20
  Actual:   count=23 (3 new handlers not documented)
  Action:   /arch-go update workflow-engine
```

- [ ] Update `~/.claude/skills/arch-go/SKILL.md` — add Stage 2 LSP check to Mode B
- [ ] Update Mode C (Targeted Update) to re-run LSP volatile facts and update the index block

---

## Phase 5 — Verification

- [ ] Read each updated arch file — confirm volatile facts blocks are parseable, counts look right
- [ ] Dry-run Mode B staleness logic mentally against one changed file — trace both stages
- [ ] Confirm CLAUDE.md `⚠ callouts are complete` directive still holds after Phase 3 additions
- [ ] `go build ./...` not required (no Go code changed) — but confirm arch-go skill syntax is valid

---

## Review

All phases complete (2026-03-09).

**Phase 1 — LSP Verification Sweep: DONE**
Actual counts found (vs. doc claims):
- delegate.Call: 205 call sites / 65 files (was "~198")
- NamedQuerySlice: 129 call sites / 75 files (was "185+ importers" — different metric, now clarified)
- errs.New: 1,065 usages / 151 files (was "496+ importers")
- errs.NewFieldsError: 616 usages / 138 files (was "~203 files")
- permissionsbus.Business: 81 files (was "89 files")
- toolcatalog: 53 constants (was "55" — actual drift found!)
- ActionHandler: 20 production implementors + 3 test mocks + 1 async adapter ✓
- Provider interface: 3 implementors (gemini/active, claude, ollama) — undocumented!
- delegate.Call sig: ✓ confirmed "Call" not "Raise"
- TriggerProcessor.Initialize sig: ✓ confirmed "Initialize" not "LoadRules"
- Edge types: ✓ confirmed only 3 (start, sequence, always)

**Phase 2 — Doc body enrichment: DONE**
6 arch files updated with exact, dated counts.

**Phase 3 — ⚠ LSP hints: DONE**
4 callouts updated: delegate.md, errs.md, workflow-engine.md, agent-chat.md.

**Phase 4 — arch-go skill upgrade: DONE**
Mode B: Stage 2 LSP semantic check added.
Mode C: Step 5 now re-runs volatile facts on targeted update.

---

## Lessons Applied

- Debate conclusion: stable knowledge (pipeline diagrams, behavioral invariants) stays in docs;
  volatile facts (counts, signatures, implementors) get LSP-verified and dated
- LSP hints belong only in ⚠ callouts, never scattered in prose
- Mode B staleness check is the force multiplier — fixes drift permanently vs. one-time patch
- "Initialize() not LoadRules()" class of bug is now caught by signature mismatch check in Mode B
