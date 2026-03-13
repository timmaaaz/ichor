# Bug Queue

File-based work queue for parallel agent bug fixing.

## Directory States

| Directory | Meaning |
|-----------|---------|
| `open/` | Unstarted — any agent can claim |
| `in_progress/` | Claimed by an agent currently working on it |
| `complete/` | Fixed and verified — awaiting distillation |
| `archive/` | Distilled — provenance preserved, not actively referenced |

## Claiming a Bug (for agents)

**Claim atomically** by moving the file — `mv` on the same filesystem is atomic,
so two agents cannot claim the same bug simultaneously:

```bash
mv .claude/bugs/open/my-bug.md .claude/bugs/in_progress/my-bug.md
```

If the `mv` fails, another agent claimed it first — pick a different bug.

**When done:**
```bash
mv .claude/bugs/in_progress/my-bug.md .claude/bugs/complete/my-bug.md
```

## Bug File Format

Filename: `{slug}.md` — short, lowercase, hyphenated description.
Example: `contactinfos-time-format.md`, `warehouse-count-mismatch.md`

```markdown
---
id: {slug}
priority: high | medium | low
package: github.com/timmaaaz/ichor/...
reported: YYYY-MM-DD
---

## Description

One paragraph: what is broken and how it manifests.

## Reproduction

Steps or test command to reproduce:
```bash
go test -v -run Test_Foo/bar ./api/cmd/services/ichor/tests/...
```

## Expected

What should happen.

## Actual

What actually happens (include DIFF/GOT/EXP if from a test failure).

## Notes

Any relevant context, related files, or suspected root cause.
```

## Workflow

1. Drop a bug file in `open/`
2. Agent claims it by moving to `in_progress/`
3. Agent investigates, fixes, runs the affected tests
4. On verified fix, agent moves to `complete/` and appends a `## Fix` section
5. (Optional) Agent adds a `⚠` callout to the relevant `docs/arch/*.md`
