# Debugging Suite

Tools and workflows for diagnosing and fixing test failures and bugs in Ichor.

## Quick Reference — Which Tool to Use

| Situation | Tool | Command |
|---|---|---|
| Batch of test failures from CI | `/debug-go-tests` | Run `make test-extract ARGS="./..."` first |
| Single known bug file in `docs/bugs/open/` | `/pick-bug` | `/pick-bug` or `/pick-bug <filename>` |
| Unexpected behavior, no bug file yet | `/investigate` | Describe the symptom |
| Vue/frontend + Go full-stack bug | `/debug-vue` | Describe the symptom |
| Process accumulated completed bugs into patterns | `/distill-bugs` | Run after a batch of `/pick-bug` fixes |

## How the Tools Connect

```
CI test failures
     │
     ▼
make test-extract  →  .claude/debug/ files
     │
     ▼
/debug-go-tests  →  clusters failures, fixes batches
     │
     ▼  (individual bugs)
/pick-bug  →  claims docs/bugs/open/
     │         checks docs/debug-patterns/INDEX.md (Step 2.5)
     │         if pattern match: apply directly (3 steps instead of 9)
     │         if no match: full investigation
     │
     ▼
docs/bugs/complete/  (fixed, verified)
     │
     ▼
/distill-bugs  →  classifies (haiku) → matches/creates patterns (opus/sonnet)
     │
     ▼
docs/debug-patterns/  (pattern library)
docs/bugs/archive/    (provenance preserved)
```

## Debug Patterns (`docs/debug-patterns/`)

The pattern library short-circuits known recurring bugs from 9-step investigation to 3-step apply-and-verify.

- **`INDEX.md`**: flat lookup table — grep by test name, package, error keyword. Read in full at Step 2.5 of `/pick-bug`.
- **Pattern files**: terse signal/root-cause/fix recipes. Read the matching file to get exact fix steps.
- **Created by** `/distill-bugs`, **consumed by** `/pick-bug` Step 2.5.

### Current Patterns

| Pattern | Typical Signal | Common Packages |
|---|---|---|
| `table-access-count` | `len(gotResp.TableAccess) != N` | `permissionsbus` |
| `missing-table-access-seed` | 401 on ALL endpoints, new entity | any new domain |
| `pg-time-leading-zeros` | `"8:00:00"` vs `"08:00:00"` | any with time columns |
| `business-default-in-test` | status field empty in EXP, populated in GOT | any with default-setting business layer |
| `test-seed-fake-user-fk` | `foreign key violation`, `uuid.New()` in test seed | `actionhandlers`, workflow tests |
| `handler-nil-bus` | action handler 404 / nil panic | `actionapi`, `workflowactions` |
| `formconfig-value-column` | `DROPDOWN_COLUMN_NOT_FOUND` | `dbtest` |
| `omitempty-to-required` | create-400 returns 200 | any new domain app layer |

## Bug Queue (`docs/bugs/`)

- `open/` — unstarted bugs, claimed atomically via `mv`
- `in_progress/` — currently being worked on (one agent per file)
- `complete/` — fixed and verified, awaiting `/distill-bugs`
- `archive/` — distilled, provenance preserved

**Parallel safety**: `/pick-bug` uses atomic `mv` to claim bugs. Two agents cannot claim the same bug simultaneously.

## Generating Bug Files

```bash
make test-extract ARGS="./api/cmd/services/ichor/tests/..."
```

This writes failure files to `.claude/debug/`. Then `/debug-go-tests` reads and processes them, or move files manually to `docs/bugs/open/`.

## Arch Docs Integration

- `/pick-bug` reads `docs/arch/*.md` based on the failing package (Step 3)
- `docs/arch/seeding.md` is always read — most count/auth mismatches are seed-related
- After multi-iteration fixes, `/pick-bug` proposes `⚠` callouts to arch docs (Step 8)
- `/distill-bugs` promotes recurring fix patterns to `docs/debug-patterns/`

See `docs/arch/testing.md` for test infrastructure details.
