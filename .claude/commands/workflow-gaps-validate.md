# Workflow Action Gap Remediation Validation Command

Run comprehensive validation checks for the Workflow Action Gap Remediation implementation.

## Your Task

### 1. Determine Validation Scope

Check if a phase parameter was provided: `/workflow-gaps-validate [N]`

- If phase number provided: Validate only that phase
- If no parameter: Validate entire project (all completed phases)

### 2. Read PROGRESS.yaml

Read `.claude/plans/WORKFLOW_GAPS_PLAN/PROGRESS.yaml` to:
- Determine which phases to validate
- Get phase-specific validation criteria
- Check current project status

### 3. Run Standard Go Validation Checks

#### Go Build

Run: `go build ./...`

Report results:
- ✅ Pass: "Go build successful - no compilation errors"
- ❌ Fail: "Go build failed with errors" (show errors)

#### Go Vet

Run: `go vet ./...`

Report results:
- ✅ Pass: "go vet passed - no issues"
- ❌ Fail: "go vet failed with issues" (show issues)

#### Go Tests

Run: `go test ./business/sdk/workflow/... ./business/domain/workflow/...`

Report results:
- ✅ Pass: "All tests pass"
- ❌ Fail: "Tests failed" (show failures)

### 4. Run Phase-Specific Validation

For each phase being validated, check the `validation` section in PROGRESS.yaml:

For each validation check:
1. Determine how to validate it (automated test, manual check, etc.)
2. If automated, run the check
3. Report pass/fail
4. Update PROGRESS.yaml with check status

### 5. Phase-Specific Validation Details

#### Phase 1 (Missing Tables)
- `grep "purchase_orders" business/sdk/workflow/workflowactions/data/tables.go` — tables added
- `go build ./business/sdk/workflow/workflowactions/...` — compiles

#### Phase 2 (FieldChanges)
- `go test ./business/sdk/workflow/temporal/... -run TestFieldChanges` — field changes populated
- Review trigger event for `changed_from`/`changed_to` fields

#### Phase 3 (send_notification)
- `go test ./business/sdk/workflow/workflowactions/communication/... -run TestSendNotification`
- Verify output port `sent` exists

#### Phase 4 (send_email)
- `go test ./business/sdk/workflow/workflowactions/communication/... -run TestSendEmail`
- Verify SMTP interface is mockable

#### Phase 5 (seek_approval)
- `make migrate` — new table created
- `go build ./business/domain/workflow/approvalrequestbus/...`
- `go test ./api/cmd/services/ichor/tests/...`

#### Phase 6 (create_purchase_order)
- `go build ./business/sdk/workflow/workflowactions/procurement/...`
- Verify PurchaseOrder/SupplierProduct buses wired in all.go

#### Phase 7 (receive_inventory)
- `go build ./business/sdk/workflow/workflowactions/inventory/...`
- Verify InventoryTransaction bus wired in all.go

#### Phase 8 (call_webhook)
- `go build ./business/sdk/workflow/workflowactions/integration/...`
- Verify HTTPS-only enforcement

#### Phase 9 (Template Arithmetic)
- `go test ./business/sdk/workflow/... -run TestEvalExpr`
- `go test ./business/sdk/workflow/... -run TestTemplate`

### 6. Generate Validation Report

Create a comprehensive report:

```
🔍 Workflow Action Gap Remediation - Validation Report

══════════════════════════════════════════════════════════════

VALIDATION SCOPE
Phase: [PHASE_SCOPE] or "All Phases"
Status: [OVERALL_STATUS]

══════════════════════════════════════════════════════════════

STANDARD CHECKS

✅ Go Build
   Result: Passed
   Errors: 0

✅ Go Vet
   Result: Passed
   Issues: 0

✅ Go Tests
   Result: Passed
   Tests run: N
   Failures: 0

══════════════════════════════════════════════════════════════

PHASE-SPECIFIC CHECKS

Phase 1: Add Missing Tables to Whitelist
  ✅ Tables added to whitelist
  ✅ Go build passes

Phase 2: Fix FieldChanges Propagation
  ✅ FieldChanges populated in DelegateHandler
  ✅ Tests pass

...

══════════════════════════════════════════════════════════════

ISSUES FOUND

❌ Critical Issues: [N]
   - [List any critical issues]

⚠️  Warnings: [N]
   - [List any warnings]

══════════════════════════════════════════════════════════════

SUMMARY

Overall Status: ✅ All checks passed / ⚠️  Passed with warnings / ❌ Failed

══════════════════════════════════════════════════════════════

NEXT STEPS

1. Fix any critical issues identified above
2. Address warnings (optional but recommended)
3. Re-run validation: /workflow-gaps-validate
4. Once all checks pass, continue: /workflow-gaps-next
```

### 7. Update PROGRESS.yaml

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
