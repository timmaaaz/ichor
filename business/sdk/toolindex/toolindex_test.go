package toolindex

import (
	"context"
	"encoding/binary"
	"io"
	"math"
	"testing"

	"github.com/timmaaaz/ichor/business/sdk/llm"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// =========================================================================
// Mock embedder: deterministic hash-based, no external dependencies
// =========================================================================

// mockEmbedder produces deterministic 8-dimensional embeddings by hashing
// the input text via FNV-like mixing. This gives stable, non-random vectors
// that are different for different inputs.
type mockEmbedder struct {
	dim int
}

func newMockEmbedder() *mockEmbedder {
	return &mockEmbedder{dim: 8}
}

func (m *mockEmbedder) Embed(_ context.Context, text string) ([]float32, error) {
	return m.hashEmbed(text), nil
}

func (m *mockEmbedder) hashEmbed(text string) []float32 {
	// FNV-1a-style hash to fill each dimension.
	vec := make([]float32, m.dim)
	for i := range vec {
		h := uint32(2166136261) ^ uint32(i)
		for _, b := range []byte(text) {
			h ^= uint32(b)
			h *= 16777619
		}
		// Map to [-1, 1].
		buf := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, h)
		bits := binary.LittleEndian.Uint32(buf)
		vec[i] = float32(bits)/float32(math.MaxUint32)*2 - 1
	}
	return normalise(vec)
}

// failEmbedder always returns nil embeddings (simulates embedding failure).
type failEmbedder struct{}

func (f *failEmbedder) Embed(_ context.Context, _ string) ([]float32, error) {
	return nil, nil
}

// =========================================================================
// Helpers
// =========================================================================

func testLogger() *logger.Logger {
	return logger.New(io.Discard, logger.LevelError, "test", nil)
}

func makeTool(name, desc string) llm.ToolDef {
	return llm.ToolDef{
		Name:        name,
		Description: desc,
	}
}

func buildIndex(t *testing.T, embedder Embedder, tools []llm.ToolDef) *ToolIndex {
	t.Helper()
	idx, err := New(context.Background(), Config{
		Embedder: embedder,
		Log:      testLogger(),
	}, tools)
	if err != nil {
		t.Fatalf("unexpected error building index: %v", err)
	}
	return idx
}

// =========================================================================
// Tests
// =========================================================================

func TestSearch_Ranking(t *testing.T) {
	tools := []llm.ToolDef{
		makeTool("list_workflow_rules", "List all workflow automation rules"),
		makeTool("create_draft", "Create a new workflow draft"),
		makeTool("get_table_config", "Get table configuration settings"),
	}

	idx := buildIndex(t, newMockEmbedder(), tools)

	matches, _, err := idx.Search(context.Background(), "workflow rules", 10, SearchOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(matches) != 3 {
		t.Fatalf("expected 3 matches, got %d", len(matches))
	}

	// Verify descending score order.
	for i := 1; i < len(matches); i++ {
		if matches[i].Score > matches[i-1].Score {
			t.Errorf("results not sorted: score[%d]=%f > score[%d]=%f",
				i, matches[i].Score, i-1, matches[i-1].Score)
		}
	}
}

func TestSearch_Allowlist(t *testing.T) {
	tools := []llm.ToolDef{
		makeTool("tool_a", "Alpha tool for workflows"),
		makeTool("tool_b", "Beta tool for tables"),
		makeTool("tool_c", "Charlie tool for configs"),
	}

	idx := buildIndex(t, newMockEmbedder(), tools)

	allowlist := map[string]bool{
		"tool_a": true,
		"tool_c": true,
	}

	matches, _, err := idx.Search(context.Background(), "anything", 10, SearchOptions{
		Allowlist: allowlist,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(matches) != 2 {
		t.Fatalf("expected 2 matches (allowlisted), got %d", len(matches))
	}

	for _, m := range matches {
		if !allowlist[m.Tool.Name] {
			t.Errorf("tool %q not in allowlist but was returned", m.Tool.Name)
		}
	}
}

func TestSearch_MinScore(t *testing.T) {
	tools := []llm.ToolDef{
		makeTool("workflow_tool", "Manage workflow automation rules and triggers"),
		makeTool("unrelated_tool", "Something completely different about cooking recipes"),
	}

	idx := buildIndex(t, newMockEmbedder(), tools)

	// Use a high threshold that should filter out at least some results.
	matches, _, err := idx.Search(context.Background(), "workflow automation", 10, SearchOptions{
		MinScore: 0.99,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// With a very high threshold, we expect fewer results.
	for _, m := range matches {
		if m.Score < 0.99 {
			t.Errorf("tool %q has score %f below min_score 0.99", m.Tool.Name, m.Score)
		}
	}

	// Verify a zero threshold returns all.
	all, _, err := idx.Search(context.Background(), "workflow automation", 10, SearchOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 matches with no threshold, got %d", len(all))
	}
}

func TestSearch_EmptyIndex(t *testing.T) {
	idx := buildIndex(t, newMockEmbedder(), nil)

	matches, _, err := idx.Search(context.Background(), "anything", 5, SearchOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 0 {
		t.Fatalf("expected 0 matches from empty index, got %d", len(matches))
	}
}

func TestEmbeddedCount(t *testing.T) {
	tools := []llm.ToolDef{
		makeTool("a", "tool a"),
		makeTool("b", "tool b"),
		makeTool("c", "tool c"),
	}

	// All embeddings succeed.
	idx := buildIndex(t, newMockEmbedder(), tools)
	if got := idx.EmbeddedCount(); got != 3 {
		t.Errorf("expected EmbeddedCount=3, got %d", got)
	}

	// No tools.
	empty := buildIndex(t, newMockEmbedder(), nil)
	if got := empty.EmbeddedCount(); got != 0 {
		t.Errorf("expected EmbeddedCount=0 for empty index, got %d", got)
	}
}

func TestSearch_NilEmbeddingsSkipped(t *testing.T) {
	tools := []llm.ToolDef{
		makeTool("a", "tool a"),
		makeTool("b", "tool b"),
	}

	// Build index with a fail embedder — all embeddings will be nil.
	idx := buildIndex(t, &failEmbedder{}, tools)

	if got := idx.EmbeddedCount(); got != 0 {
		t.Errorf("expected EmbeddedCount=0 with failed embedder, got %d", got)
	}

	// Search should still work, just return no matches.
	matches, _, err := idx.Search(context.Background(), "anything", 10, SearchOptions{})
	if err != nil {
		// failEmbedder.Embed returns (nil, nil) — the query embed will also
		// return nil with no error. The dot product loop simply skips nil
		// embeddings, so no matches.
		// Actually, embed returns nil vec — dot with nil is fine (loop 0 iters).
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 0 {
		t.Fatalf("expected 0 matches when embeddings are nil, got %d", len(matches))
	}
}
