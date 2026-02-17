package toolindex

import (
	"context"
	"hash/fnv"
	"math"
)

// MockEmbedder produces deterministic embeddings derived from text hashes.
// Useful for unit tests that need stable, repeatable similarity scores
// without a real embedding service.
type MockEmbedder struct {
	Dim int // vector dimension; defaults to 128
}

func (m *MockEmbedder) dim() int {
	if m.Dim > 0 {
		return m.Dim
	}
	return 128
}

// Embed returns a normalised pseudo-random vector seeded by the text hash.
func (m *MockEmbedder) Embed(_ context.Context, text string) ([]float32, error) {
	h := fnv.New64a()
	h.Write([]byte(text))
	seed := h.Sum64()

	d := m.dim()
	vec := make([]float32, d)
	var sumSq float64
	for i := range vec {
		seed = seed*6364136223846793005 + 1442695040888963407 // LCG
		vec[i] = float32(int32(seed>>33)) / float32(math.MaxInt32)
		sumSq += float64(vec[i]) * float64(vec[i])
	}

	norm := float32(1.0 / math.Sqrt(sumSq))
	for i := range vec {
		vec[i] *= norm
	}
	return vec, nil
}

// BatchEmbed calls Embed sequentially for each text.
func (m *MockEmbedder) BatchEmbed(ctx context.Context, texts []string) ([][]float32, error) {
	out := make([][]float32, len(texts))
	for i, t := range texts {
		v, err := m.Embed(ctx, t)
		if err != nil {
			return nil, err
		}
		out[i] = v
	}
	return out, nil
}
