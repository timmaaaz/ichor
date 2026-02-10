package edgedb_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/temporal"
	"github.com/timmaaaz/ichor/business/sdk/workflow/temporal/stores/edgedb"
)

func Test_EdgeDB(t *testing.T) {
	db := dbtest.NewDatabase(t, "Test_EdgeDB")

	ctx := context.Background()

	// Seed a user (needed for TestSeedFullWorkflow's createdBy FK).
	users, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, db.BusDomain.User)
	if err != nil {
		t.Fatalf("Seeding users: %s", err)
	}

	// Seed full workflow data: 5 rules, 3 templates, 10 actions, edges.
	wfData, err := workflow.TestSeedFullWorkflow(ctx, users[0].ID, db.BusDomain.Workflow)
	if err != nil {
		t.Fatalf("Seeding workflow: %s", err)
	}

	store := edgedb.NewStore(db.Log, db.DB)

	// Pick the first rule (guaranteed to have actions seeded).
	ruleID := wfData.AutomationRules[0].ID

	// -------------------------------------------------------------------------

	t.Run("query-actions-success", func(t *testing.T) {
		actions, err := store.QueryActionsByRule(ctx, ruleID)
		if err != nil {
			t.Fatalf("QueryActionsByRule: %s", err)
		}
		if len(actions) == 0 {
			t.Fatal("expected at least one action")
		}

		for _, a := range actions {
			if a.ID == uuid.Nil {
				t.Error("action ID should not be nil")
			}
			if a.Name == "" {
				t.Error("action Name should not be empty")
			}
			// All seeded actions have templates, so ActionType should resolve.
			if a.ActionType == "" {
				t.Error("ActionType should be resolved from template")
			}
			if !json.Valid(a.Config) {
				t.Error("action Config is not valid JSON")
			}
			if !a.IsActive {
				t.Error("seeded actions should be active")
			}
			if a.DeactivatedBy != uuid.Nil {
				t.Error("seeded actions should not be deactivated")
			}
		}
	})

	// -------------------------------------------------------------------------

	t.Run("query-actions-not-found", func(t *testing.T) {
		actions, err := store.QueryActionsByRule(ctx, uuid.New())
		if err != nil {
			t.Fatalf("QueryActionsByRule: expected nil error, got %s", err)
		}
		if len(actions) != 0 {
			t.Fatalf("expected empty slice, got %d actions", len(actions))
		}
	})

	// -------------------------------------------------------------------------

	t.Run("query-edges-success", func(t *testing.T) {
		edges, err := store.QueryEdgesByRule(ctx, ruleID)
		if err != nil {
			t.Fatalf("QueryEdgesByRule: %s", err)
		}
		if len(edges) == 0 {
			t.Fatal("expected at least one edge")
		}

		// Verify at least one start edge with nil SourceActionID.
		hasStart := false
		for _, e := range edges {
			if e.EdgeType == temporal.EdgeTypeStart {
				hasStart = true
				if e.SourceActionID != nil {
					t.Error("start edge should have nil SourceActionID")
				}
			}
			// All non-start edges should have a SourceActionID.
			if e.EdgeType != temporal.EdgeTypeStart && e.SourceActionID == nil {
				t.Errorf("non-start edge %s should have SourceActionID", e.ID)
			}
		}
		if !hasStart {
			t.Error("expected at least one start edge")
		}

		// Verify ordering: SortOrder should be non-decreasing.
		for i := 1; i < len(edges); i++ {
			if edges[i].SortOrder < edges[i-1].SortOrder {
				t.Errorf("edges not ordered: [%d].SortOrder=%d < [%d].SortOrder=%d",
					i, edges[i].SortOrder, i-1, edges[i-1].SortOrder)
			}
		}
	})

	// -------------------------------------------------------------------------

	t.Run("query-edges-not-found", func(t *testing.T) {
		edges, err := store.QueryEdgesByRule(ctx, uuid.New())
		if err != nil {
			t.Fatalf("QueryEdgesByRule: expected nil error, got %s", err)
		}
		if len(edges) != 0 {
			t.Fatalf("expected empty slice, got %d edges", len(edges))
		}
	})

	// -------------------------------------------------------------------------

	t.Run("round-trip", func(t *testing.T) {
		actions, err := store.QueryActionsByRule(ctx, ruleID)
		if err != nil {
			t.Fatalf("QueryActionsByRule: %s", err)
		}
		edges, err := store.QueryEdgesByRule(ctx, ruleID)
		if err != nil {
			t.Fatalf("QueryEdgesByRule: %s", err)
		}

		// Build action ID set.
		actionIDs := make(map[uuid.UUID]bool, len(actions))
		for _, a := range actions {
			actionIDs[a.ID] = true
		}

		// Verify all edge source/target IDs reference valid actions.
		for _, e := range edges {
			if !actionIDs[e.TargetActionID] {
				t.Errorf("edge %s targets unknown action %s", e.ID, e.TargetActionID)
			}
			if e.SourceActionID != nil && !actionIDs[*e.SourceActionID] {
				t.Errorf("edge %s sources unknown action %s", e.ID, *e.SourceActionID)
			}
		}

		// Verify graph can be used to build a GraphDefinition.
		gd := temporal.GraphDefinition{
			Actions: actions,
			Edges:   edges,
		}
		if len(gd.Actions) == 0 || len(gd.Edges) == 0 {
			t.Error("GraphDefinition should have actions and edges")
		}
	})
}
