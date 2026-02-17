# Ship Go Skill

Ship the current Go changes: build, targeted test, lint, commit, and push.

**IMPORTANT: NEVER run `go test ./...`. Always scope tests to changed packages only.**

## Steps

### 1. Build
Run `go build ./...` and fix any compile errors.

### 2. Targeted Tests

Only test packages that were actually changed.

**Determine affected packages:**
```bash
git diff HEAD --name-only -- '*.go' | xargs -I{} dirname {} | sort -u | sed 's|^|./|'
```

If no `.go` files changed, skip testing entirely.

**Run tests with JSON output and filter to failures only:**
```bash
go test -count=1 -timeout 120s -json <packages> 2>&1 | jq -r '
  select(.Action == "fail") |
  if .Test then "FAIL: \(.Package) / \(.Test) (\(.Elapsed)s)"
  else "FAIL: \(.Package) (\(.Elapsed)s)"
  end
'
```

- If the jq output is empty, all tests passed — report that and move on.
- If there are failures, get failure details for ONLY the failing packages:
```bash
go test -v -count=1 -timeout 120s -run 'TestName1|TestName2' ./failing/pkg 2>&1 | grep -v '^{' | head -150
```
  The `grep -v '^{'` strips structured JSON log lines (Ardan-style logging). The `head -150` caps output to preserve tokens while keeping enough for 3-4 failures with full stack traces. Fix failures and re-run the targeted test.

**Escape hatch:** If tests are taking too long (e.g. integration tests pulling in database containers), stop them and ask the user whether to skip, increase the timeout, or narrow the test scope further with `-run`.

### 3. Lint
Run `make lint` and fix any lint issues.

### 4. Commit and Push
- Stage all changed files with `git add` (prefer specific files over `-A`)
- Write a conventional commit message describing the changes
- Push to the current branch
