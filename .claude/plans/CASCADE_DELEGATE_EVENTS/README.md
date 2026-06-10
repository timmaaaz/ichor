# CASCADE_DELEGATE_EVENTS

Workflow actions that mutate the DB but emit **no delegate event**, so they cannot trigger other automation rules (no cascade). Decision: cascade IS wanted; the open problem is guarding against runtime cross-rule cascade loops.

- **Build roadmap (start here to execute):** [`PLAN.md`](./PLAN.md) — the overarching phased plan (core P0–P5 + committed/on-demand follow-up tracks F1–F7) with a don't-skip tracking table.
- **Facts:** [`INVESTIGATION.md`](./INVESTIGATION.md) — full map of all action handlers, the two failure mechanisms, the loop problem, design options, and the decision surface (verified facts + §13 Round 5 probe sweep).
- **Design:** [`DESIGN.md`](./DESIGN.md) — the solution-space exploration: two-time cycle detection (static prevent + runtime catch), value-aware edges, the single-manifest source-of-truth, policy A-vs-B, intentional loops. Decisions tagged `[DECIDED]`/`[LEANING]`/`[OPEN]`.
- **Write-path / validation:** [`WRITE_PATH.md`](./WRITE_PATH.md) — the *second* dimension (orthogonal to loops): generic raw-SQL actions skip validation, not just events. Holds the protected-field reality (~9 entities/~30 pairs), the bus-routing discovery (FormData is a generic entity→bus dispatcher), the domain-trace correction (bus-routing does NOT protect guarded fields), and the decided write-path core.
- **Follow-ups:** [`FOLLOW_UP.md`](./FOLLOW_UP.md) — deferred/downstream tracks intentionally kept OUT of the core plan: Option B bus-routing consolidation (~58-entity migration), the FormData tx bug, reliability hardening, PR #176 doc salvage, the frontend picker UX, and stale arch-doc cleanup.
- **Visual:** [`call-chains.html`](./call-chains.html) — open in a browser. Color-coded call chains: the working cascade, the two break points, the loop, and where every fix intervenes. Plaintext version: [`CALL_CHAINS.md`](./CALL_CHAINS.md).
- **Origin:** PR #176 (`docs/update-field-no-cascade`) documented the `update_field` case in isolation; this investigation found it is systemic (~8 handlers) and that PR #176's recommended `allocation_results` workaround is itself broken.
- **Branch:** `feature/cascade-delegate-events` (base `master`).
- **State:** investigation complete; **design fully decided** (DESIGN §10 ledger — loop guard + write-path core all `[DECIDED]`). Core plan = **loop guard → M2 (`delegate.Call`) → M1-synthesize → protected-list**. Deferred/downstream tracks (Option B bus-routing, etc.) in [`FOLLOW_UP.md`](./FOLLOW_UP.md).
- **Evidence base:** 10 parallel probes run 2026-06-09/10 (INVESTIGATION §13 Round 5). Key corrections: cascade-map endpoint DOES consume the manifest; 0 real loops in seeded rules; FormData bus-dispatcher exists but does NOT protect guarded fields (domain-trace Verdict a); `update_field` FE config is free-text.

Next: execute **[`PLAN.md`](./PLAN.md)** — core P0→P5 (loop guard P1+P2 verified before the P4 gate, per DESIGN §0); mirror into a `PROGRESS.yaml` when build starts. Committed follow-ups (F1 bus-routing, **F2 outbox**) are scheduled, not optional.
