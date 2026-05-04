# ZPL Template Review Follow-ups Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Address three findings from PR #143's code review on the location/tote ZPL resize: (1) render-safe Code length bound at the validator layer, (2) test pin for the deliberate Tote≡Location byte-identical body invariant, and (3) cosmetic comment correction.

**Architecture:** Validator-level guard in `app/domain/labels/labelapp/model.go` tightens `Code` upper bound from `max=32` to `max=12` (the practical Code128/BY4 alphanumeric capacity for the 4"×6" canvas). Failing loud at API ingress is preferred over template-level defensive truncation, which would produce labels whose visible text doesn't match the barcode encoding. Schema column stays `VARCHAR(32)` for future media flexibility. Identity invariant pinned by a small unit test in `business/domain/labels/labelbus/zpl/zpl_test.go`.

**Tech Stack:** Go 1.23, `validator/v10` tags via `app/sdk/errs.Check`, pure-function ZPL string-builders.

---

## File Structure

| File | Action | Responsibility |
|------|--------|----------------|
| `app/domain/labels/labelapp/model.go` | Modify | Tighten `Code` validator tags + document the dimensional rationale |
| `app/domain/labels/labelapp/validate_test.go` | Modify | Add boundary tests for the new `Code` length cap |
| `business/domain/labels/labelbus/zpl/zpl_test.go` | Modify | Add identity-invariant test + max-safe-code snapshot test |
| `business/domain/labels/labelbus/zpl/location.go` | Modify | Comment cleanup (line 13) |

No new files. No schema migration. No business-layer change. No frontend change.

---

## Pre-Flight

- [ ] **Pre-Flight Step 1: Confirm working tree is clean and on master**

Run:
```bash
git status && git log -1 --oneline
```
Expected: `nothing to commit, working tree clean` and HEAD = `19642003 fix(labels): resize location/tote ZPL templates to 4"×6" GK420t media (#143)`.

- [ ] **Pre-Flight Step 2: Cut a feature branch**

Run:
```bash
git checkout -b fix/labels-zpl-review-followups
```
Expected: `Switched to a new branch 'fix/labels-zpl-review-followups'`.

---

## Task 1: Pin the Tote≡Location body identity invariant

**Files:**
- Modify: `business/domain/labels/labelbus/zpl/zpl_test.go` (append new test below `Test_Tote_Snapshot`)

**Why:** PR #143 deliberately kept `Tote()` and `Location()` byte-identical in body per the "identical body to Location at Phase 0b" intent. Today that promise is enforced only by a code comment. A future contributor running a "DRY this up" refactor or accidentally drifting one template can silently break the invariant. Adding a test converts the comment promise into a checked one. When Phase 1+ legitimately diverges them, the test fails clearly and the maintainer updates it intentionally.

- [ ] **Step 1: Append the identity test**

Add this test function at the end of `business/domain/labels/labelbus/zpl/zpl_test.go` (after `Test_Product_Snapshot_NilLot`):

```go
// Test_Tote_Body_Identical_To_Location pins the deliberate Phase 0b
// invariant that Tote and Location produce byte-identical ZPL when
// given the same code. Per tote.go's package comment, the two
// templates diverge in Phase 1+ (lot-expiry/icon fields on totes);
// when that lands, this test must be deleted in the same commit
// that diverges them.
func Test_Tote_Body_Identical_To_Location(t *testing.T) {
	codes := []string{"STG-A02", "TOTE-007", "X", "12-CHAR-CODE"}
	for _, c := range codes {
		loc := zpl.Location(zpl.LocationData{Code: c})
		tote := zpl.Tote(zpl.ToteData{Code: c})
		if loc != tote {
			t.Fatalf("Tote/Location body drift for code %q.\nlocation:\n%q\ntote:\n%q\n",
				c, loc, tote)
		}
	}
}
```

- [ ] **Step 2: Run the new test and verify it passes**

Run:
```bash
go test -run Test_Tote_Body_Identical_To_Location ./business/domain/labels/labelbus/zpl/ -v
```
Expected: `--- PASS: Test_Tote_Body_Identical_To_Location` (the templates are already identical post-#143).

- [ ] **Step 3: Verify the test catches drift (sanity check, do not commit)**

Temporarily edit `business/domain/labels/labelbus/zpl/tote.go` line 16, change `^FO40,300` to `^FO40,301`. Run:
```bash
go test -run Test_Tote_Body_Identical_To_Location ./business/domain/labels/labelbus/zpl/
```
Expected: `--- FAIL` with the diff printed.
Then revert the tote.go edit:
```bash
git checkout -- business/domain/labels/labelbus/zpl/tote.go
```
Expected: `git status` shows tote.go unmodified, only zpl_test.go modified.

- [ ] **Step 4: Commit**

```bash
git add business/domain/labels/labelbus/zpl/zpl_test.go
git commit -m "test(labels): pin Tote≡Location body identity invariant

Adds Test_Tote_Body_Identical_To_Location asserting that the two
templates produce byte-identical output for the same code. Pins the
deliberate Phase 0b invariant documented in tote.go's package
comment so a future drift (accidental refactor, copy-paste error)
fails loud rather than silently producing two different labels.

When Phase 1+ legitimately diverges the templates, this test must
be deleted in the same commit that diverges them."
```

---

## Task 2: Add render-safe `Code` length guard at the validator layer

**Files:**
- Modify: `app/domain/labels/labelapp/model.go:73` (NewLabel.Code tag)
- Modify: `app/domain/labels/labelapp/model.go:103` (UpdateLabel.Code tag)
- Modify: `app/domain/labels/labelapp/model.go:168` (RenderPrintRequest.Code tag)
- Modify: `app/domain/labels/labelapp/validate_test.go` (add boundary tests)

**Why:** PR #143's review surfaced that codes >12 chars overflow the Code128/BY4 barcode at the right edge of the 4"×6" canvas (812 dots wide, 40-dot margin, ~140-dot start/check/stop overhead, ~80-dot quiet zones, 44 dots/char → ~12-char ceiling for mixed alphanumerics). The DB schema (`VARCHAR(32)`) and current validator (`max=32`) both permit codes the template can't render correctly. Tightening the validator to `max=12` rejects unrenderable input at API ingress instead of silently producing a label whose barcode is clipped at print time. Verified safe to tighten: the longest seeded `code` in `business/sdk/dbtest/seed_labels.go` is 7 chars; `labelbus.TestNewLabels` generates `TEST-9999` (9 chars max).

- [ ] **Step 1: Audit existing seed data for codes >12 chars**

Run:
```bash
grep -rn 'code:[[:space:]]*"[^"]\{13,\}"\|Code:[[:space:]]*"[^"]\{13,\}"' \
  /Users/jaketimmer/src/work/superior/ichor/ichor/business \
  /Users/jaketimmer/src/work/superior/ichor/ichor/app | grep -i label
```
Expected: zero output. (If non-empty, those seeds need updating before tightening the validator — the plan needs a follow-up step to truncate them. Pre-research at plan-write time confirmed zero hits, but re-verify in case of intervening commits.)

- [ ] **Step 2: Write the failing validator test for over-cap codes**

Append to `app/domain/labels/labelapp/validate_test.go`:

```go
func Test_NewLabel_Validate_CodeTooLong(t *testing.T) {
	req := labelapp.NewLabel{
		Code: "WAREHOUSE-RECEIVING-DOCK-12A", // 28 chars, schema-allowed but not renderable
		Type: labelbus.TypeLocation,
	}
	if err := req.Validate(); err == nil {
		t.Fatal("expected validation error for 28-char code, got nil")
	}
}

func Test_NewLabel_Validate_CodeAtRenderableCap(t *testing.T) {
	req := labelapp.NewLabel{
		Code: "STG-A01-B12C", // 12 chars, the BY4/812-dot upper bound
		Type: labelbus.TypeLocation,
	}
	if err := req.Validate(); err != nil {
		t.Fatalf("expected 12-char code to validate, got: %v", err)
	}
}

func Test_RenderPrintRequest_Validate_CodeTooLong(t *testing.T) {
	req := labelapp.RenderPrintRequest{
		Type: labelbus.TypeLocation,
		Code: "WAREHOUSE-RECEIVING-DOCK-12A",
	}
	if err := req.Validate(); err == nil {
		t.Fatal("expected validation error for 28-char code, got nil")
	}
}
```

- [ ] **Step 3: Run the new tests, expect the "too long" cases to FAIL (validator still says max=32)**

Run:
```bash
go test -run "Test_NewLabel_Validate_CodeTooLong|Test_NewLabel_Validate_CodeAtRenderableCap|Test_RenderPrintRequest_Validate_CodeTooLong" ./app/domain/labels/labelapp/ -v
```
Expected:
- `Test_NewLabel_Validate_CodeTooLong` — **FAIL** ("expected validation error for 28-char code, got nil")
- `Test_NewLabel_Validate_CodeAtRenderableCap` — **PASS**
- `Test_RenderPrintRequest_Validate_CodeTooLong` — **FAIL**

- [ ] **Step 4: Tighten the three validator tags + add explanatory comments**

Edit `app/domain/labels/labelapp/model.go`:

Replace line 73 (NewLabel.Code):
```go
	Code        string `json:"code" validate:"required,min=1,max=32"`
```
with:
```go
	// Code length max=12 reflects the practical render budget on the
	// established Zebra GK420t 4"×6" media: ≈12 alphanumeric chars
	// fit a Code128 BY4 barcode at 812-dot width after start/check/stop
	// (~140 dots) + quiet zones (~80 dots) at 44 dots/char. Schema
	// column is VARCHAR(32) for future media flexibility; the
	// validator is tightened here so we fail loud at ingress instead
	// of silently producing labels with clipped barcodes.
	Code        string `json:"code" validate:"required,min=1,max=12"`
```

Replace line 103 (UpdateLabel.Code) — same `max=12` reasoning, abbreviated comment to avoid duplication:
```go
	Code        *string `json:"code,omitempty" validate:"omitempty,min=1,max=32"`
```
with:
```go
	// Code: see NewLabel.Code for the max=12 rationale.
	Code        *string `json:"code,omitempty" validate:"omitempty,min=1,max=12"`
```

Replace line 168 (RenderPrintRequest.Code):
```go
	Code    string          `json:"code,omitempty" validate:"omitempty,max=32"`
```
with:
```go
	// Code: see NewLabel.Code for the max=12 rationale.
	Code    string          `json:"code,omitempty" validate:"omitempty,max=12"`
```

- [ ] **Step 5: Run the new tests, expect all three to PASS**

Run:
```bash
go test -run "Test_NewLabel_Validate_CodeTooLong|Test_NewLabel_Validate_CodeAtRenderableCap|Test_RenderPrintRequest_Validate_CodeTooLong" ./app/domain/labels/labelapp/ -v
```
Expected: all three `--- PASS`.

- [ ] **Step 6: Run the full labelapp test suite to confirm no regressions**

Run:
```bash
go test ./app/domain/labels/...
```
Expected: `ok` for all packages. If any existing test seeds a code >12 chars, it will fail here; in that case fix the seed inline (the audit in Step 1 confirms none exist at plan-write time, but tests can shift).

- [ ] **Step 7: Confirm broader build is still clean**

Run:
```bash
go build ./... 2>&1 | tail -5
```
Expected: no output (clean build) or only pre-existing warnings.

- [ ] **Step 8: Commit**

```bash
git add app/domain/labels/labelapp/model.go app/domain/labels/labelapp/validate_test.go
git commit -m "fix(labels): cap Code length at 12 chars for render safety

Tightens NewLabel.Code, UpdateLabel.Code, and RenderPrintRequest.Code
validators from max=32 to max=12 — the practical Code128/BY4
alphanumeric capacity on the 812-dot 4×6 canvas after start/check/stop
overhead and quiet zones. Previously, schema- and validator-permitted
13-32 char codes would silently produce labels with clipped barcodes
at print time. Schema column stays VARCHAR(32) for future media
flexibility; the cap is enforced at API ingress.

Audit: zero seeded label codes exceed 7 chars across business/ and
app/. Tightening is safe today.

Surfaced by code review on PR #143."
```

---

## Task 3: Pin the worst-case-permitted code at the template layer (defense in depth)

**Files:**
- Modify: `business/domain/labels/labelbus/zpl/zpl_test.go` (append new snapshot test)

**Why:** Task 2 enforces the bound at the validator layer, but the ZPL package is also reachable via direct calls (e.g., from internal tooling, future call sites bypassing the API). A snapshot test that exercises the template at the maximum permitted Code length pins the byte output for the worst legal input — so a future regression that loosens the validator gets paired with an obvious template-side fail rather than a silent production print issue.

- [ ] **Step 1: Append the max-safe snapshot test**

Add to `business/domain/labels/labelbus/zpl/zpl_test.go` (after the new identity test from Task 1):

```go
// Test_Location_Snapshot_MaxSafeCode pins template byte output for a
// 12-char code — the upper bound enforced at the API validator layer
// (app/domain/labels/labelapp/model.go). If this test fails after a
// validator change, the template layout has not been re-budgeted for
// the new bound and the new bound will produce clipped barcodes.
func Test_Location_Snapshot_MaxSafeCode(t *testing.T) {
	got := zpl.Location(zpl.LocationData{Code: "STG-A01-B12C"})
	want := "^XA\n" +
		"^FO40,80^A0N,150,150^FDSTG-A01-B12C^FS\n" +
		"^FO40,300^BY4^BCN,250,Y,N,N^FDSTG-A01-B12C^FS\n" +
		"^XZ\n"
	if got != want {
		t.Fatalf("location max-safe snapshot drift.\nwant:\n%q\ngot:\n%q\n", want, got)
	}
}
```

- [ ] **Step 2: Run the new test and verify it passes**

Run:
```bash
go test -run Test_Location_Snapshot_MaxSafeCode ./business/domain/labels/labelbus/zpl/ -v
```
Expected: `--- PASS`.

- [ ] **Step 3: Run full zpl package tests**

Run:
```bash
go test ./business/domain/labels/labelbus/zpl/
```
Expected: `ok ... cached` or fresh `ok` line.

- [ ] **Step 4: Commit**

```bash
git add business/domain/labels/labelbus/zpl/zpl_test.go
git commit -m "test(labels): pin Location template at max-safe 12-char code

Snapshots the byte output of Location() for a 12-char code (the upper
bound enforced by the labelapp validator). If a future validator
loosening drifts past this bound without a paired template re-budget,
this test fails — surfacing the dimensional mismatch before broken
labels reach the printer."
```

---

## Task 4: Fix the cosmetic "middle" comment

**Files:**
- Modify: `business/domain/labels/labelbus/zpl/location.go:13`

**Why:** Reviewer noted the layout comment claims the barcode is in the "middle" of the label, but at y=300 on a 1218-dot canvas it's actually in the upper third. Pure cosmetics — no behavior change.

- [ ] **Step 1: Update the comment**

Edit `business/domain/labels/labelbus/zpl/location.go`. Find the line:
```go
//   Code128 barcode (BY4, 250-dot height, with human-readable) — middle
```
Replace with:
```go
//   Code128 barcode (BY4, 250-dot height, with human-readable) — upper third
```

- [ ] **Step 2: Verify build still clean (comment-only change)**

Run:
```bash
go build ./business/domain/labels/...
```
Expected: clean build.

- [ ] **Step 3: Commit**

```bash
git add business/domain/labels/labelbus/zpl/location.go
git commit -m "docs(labels): correct location.go layout comment

Barcode at y=300 on a 1218-dot canvas is in the upper third, not the
middle. Cosmetic; no behavior change."
```

---

## Wrap-Up

- [ ] **Wrap-Up Step 1: Run full labels-domain test suite to confirm clean state**

Run:
```bash
go test ./business/domain/labels/labelbus/... ./app/domain/labels/labelapp/...
```
Expected: all `ok`.

- [ ] **Wrap-Up Step 2: Push branch and open PR**

Run:
```bash
git push -u github fix/labels-zpl-review-followups
gh pr create --repo timmaaaz/ichor --base master --head fix/labels-zpl-review-followups \
  --title 'fix(labels): address PR #143 review follow-ups (code length cap, identity test, comment)' \
  --body "$(cat <<'EOF'
## Summary

Addresses three findings from the code review on PR #143's location/tote ZPL resize:

1. **Render-safety guard at validator layer** — `NewLabel.Code`, `UpdateLabel.Code`, `RenderPrintRequest.Code` capped from `max=32` to `max=12` (the practical Code128/BY4 alphanumeric ceiling on the 812-dot 4×6 canvas). Fails loud at API ingress; schema column stays `VARCHAR(32)` for future media flexibility. Audit confirms no seeded code exceeds 7 chars today.
2. **Pinned identity invariant** — `Test_Tote_Body_Identical_To_Location` enforces the deliberate Phase 0b "Tote body is byte-identical to Location" intent in code, not just in a comment. Will be deleted when Phase 1+ legitimately diverges them.
3. **Comment cleanup** — `location.go:13` now correctly describes the barcode position as "upper third" rather than "middle".

Defense-in-depth: `Test_Location_Snapshot_MaxSafeCode` pins the template's byte output at the 12-char upper bound, so a future validator loosening fails loud at the template layer too.

## Test plan

- [x] `go test ./business/domain/labels/labelbus/...` (zpl, labelbus, labeldb, tcpprint)
- [x] `go test ./app/domain/labels/labelapp/...`
- [x] Boundary tests for the new Code cap (12 chars passes, 28 chars rejected)
- [x] Identity invariant test verified to fail on tote.go drift (sanity check, reverted)
- [ ] Smoke print on the GK420t with a deliberately edge-case 12-char code (e.g., STG-A01-B12C) — visually inspect that text and barcode both fit within the 4×6 stock without clipping

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```
Expected: PR URL printed.

- [ ] **Wrap-Up Step 3: Squash-merge and sync Bitbucket mirror**

After CI passes (or immediately, since this repo has no CI per project memory):
```bash
gh pr merge --repo timmaaaz/ichor --squash --delete-branch <PR_NUMBER>
git checkout master
git fetch github
git merge --ff-only github/master
git push origin master
git branch -d fix/labels-zpl-review-followups
```
Expected: master fast-forwarded; Bitbucket mirror in sync.

---

## Self-Review

**Spec coverage check:**
- Reviewer's "Important: input-length overflow" → Task 2 (validator cap) + Task 3 (template snapshot) ✓
- Reviewer's "Important: byte-identical body comment-load-bearing" → Task 1 (identity invariant test) ✓
- Reviewer's "Minor: 'middle' comment nit" → Task 4 ✓
- Reviewer's "Minor: vertical real estate underused" → **deliberately not addressed.** Mirrors `product.go`'s pattern; non-regression; no functional issue.
- Reviewer's "Recommendation 2: extract helper for Tote/Location DRY" → **deliberately deferred** per reviewer's own caveat ("Defer until the Phase 1+ divergence happens — premature now").
- Reviewer's "Recommendation 3: log physical print verification" → out of plan scope; user-side action after merge.

**Placeholder scan:** No TBDs, no "fill in details", every code block is concrete and complete.

**Type consistency:** Validator tag syntax `validate:"..."` is the established `validator/v10` format used elsewhere in `app/domain/labels/labelapp/model.go`. Test function naming follows the existing `Test_Foo_Validate_Bar` convention from `validate_test.go`.

**No scope creep:** Schema unchanged. Business layer unchanged. Frontend unchanged. Only validator + tests + one comment.
