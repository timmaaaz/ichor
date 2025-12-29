# Default Status Management Validation Command

Run comprehensive validation checks for the Default Status Management implementation.

## Your Task

### 1. Determine Validation Scope

Check if a phase parameter was provided: `/default-statuses-validate [N]`

- If phase number provided: Validate only that phase
- If no parameter: Validate entire project (all completed phases)

### 2. Read PROGRESS.yaml

Read `.claude/plans/DEFAULT_STATUSES_PLAN/PROGRESS.yaml` to:
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

Run: `make test` or `go test ./...`

Report results:
- ✅ Pass: "All tests passed"
- ❌ Fail: "Tests failed" (show failures)

#### Linting

Run: `make lint`

Report results:
- ✅ Pass: "Linting passed - no errors"
- ❌ Fail: "Linting failed with N errors" (show errors)

### 4. Run Phase-Specific Validation

For each phase being validated, check the `validation` section in PROGRESS.yaml.

**Phase 1 Validations:**
- Form config with `default_value: "Pending"` resolves to correct UUID
- Orders created via formdata have correct fulfillment_status_id
- Line items have correct line_item_fulfillment_statuses_id
- Invalid status names produce clear validation errors

**Phase 2 Validations:**
- Order creation triggers allocation workflow
- Allocation success updates line items to ALLOCATED
- Allocation failure keeps PENDING and creates alert

**Phase 3 Validations:**
- Alerts created with role-based recipients
- Users can view and acknowledge alerts

### 5. Generate Validation Report

Create a comprehensive report:

```
🔍 Default Status Management - Validation Report

══════════════════════════════════════════════════════════════

VALIDATION SCOPE
Phase: {{PHASE_SCOPE}} or "All Phases"
Status: {{OVERALL_STATUS}}

══════════════════════════════════════════════════════════════

STANDARD CHECKS

✅ Go Compilation
   Result: Passed

✅ Go Tests
   Result: Passed
   Tests: 142 passed, 0 failed

✅ Linting
   Result: Passed

══════════════════════════════════════════════════════════════

PHASE-SPECIFIC CHECKS

Phase 1: Form Configuration FK Default Resolution
  ✅ FK default resolution implemented
  ✅ Form seeds updated with default values
  ⏳ Integration test pending

══════════════════════════════════════════════════════════════

FILE VERIFICATION

✅ All deliverable files exist
   Modified: 4 files

══════════════════════════════════════════════════════════════

ISSUES FOUND

❌ Critical Issues: 0

⚠️  Warnings: 0

══════════════════════════════════════════════════════════════

SUMMARY

Overall Status: ✅ Passed

══════════════════════════════════════════════════════════════

NEXT STEPS

1. Continue with /default-statuses-next
```

### 6. Update PROGRESS.yaml

Update validation status in PROGRESS.yaml:
- Mark completed validation checks as `completed`
- If validation fails, consider marking phase as `blocked`
- Add failures to `blockers` section

## Tips

- Be thorough but not overwhelming
- Prioritize critical issues over warnings
- Provide actionable next steps
- Update PROGRESS.yaml with results
- If everything passes, congratulate the user!
