package rule_test

import (
	"testing"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

func Test_RuleAPI(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_RuleAPI")

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// Query rules tests
	test.Run(t, queryRules200(sd), "queryRules-200")
	test.Run(t, queryRules401(sd), "queryRules-401")

	// Query rule by ID tests
	test.Run(t, queryRuleByID200(sd), "queryRuleByID-200")
	test.Run(t, queryRuleByID404(sd), "queryRuleByID-404")
	test.Run(t, queryRuleByID401(sd), "queryRuleByID-401")

	// Create rule tests
	test.Run(t, createRule201(sd), "createRule-201")
	test.Run(t, createRule400(sd), "createRule-400")
	test.Run(t, createRule401(sd), "createRule-401")

	// Update rule tests
	test.Run(t, updateRule200(sd), "updateRule-200")
	test.Run(t, updateRule404(sd), "updateRule-404")
	test.Run(t, updateRule401(sd), "updateRule-401")

	// Delete rule tests
	test.Run(t, deleteRule200(sd), "deleteRule-200")
	test.Run(t, deleteRule404(sd), "deleteRule-404")
	test.Run(t, deleteRule401(sd), "deleteRule-401")

	// Toggle active tests
	test.Run(t, toggleActive200(sd), "toggleActive-200")
	test.Run(t, toggleActive404(sd), "toggleActive-404")
	test.Run(t, toggleActive401(sd), "toggleActive-401")

	// ============================================================
	// Phase 4C: Action CRUD Tests
	// ============================================================

	// Query actions tests
	test.Run(t, queryActions200(sd), "queryActions-200")
	test.Run(t, queryActions404(sd), "queryActions-404")
	test.Run(t, queryActions401(sd), "queryActions-401")

	// Create action tests
	test.Run(t, createAction201(sd), "createAction-201")
	test.Run(t, createAction400(sd), "createAction-400")
	test.Run(t, createAction404(sd), "createAction-404")
	test.Run(t, createAction401(sd), "createAction-401")

	// Update action tests
	test.Run(t, updateAction200(sd), "updateAction-200")
	test.Run(t, updateAction404(sd), "updateAction-404")
	test.Run(t, updateAction401(sd), "updateAction-401")

	// Delete action tests
	test.Run(t, deleteAction200(sd), "deleteAction-200")
	test.Run(t, deleteAction404(sd), "deleteAction-404")
	test.Run(t, deleteAction401(sd), "deleteAction-401")

	// ============================================================
	// Phase 4C: Validation Tests
	// ============================================================

	// Validate rule tests
	test.Run(t, validateRule200(sd), "validateRule-200")
	test.Run(t, validateRule200WithWarnings(sd), "validateRule-200-warnings")
	test.Run(t, validateRule404(sd), "validateRule-404")
	test.Run(t, validateRule401(sd), "validateRule-401")

	// ============================================================
	// Phase 5: Simulation & Execution History Tests
	// ============================================================

	// Test/simulate rule tests
	test.Run(t, testRule200(sd), "testRule-200")
	test.Run(t, testRule404(sd), "testRule-404")
	test.Run(t, testRule401(sd), "testRule-401")

	// Rule execution history tests
	test.Run(t, queryRuleExecutions200(sd), "queryRuleExecutions-200")
	test.Run(t, queryRuleExecutions404(sd), "queryRuleExecutions-404")
	test.Run(t, queryRuleExecutions401(sd), "queryRuleExecutions-401")
}
