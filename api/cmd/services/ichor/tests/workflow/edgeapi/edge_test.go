package edge_test

import (
	"testing"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

func Test_EdgeAPI(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_EdgeAPI")

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// ============================================================
	// Query Edge Tests
	// ============================================================
	test.Run(t, queryEdges200(sd), "queryEdges-200")
	test.Run(t, queryEdgesRuleNotFound404(sd), "queryEdges-rule-not-found-404")
	test.Run(t, queryEdges401(sd), "queryEdges-401")
	test.Run(t, queryEdgeByID200(sd), "queryEdgeByID-200")
	test.Run(t, queryEdgeByIDNotFound404(sd), "queryEdgeByID-not-found-404")
	test.Run(t, queryEdgeByIDWrongRule404(sd), "queryEdgeByID-wrong-rule-404")

	// ============================================================
	// Create Edge Tests
	// ============================================================
	test.Run(t, createEdgeStart200(sd), "createEdge-start-200")
	test.Run(t, createEdgeSequence200(sd), "createEdge-sequence-200")
	test.Run(t, createEdgeBranch200(sd), "createEdge-branch-200")
	test.Run(t, createEdgeInvalidType400(sd), "createEdge-invalid-type-400")
	test.Run(t, createEdgeMissingTarget400(sd), "createEdge-missing-target-400")
	test.Run(t, createEdgeStartWithSource400(sd), "createEdge-start-with-source-400")
	test.Run(t, createEdgeNonStartWithoutSource400(sd), "createEdge-non-start-without-source-400")
	test.Run(t, createEdgeTargetNotFound404(sd), "createEdge-target-not-found-404")
	test.Run(t, createEdgeSourceNotFound404(sd), "createEdge-source-not-found-404")
	test.Run(t, createEdgeRuleNotFound404(sd), "createEdge-rule-not-found-404")
	test.Run(t, createEdgeTargetNotInRule400(sd), "createEdge-target-not-in-rule-400")
	test.Run(t, createEdgeSourceNotInRule400(sd), "createEdge-source-not-in-rule-400")
	test.Run(t, createEdge401(sd), "createEdge-401")

	// ============================================================
	// Delete Edge Tests
	// ============================================================
	test.Run(t, deleteEdge200(sd), "deleteEdge-200")
	test.Run(t, deleteEdgeNotFound404(sd), "deleteEdge-not-found-404")
	test.Run(t, deleteEdgeWrongRule404(sd), "deleteEdge-wrong-rule-404")
	test.Run(t, deleteEdgeRuleNotFound404(sd), "deleteEdge-rule-not-found-404")
	test.Run(t, deleteEdge401(sd), "deleteEdge-401")
	test.Run(t, deleteAllEdges200(sd), "deleteAllEdges-200")
	test.Run(t, deleteAllEdgesRuleNotFound404(sd), "deleteAllEdges-rule-not-found-404")
}
