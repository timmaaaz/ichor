# Universal Action Edge Enforcement Validation Command

Run comprehensive validation checks for the Universal Action Edge Enforcement implementation.

## Your Task

### 1. Determine Validation Scope

Check if a phase parameter was provided: `/missing-action-edges-validate [N]`

- If phase number provided: Validate only that phase
- If no parameter: Validate entire project (all completed phases)

### 2. Read PROGRESS.yaml

Read `.claude/plans/MISSING_ACTION_EDGES_PLAN/PROGRESS.yaml` to:
- Determine which phases to validate
- Get phase-specific validation criteria
- Check current project status

### 3. Run Standard Validation Checks

#### Go Compilation

Run: `go build ./...`

Report results:
- ✅ Pass: "Go compilation successful - no errors"
- ❌ Fail: "Go compilation failed with N errors" (show errors)

#### Go Tests

Run: `make test`

Report results:
- ✅ Pass: "All tests passed"
- ❌ Fail: "Tests failed" (show failures)

#### Linting

Run: `make lint`

Report results:
- ✅ Pass: "Lint passed - no errors"
- ❌ Fail: "Lint failed with N errors" (show errors)

### 4. Run Phase-Specific Validation

For each phase being validated, check the `validation` section in PROGRESS.yaml:

```yaml
validation:
  - check: "Go compilation passes"
    status: "pending"
  - check: "Existing tests still pass"
    status: "pending"
  - check: "New validation rejects actions without edges"
    status: "pending"
```

For each validation check:
1. Determine how to validate it (automated test, manual check, etc.)
2. If automated, run the check
3. Report pass/fail
4. Update PROGRESS.yaml with check status

### 5. Optional Validation (Ask User First)

Ask if user wants to run:

- **Full Test Suite**: `make test`
  - May take time
  - Shows test coverage

- **Database Migration Test**: Test migration applies
  - Ensures schema changes work
  - May need test database

If user says yes, run these and report results.

### 6. Check for Common Issues

#### File Existence

Verify all files listed in PROGRESS.yaml deliverables exist:
- ✅ Pass: "All deliverable files exist"
- ❌ Fail: "Missing files: [list]"

#### Import Errors

Check for common import issues:
- Missing imports
- Circular dependencies
- Unused imports (if lint didn't catch)

### 7. Generate Validation Report

Create a comprehensive report:

```
🔍 Universal Action Edge Enforcement - Validation Report

══════════════════════════════════════════════════════════════

VALIDATION SCOPE
Phase: {{PHASE_SCOPE}} or "All Phases"
Status: {{OVERALL_STATUS}}

══════════════════════════════════════════════════════════════

STANDARD CHECKS

✅ Go Compilation
   Result: Passed
   Packages checked: 142
   Errors: 0

✅ Go Tests
   Result: Passed
   Tests run: 450
   Passed: 450
   Failed: 0

✅ Lint
   Result: Passed
   Files checked: 98
   Errors: 0
   Warnings: 2

══════════════════════════════════════════════════════════════

PHASE-SPECIFIC CHECKS

Phase 1: Validation Layer Changes
  ✅ Go compilation passes
  ✅ Existing tests still pass
  ✅ New validation rejects actions without edges

Phase 2: Remove execution_order Field
  ✅ Migration applies successfully
  ✅ No references to execution_order remain
  ❌ Some tests still reference execution_order

══════════════════════════════════════════════════════════════

OPTIONAL CHECKS

⏩ Full Test Suite: Skipped (not requested)
⏩ Database Migration Test: Skipped (not requested)

══════════════════════════════════════════════════════════════

FILE VERIFICATION

✅ All deliverable files exist
   Created: 15 files
   Modified: 12 files
   Missing: 0 files

══════════════════════════════════════════════════════════════

ISSUES FOUND

❌ Critical Issues: 1
   - Some tests still reference execution_order in Phase 2

⚠️  Warnings: 2
   - Lint warning: Unused variable in executor.go
   - Lint warning: Line too long in workflowdb.go

══════════════════════════════════════════════════════════════

SUMMARY

Overall Status: ⚠️  Passed with warnings
Critical Issues: 1
Warnings: 2

Recommendation: Fix test references to execution_order before proceeding

══════════════════════════════════════════════════════════════

NEXT STEPS

1. Fix critical issues identified above
2. Address warnings (optional but recommended)
3. Re-run validation: /missing-action-edges-validate
4. Once all checks pass, continue: /missing-action-edges-next
```

### 8. Update PROGRESS.yaml

Update validation status in PROGRESS.yaml:
- Mark completed validation checks as `completed`
- If validation fails, consider marking phase as `blocked`
- Add failures to `blockers` section

## Validation Categories

### Backend Validation

- Go compilation: `go build ./...`
- Go tests: `make test`
- API endpoint testing
- Database migration validation
- Schema verification

### Testing Validation

- Test file structure
- Test coverage goals
- All tests pass
- No flaky tests

### Documentation Validation

- Documentation is accurate
- No references to removed features
- Examples are correct

## Tips

- Be thorough but not overwhelming
- Prioritize critical issues over warnings
- Provide actionable next steps
- Update PROGRESS.yaml with results
- If everything passes, congratulate the user!
