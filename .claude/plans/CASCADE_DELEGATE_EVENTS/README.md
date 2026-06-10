# CASCADE_DELEGATE_EVENTS

Workflow actions that mutate the DB but emit **no delegate event**, so they cannot trigger other automation rules (no cascade). Decision: cascade IS wanted; the open problem is guarding against runtime cross-rule cascade loops.

- **Start here:** [`INVESTIGATION.md`](./INVESTIGATION.md) — full map of all action handlers, the two failure mechanisms, the loop problem, design options, and the decision surface (verified facts).
- **Design:** [`DESIGN.md`](./DESIGN.md) — the solution-space exploration: two-time cycle detection (static prevent + runtime catch), value-aware edges, the single-manifest source-of-truth, policy A-vs-B, intentional loops. Decisions tagged `[DECIDED]`/`[LEANING]`/`[OPEN]`.
- **Visual:** [`call-chains.html`](./call-chains.html) — open in a browser. Color-coded call chains: the working cascade, the two break points, the loop, and where every fix intervenes. Plaintext version: [`CALL_CHAINS.md`](./CALL_CHAINS.md).
- **Origin:** PR #176 (`docs/update-field-no-cascade`) documented the `update_field` case in isolation; this investigation found it is systemic (~8 handlers) and that PR #176's recommended `allocation_results` workaround is itself broken.
- **Branch:** `feature/cascade-delegate-events` (base `master`).
- **State:** investigation complete; design not started.

Next chat should pick up at INVESTIGATION.md §8 (design options) / §9 (open questions).
