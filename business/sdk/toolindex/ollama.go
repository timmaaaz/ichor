package toolindex

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OllamaEmbedder calls Ollama's /api/embed endpoint.
type OllamaEmbedder struct {
	baseURL string
	model   string
	client  *http.Client
}

// NewOllamaEmbedder creates an embedder backed by a local Ollama instance.
// model should be an embedding model such as "nomic-embed-text".
func NewOllamaEmbedder(baseURL, model string) *OllamaEmbedder {
	return &OllamaEmbedder{
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

// Embed returns a single embedding vector for text.
func (o *OllamaEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	vecs, err := o.call(ctx, text)
	if err != nil {
		return nil, err
	}
	if len(vecs) == 0 {
		return nil, fmt.Errorf("ollama: no embeddings returned")
	}
	return vecs[0], nil
}

// BatchEmbed returns embedding vectors for multiple texts in one call.
func (o *OllamaEmbedder) BatchEmbed(ctx context.Context, texts []string) ([][]float32, error) {
	return o.call(ctx, texts)
}

// call posts to /api/embed. input can be a string or []string.
func (o *OllamaEmbedder) call(ctx context.Context, input any) ([][]float32, error) {
	body, err := json.Marshal(map[string]any{
		"model": o.model,
		"input": input,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.baseURL+"/api/embed", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama %d: %s", resp.StatusCode, string(b))
	}

	var result struct {
		Embeddings [][]float32 `json:"embeddings"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return result.Embeddings, nil
}
