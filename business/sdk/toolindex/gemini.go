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

// GeminiEmbedder calls Google's Gemini embedding API.
type GeminiEmbedder struct {
	apiKey string
	model  string
	client *http.Client
}

// NewGeminiEmbedder creates an embedder backed by Gemini's embedding API.
// model should be an embedding model such as "gemini-embedding-001".
func NewGeminiEmbedder(apiKey, model string) *GeminiEmbedder {
	if model == "" {
		model = "gemini-embedding-001"
	}
	return &GeminiEmbedder{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// Embed returns a single embedding vector for text.
func (g *GeminiEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:embedContent",
		g.model,
	)

	body, err := json.Marshal(map[string]any{
		"model": "models/" + g.model,
		"content": map[string]any{
			"parts": []map[string]string{
				{"text": text},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", g.apiKey)

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gemini embedding %d: %s", resp.StatusCode, string(b))
	}

	var result struct {
		Embedding struct {
			Values []float32 `json:"values"`
		} `json:"embedding"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return result.Embedding.Values, nil
}

// BatchEmbed returns embedding vectors for multiple texts in one call.
func (g *GeminiEmbedder) BatchEmbed(ctx context.Context, texts []string) ([][]float32, error) {
	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:batchEmbedContents",
		g.model,
	)

	requests := make([]map[string]any, len(texts))
	for i, text := range texts {
		requests[i] = map[string]any{
			"model": "models/" + g.model,
			"content": map[string]any{
				"parts": []map[string]string{
					{"text": text},
				},
			},
		}
	}

	body, err := json.Marshal(map[string]any{
		"requests": requests,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", g.apiKey)

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gemini batch embedding %d: %s", resp.StatusCode, string(b))
	}

	var result struct {
		Embeddings []struct {
			Values []float32 `json:"values"`
		} `json:"embeddings"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	vecs := make([][]float32, len(result.Embeddings))
	for i, emb := range result.Embeddings {
		vecs[i] = emb.Values
	}
	return vecs, nil
}
