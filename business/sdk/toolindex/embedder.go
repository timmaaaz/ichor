// Package toolindex provides embedding-based tool retrieval (Tool RAG) for
// the agent chat system. It embeds tool descriptions and example queries at
// startup, then uses cosine similarity to find the most relevant tools for
// each user message.
//
// At 30-50 tools the entire index fits in memory and brute-force cosine
// similarity completes in under 1 ms, so no external vector store is needed.
package toolindex

import "context"

// Embedder converts text into a dense vector for semantic similarity search.
type Embedder interface {
	// Embed returns a normalised embedding vector for the given text.
	Embed(ctx context.Context, text string) ([]float32, error)
}

// BatchEmbedder optionally embeds multiple texts in a single round-trip.
// If the concrete Embedder also implements BatchEmbedder the index builder
// will prefer it for startup efficiency.
type BatchEmbedder interface {
	Embedder
	BatchEmbed(ctx context.Context, texts []string) ([][]float32, error)
}
