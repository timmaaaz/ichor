package toolindex

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/timmaaaz/ichor/business/sdk/llm"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// ToolMatch pairs a tool definition with its cosine similarity score.
type ToolMatch struct {
	Tool  llm.ToolDef
	Score float32 // cosine similarity in [-1, 1]; higher = more relevant
}

// indexedTool stores a tool with its precomputed embedding.
type indexedTool struct {
	def       llm.ToolDef
	embedding []float32 // L2-normalised
}

// ToolIndex performs semantic search over tool definitions.
type ToolIndex struct {
	tools    []indexedTool
	embedder Embedder
	log      *logger.Logger
}

// Config holds the dependencies for building a ToolIndex.
type Config struct {
	Embedder Embedder
	Log      *logger.Logger
}

// New creates a ToolIndex by embedding every tool's description and example
// queries. If embedding fails the index is still usable but Search will
// return empty results (fail-open).
func New(ctx context.Context, cfg Config, tools []llm.ToolDef) (*ToolIndex, error) {
	if cfg.Embedder == nil {
		return nil, fmt.Errorf("toolindex: embedder is required")
	}
	if cfg.Log == nil {
		return nil, fmt.Errorf("toolindex: logger is required")
	}

	idx := &ToolIndex{
		embedder: cfg.Embedder,
		log:      cfg.Log,
		tools:    make([]indexedTool, 0, len(tools)),
	}

	start := time.Now()

	// Build the text that will be embedded for each tool.
	texts := make([]string, len(tools))
	for i, t := range tools {
		texts[i] = embeddingText(t)
	}

	// Prefer batch embedding when available.
	embeddings, err := embed(ctx, cfg.Embedder, texts)
	if err != nil {
		cfg.Log.Error(ctx, "toolindex: embedding failed, index will return empty results",
			"error", err)
		for _, t := range tools {
			idx.tools = append(idx.tools, indexedTool{def: t})
		}
		return idx, nil
	}

	for i, t := range tools {
		idx.tools = append(idx.tools, indexedTool{
			def:       t,
			embedding: embeddings[i],
		})
	}

	cfg.Log.Info(ctx, "toolindex: index built",
		"tool_count", len(tools),
		"elapsed_ms", time.Since(start).Milliseconds())

	return idx, nil
}

// SearchOptions controls filtering behaviour for Search.
// A zero-value SearchOptions applies no filtering (backward compatible).
type SearchOptions struct {
	// Allowlist restricts results to tools whose Name is in the map.
	// A nil or empty map means all tools are eligible.
	Allowlist map[string]bool

	// MinScore excludes matches with a cosine similarity below this value.
	// Zero means no threshold.
	MinScore float32
}

// Search returns the top-K tools most relevant to message, sorted by
// descending similarity score. If topK <= 0 every indexed tool is returned.
//
// opts controls allowlist and score-threshold filtering. A zero-value
// SearchOptions applies no filtering.
//
// On embedding failure Search returns (nil, nil) so callers can fall back
// to a broader tool set.
func (idx *ToolIndex) Search(ctx context.Context, message string, topK int, opts SearchOptions) ([]ToolMatch, time.Duration, error) {
	start := time.Now()

	qVec, err := idx.embedder.Embed(ctx, message)
	if err != nil {
		return nil, time.Since(start), fmt.Errorf("embed query: %w", err)
	}

	hasAllowlist := len(opts.Allowlist) > 0

	matches := make([]ToolMatch, 0, len(idx.tools))
	for _, t := range idx.tools {
		if t.embedding == nil {
			continue
		}
		if hasAllowlist && !opts.Allowlist[t.def.Name] {
			continue
		}
		score := dot(qVec, t.embedding)
		if opts.MinScore > 0 && score < opts.MinScore {
			continue
		}
		matches = append(matches, ToolMatch{
			Tool:  t.def,
			Score: score,
		})
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

	if topK > 0 && topK < len(matches) {
		matches = matches[:topK]
	}

	return matches, time.Since(start), nil
}

// EmbeddedCount returns the number of tools that have a non-nil embedding.
// Useful for startup diagnostics to detect silent embedding failures.
func (idx *ToolIndex) EmbeddedCount() int {
	n := 0
	for _, t := range idx.tools {
		if t.embedding != nil {
			n++
		}
	}
	return n
}

// =========================================================================
// Helpers
// =========================================================================

// embeddingText builds the text that gets embedded for a tool.
func embeddingText(t llm.ToolDef) string {
	var b strings.Builder
	b.WriteString(t.Name)
	b.WriteString(" — ")
	b.WriteString(t.Description)
	for _, q := range t.ExampleQueries {
		b.WriteString(" | ")
		b.WriteString(q)
	}
	return b.String()
}

// embed uses BatchEmbedder if available, otherwise falls back to sequential.
func embed(ctx context.Context, e Embedder, texts []string) ([][]float32, error) {
	if be, ok := e.(BatchEmbedder); ok {
		vecs, err := be.BatchEmbed(ctx, texts)
		if err != nil {
			return nil, err
		}
		for i := range vecs {
			vecs[i] = normalise(vecs[i])
		}
		return vecs, nil
	}

	vecs := make([][]float32, len(texts))
	for i, t := range texts {
		v, err := e.Embed(ctx, t)
		if err != nil {
			return nil, fmt.Errorf("embed tool %d: %w", i, err)
		}
		vecs[i] = normalise(v)
	}
	return vecs, nil
}

// dot computes the dot product of two vectors (cosine similarity when both
// are L2-normalised).
func dot(a, b []float32) float32 {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var s float32
	for i := 0; i < n; i++ {
		s += a[i] * b[i]
	}
	return s
}

// normalise scales a vector to unit length.
func normalise(v []float32) []float32 {
	var sumSq float64
	for _, x := range v {
		sumSq += float64(x) * float64(x)
	}
	if sumSq == 0 {
		return v
	}
	norm := float32(1.0 / math.Sqrt(sumSq))
	out := make([]float32, len(v))
	for i, x := range v {
		out[i] = x * norm
	}
	return out
}
