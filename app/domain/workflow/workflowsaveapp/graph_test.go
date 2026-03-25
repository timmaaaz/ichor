package workflowsaveapp

import (
	"strings"
	"testing"
)

// helper to create a SaveActionRequest with the given name and type.
func action(name string) SaveActionRequest {
	return SaveActionRequest{
		Name:         name,
		ActionType:   "send_email",
		ActionConfig: []byte(`{"recipients":["a@b.com"],"subject":"s","body":"b"}`),
		IsActive:     true,
	}
}

// helper to create an action with an existing ID.
func actionWithID(name, id string) SaveActionRequest {
	a := action(name)
	a.ID = &id
	return a
}

func edge(edgeType, source, target string) SaveEdgeRequest {
	return SaveEdgeRequest{
		EdgeType:       edgeType,
		SourceActionID: source,
		TargetActionID: target,
	}
}

func startEdge(target string) SaveEdgeRequest {
	return edge("start", "", target)
}

func seqEdge(source, target string) SaveEdgeRequest {
	return edge("sequence", source, target)
}

func TestValidateGraph_EmptyActions(t *testing.T) {
	err := ValidateGraph(nil, nil)
	if err != nil {
		t.Fatalf("empty graph should be valid, got: %s", err)
	}
}

func TestValidateGraph_ActionsWithoutEdges(t *testing.T) {
	actions := []SaveActionRequest{action("a")}
	err := ValidateGraph(actions, nil)
	if err == nil {
		t.Fatal("expected error for actions without edges")
	}
	if !strings.Contains(err.Error(), "at least one edge") {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestValidateGraph_NoStartEdge(t *testing.T) {
	actions := []SaveActionRequest{action("a"), action("b")}
	edges := []SaveEdgeRequest{seqEdge("temp:0", "temp:1")}
	err := ValidateGraph(actions, edges)
	if err == nil {
		t.Fatal("expected error for missing start edge")
	}
	if !strings.Contains(err.Error(), "start edge is required") {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestValidateGraph_MultipleStartEdges(t *testing.T) {
	actions := []SaveActionRequest{action("a"), action("b")}
	edges := []SaveEdgeRequest{
		startEdge("temp:0"),
		startEdge("temp:1"),
	}
	err := ValidateGraph(actions, edges)
	if err == nil {
		t.Fatal("expected error for multiple start edges")
	}
	if !strings.Contains(err.Error(), "only one start edge") {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestValidateGraph_LinearChain(t *testing.T) {
	actions := []SaveActionRequest{action("a"), action("b"), action("c")}
	edges := []SaveEdgeRequest{
		startEdge("temp:0"),
		seqEdge("temp:0", "temp:1"),
		seqEdge("temp:1", "temp:2"),
	}
	if err := ValidateGraph(actions, edges); err != nil {
		t.Fatalf("linear chain should be valid: %s", err)
	}
}

func TestValidateGraph_Diamond(t *testing.T) {
	actions := []SaveActionRequest{action("a"), action("b"), action("c"), action("d")}
	edges := []SaveEdgeRequest{
		startEdge("temp:0"),
		seqEdge("temp:0", "temp:1"),
		seqEdge("temp:0", "temp:2"),
		seqEdge("temp:1", "temp:3"),
		seqEdge("temp:2", "temp:3"),
	}
	if err := ValidateGraph(actions, edges); err != nil {
		t.Fatalf("diamond should be valid: %s", err)
	}
}

func TestValidateGraph_CycleDetection(t *testing.T) {
	tests := []struct {
		name    string
		actions []SaveActionRequest
		edges   []SaveEdgeRequest
	}{
		{
			name:    "simple cycle A->B->C->A",
			actions: []SaveActionRequest{action("a"), action("b"), action("c")},
			edges: []SaveEdgeRequest{
				startEdge("temp:0"),
				seqEdge("temp:0", "temp:1"),
				seqEdge("temp:1", "temp:2"),
				seqEdge("temp:2", "temp:0"),
			},
		},
		{
			name:    "self-loop",
			actions: []SaveActionRequest{action("a")},
			edges: []SaveEdgeRequest{
				startEdge("temp:0"),
				seqEdge("temp:0", "temp:0"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGraph(tt.actions, tt.edges)
			if err == nil {
				t.Fatal("expected cycle detection error")
			}
			if !strings.Contains(err.Error(), "cycle") {
				t.Fatalf("expected cycle error, got: %s", err)
			}
		})
	}
}

func TestValidateGraph_UnreachableAction(t *testing.T) {
	actions := []SaveActionRequest{action("a"), action("b"), action("orphan")}
	edges := []SaveEdgeRequest{
		startEdge("temp:0"),
		seqEdge("temp:0", "temp:1"),
		// temp:2 ("orphan") has no incoming edge and is not the start target
	}
	err := ValidateGraph(actions, edges)
	if err == nil {
		t.Fatal("expected unreachable action error")
	}
	if !strings.Contains(err.Error(), "not reachable") {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestValidateGraph_InvalidActionRef(t *testing.T) {
	actions := []SaveActionRequest{action("a")}
	edges := []SaveEdgeRequest{
		startEdge("temp:0"),
		seqEdge("temp:0", "temp:99"),
	}
	err := ValidateGraph(actions, edges)
	if err == nil {
		t.Fatal("expected error for invalid action reference")
	}
}

func TestValidateGraph_UUIDReferences(t *testing.T) {
	existingID := "550e8400-e29b-41d4-a716-446655440000"
	actions := []SaveActionRequest{
		actionWithID("existing", existingID),
		action("new"),
	}
	edges := []SaveEdgeRequest{
		startEdge(existingID),
		seqEdge(existingID, "temp:1"),
	}
	if err := ValidateGraph(actions, edges); err != nil {
		t.Fatalf("UUID references should work: %s", err)
	}
}

func TestResolveActionRef(t *testing.T) {
	refMap := map[string]string{
		"temp:0":    "temp:0",
		"temp:1":    "temp:1",
		"some-uuid": "temp:0",
	}

	tests := []struct {
		name    string
		ref     string
		want    string
		wantErr bool
	}{
		{"temp ref", "temp:0", "temp:0", false},
		{"uuid ref", "some-uuid", "temp:0", false},
		{"empty ref", "", "", true},
		{"unknown ref", "unknown-uuid", "", true},
		{"temp out of range", "temp:99", "", true},
		{"invalid temp format", "temp:abc", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveActionRef(tt.ref, refMap, 2)
			if tt.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if !tt.wantErr && got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
