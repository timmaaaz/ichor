# Workflow Temporal Implementation Validation Command

Run comprehensive validation checks for the Workflow Temporal Implementation.

## Your Task

### 1. Determine Validation Scope

Check if a phase parameter was provided: `/workflow-temporal-validate [N]`

- If phase number provided: Validate only that phase
- If no parameter: Validate entire project (all completed phases)

### 2. Read PROGRESS.yaml

Read `.claude/plans/WORKFLOW_TEMPORAL_PLAN/PROGRESS.yaml` to:
- Determine which phases to validate
- Get phase-specific validation criteria
- Check current project status

### 3. Run Standard Validation Checks

#### Go Compilation

Run: `go build ./...`

Report results:
- ✅ Pass: "Go compilation successful - no errors"
- ❌ Fail: "Go compilation failed with errors" (show errors)

#### Go Tests

Run: `make test`

Report results:
- ✅ Pass: "All tests passed"
- ❌ Fail: "Tests failed" (show failures)

#### Linting

Run: `make lint`

Report results:
- ✅ Pass: "Linting passed - no errors"
- ❌ Fail: "Linting failed" (show errors)

### 4. Run Phase-Specific Validation

For each phase being validated, check the `validation` section in PROGRESS.yaml:

```yaml
validation:
  - check: "make dev-bounce completes successfully"
    status: "pending"
  - check: "Temporal UI accessible at configured port"
    status: "pending"
  - check: "workflow-worker pod reaches Running state"
    status: "pending"
```

For each validation check:
1. Determine how to validate it (automated test, manual check, etc.)
2. If automated, run the check
3. Report pass/fail
4. Update PROGRESS.yaml with check status

### 5. Optional Validation (Ask User First)

Ask if user wants to run:

- **Full Dev Bounce**: `make dev-bounce`
  - May take time
  - Verifies complete local deployment

- **Determinism Tests**: Run graph executor with same input multiple times
  - Verify identical output

If user says yes, run these and report results.

### 6. Check for Common Issues

#### File Existence

Verify all files listed in PROGRESS.yaml deliverables exist:
- ✅ Pass: "All deliverable files exist"
- ❌ Fail: "Missing files: [list]"

#### Temporal-Specific Checks

- No `time.Now()` in workflow code
- No `rand` functions in workflow code
- All map iterations sorted
- No direct HTTP/DB calls in workflow code

### 7. Generate Validation Report

Create a comprehensive report:

```
🔍 Workflow Temporal Implementation - Validation Report

══════════════════════════════════════════════════════════════

VALIDATION SCOPE
Phase: {{PHASE_SCOPE}} or "All Phases"
Status: {{OVERALL_STATUS}}

══════════════════════════════════════════════════════════════

STANDARD CHECKS

✅ Go Compilation
   Result: Passed
   Packages built: 42

✅ Go Tests
   Result: Passed
   Tests run: 156
   Coverage: 78%

✅ Linting
   Result: Passed
   Files checked: 98

══════════════════════════════════════════════════════════════

PHASE-SPECIFIC CHECKS

Phase 1: Infrastructure Setup
  ✅ Temporal K8s manifests exist
  ✅ workflow-worker Dockerfile exists
  ✅ Makefile targets added

Phase 4: Graph Executor
  ✅ All map iterations are sorted
  ✅ Determinism test passed (1000 iterations)
  ❌ Missing unit test for edge case

══════════════════════════════════════════════════════════════

TEMPORAL DETERMINISM CHECKS

✅ No time.Now() in workflow code
✅ No rand functions in workflow code
✅ All map iterations sorted
✅ No direct HTTP/DB calls in workflow code

══════════════════════════════════════════════════════════════

OPTIONAL CHECKS

⏩ Full Dev Bounce: Skipped (not requested)
⏩ Determinism Stress Test: Skipped (not requested)

══════════════════════════════════════════════════════════════

FILE VERIFICATION

✅ All deliverable files exist
   Created: 45 files
   Modified: 12 files
   Missing: 0 files

══════════════════════════════════════════════════════════════

ISSUES FOUND

❌ Critical Issues: 1
   - Missing unit test for edge case in Phase 4

⚠️  Warnings: 0

══════════════════════════════════════════════════════════════

SUMMARY

Overall Status: ⚠️  Passed with issues
Critical Issues: 1
Warnings: 0

Recommendation: Add missing unit test before proceeding

══════════════════════════════════════════════════════════════

NEXT STEPS

1. Fix critical issues identified above
2. Re-run validation: /workflow-temporal-validate
3. Once all checks pass, continue: /workflow-temporal-next
```

### 8. Update PROGRESS.yaml

Update validation status in PROGRESS.yaml:
- Mark completed validation checks as `completed`
- If validation fails, consider marking phase as `blocked`
- Add failures to `blockers` section

## Validation Categories

### Infrastructure Validation

- K8s manifests valid (kustomize build)
- Dockerfile builds successfully
- Makefile targets work

### Backend Validation

- Go compilation: `go build ./...`
- Go tests: `make test`
- Linting: `make lint`

### Temporal-Specific Validation

- Determinism requirements met
- No forbidden operations in workflow code
- Proper activity separation

### Integration Validation

- Worker connects to Temporal
- Workflows can be started
- Activities execute correctly

## Tips

- Be thorough but not overwhelming
- Prioritize critical issues over warnings
- Provide actionable next steps
- Update PROGRESS.yaml with results
- If everything passes, congratulate the user!
