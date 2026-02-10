package workflowsaveapi_test

import (
	"testing"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

func Test_WorkflowSaveAPI(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_WorkflowSaveAPI")

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// ============================================================
	// Phase 6: Create Workflow Tests (POST /v1/workflow/rules/full)
	// ============================================================

	test.Run(t, create200Basic(sd), "create-200-basic")
	test.Run(t, create200WithSequence(sd), "create-200-with-sequence")
	test.Run(t, create200WithBranch(sd), "create-200-with-branch")
	test.Run(t, create200WithCanvasLayout(sd), "create-200-with-canvas-layout")
	test.Run(t, create200TempIDResolution(sd), "create-200-temp-id-resolution")
	test.Run(t, create400(sd), "create-400")
	test.Run(t, create401(sd), "create-401")

	// ============================================================
	// Phase 6: Update Workflow Tests (PUT /v1/workflow/rules/{id}/full)
	// ============================================================

	test.Run(t, update200RuleOnly(sd), "update-200-rule-only")
	test.Run(t, update200AddAction(sd), "update-200-add-action")
	test.Run(t, update200UpdateAction(sd), "update-200-update-action")
	test.Run(t, update200DeleteAction(sd), "update-200-delete-action")
	test.Run(t, update200ReplaceEdges(sd), "update-200-replace-edges")
	test.Run(t, update200CanvasLayout(sd), "update-200-canvas-layout")
	test.Run(t, update400(sd), "update-400")
	test.Run(t, update401(sd), "update-401")
	test.Run(t, update404(sd), "update-404")

	// ============================================================
	// Phase 6: Validation Tests
	// ============================================================

	test.Run(t, validationActionConfig(sd), "validation-action-config")
	test.Run(t, validationGraph(sd), "validation-graph")
	test.Run(t, validationEdgeRequirement(sd), "validation-edge-requirement")

	// Phase 13: Execution, trigger, action, and error tests excluded.
	// These tests depend on the old workflow.Engine which was removed.
	// Phase 15 will rewrite them for Temporal.
}
