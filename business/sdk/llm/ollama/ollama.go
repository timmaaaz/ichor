// Package ollama implements the llm.Provider interface using Ollama's
// OpenAI-compatible chat completions API.
package ollama

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/timmaaaz/ichor/business/sdk/llm"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Provider calls Ollama's OpenAI-compatible endpoint with streaming and tool
// support.
type Provider struct {
	host      string
	model     string
	maxTokens int
	log       *logger.Logger
	client    *http.Client
}

// NewProvider builds an Ollama provider from a host URL (e.g.
// "http://ollama-service:11434"), model name, and token limit.
// It starts a background goroutine to pre-load the model so the first
// user request doesn't block on model loading.
func NewProvider(host, model string, maxTokens int, log *logger.Logger) *Provider {
	p := &Provider{
		host:      strings.TrimRight(host, "/"),
		model:     model,
		maxTokens: maxTokens,
		log:       log,
		client:    &http.Client{},
	}

	go p.warmModel()

	return p
}

// warmModel sends a lightweight request to Ollama to trigger model loading.
// This runs in the background so the first real chat request doesn't time out
// waiting for the model to load into memory.
func (p *Provider) warmModel() {
	p.log.Info(context.Background(), "AGENT-CHAT: ollama warming model", "model", p.model)

	start := time.Now()

	body, _ := json.Marshal(map[string]any{
		"model":      p.model,
		"messages":   []map[string]string{{"role": "user", "content": "hi"}},
		"max_tokens": 1,
		"stream":     false,
	})

	resp, err := p.client.Post(p.host+"/v1/chat/completions", "application/json", bytes.NewReader(body))
	if err != nil {
		p.log.Error(context.Background(), "AGENT-CHAT: ollama warm failed", "elapsed", time.Since(start), "error", err)
		return
	}
	resp.Body.Close()

	p.log.Info(context.Background(), "AGENT-CHAT: ollama model ready", "model", p.model, "elapsed", time.Since(start))
}

// StreamChat implements llm.Provider. It sends a request to Ollama's
// OpenAI-compatible endpoint and translates SSE events into llm.StreamEvent
// values on the returned channel.
func (p *Provider) StreamChat(ctx context.Context, req llm.ChatRequest) (<-chan llm.StreamEvent, error) {
	msgs := toOpenAIMessages(req.SystemPrompt, req.Messages)
	tools := toOpenAITools(req.Tools)

	maxTokens := p.maxTokens
	if req.MaxTokens > 0 {
		maxTokens = req.MaxTokens
	}

	body := openAIRequest{
		Model:     p.model,
		Messages:  msgs,
		Stream:    true,
		MaxTokens: maxTokens,
	}
	if len(tools) > 0 {
		body.Tools = tools
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("ollama: marshal request: %w", err)
	}

	p.log.Info(ctx, "AGENT-CHAT: ollama request",
		"model", p.model,
		"messages", len(msgs),
		"tools", len(tools),
		"max_tokens", maxTokens)

	start := time.Now()

	// Use a context without the parent's deadline so model loading on
	// first request (or slow inference) isn't killed by the HTTP server's
	// write timeout. Cancellation still propagates on client disconnect.
	llmCtx, llmCancel := context.WithCancel(context.Background())
	go func() {
		<-ctx.Done()
		llmCancel()
	}()

	url := p.host + "/v1/chat/completions"
	httpReq, err := http.NewRequestWithContext(llmCtx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		llmCancel()
		return nil, fmt.Errorf("ollama: new request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		p.log.Error(ctx, "AGENT-CHAT: ollama request failed", "elapsed", time.Since(start), "error", err)
		return nil, fmt.Errorf("ollama: do request: %w", err)
	}

	p.log.Info(ctx, "AGENT-CHAT: ollama stream started", "status", resp.StatusCode, "elapsed", time.Since(start))

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama: unexpected status %d: %s", resp.StatusCode, string(b))
	}

	ch := make(chan llm.StreamEvent, 64)

	go func() {
		defer llmCancel()
		defer close(ch)
		defer resp.Body.Close()

		sent := func(e llm.StreamEvent) bool {
			select {
			case ch <- e:
				return true
			case <-ctx.Done():
				return false
			}
		}

		sent(llm.StreamEvent{Type: llm.EventMessageStart})

		scanner := bufio.NewScanner(resp.Body)
		stopForTools := false
		var toolNames []string

		for scanner.Scan() {
			line := scanner.Text()

			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")

			if data == "[DONE]" {
				break
			}

			var chunk openAIChunk
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				sent(llm.StreamEvent{Type: llm.EventError, Err: fmt.Errorf("ollama: unmarshal chunk: %w", err)})
				return
			}

			if len(chunk.Choices) == 0 {
				continue
			}
			choice := chunk.Choices[0]
			delta := choice.Delta

			// Text content delta.
			if delta.Content != "" {
				if !sent(llm.StreamEvent{Type: llm.EventContentDelta, Text: delta.Content}) {
					return
				}
			}

			// Tool calls.
			for _, tc := range delta.ToolCalls {
				if tc.Function.Name != "" {
					p.log.Info(ctx, "AGENT-CHAT: ollama tool_call received",
						"name", tc.Function.Name,
						"tool_call_id", tc.ID)
					toolNames = append(toolNames, tc.Function.Name)

					if !sent(llm.StreamEvent{
						Type:         llm.EventToolUseStart,
						ToolCallID:   tc.ID,
						ToolCallName: tc.Function.Name,
					}) {
						return
					}
				}
				if tc.Function.Arguments != "" {
					if !sent(llm.StreamEvent{
						Type:         llm.EventToolUseInput,
						PartialInput: tc.Function.Arguments,
					}) {
						return
					}
				}
			}

			// Check finish reason. Some Ollama models don't return
			// "tool_calls" as the finish reason even when tool calls
			// are present, so also check if we accumulated any.
			if choice.FinishReason == "tool_calls" || len(toolNames) > 0 {
				stopForTools = true
			}
		}

		if err := scanner.Err(); err != nil {
			p.log.Error(ctx, "AGENT-CHAT: ollama scan error", "elapsed", time.Since(start), "error", err)
			sent(llm.StreamEvent{Type: llm.EventError, Err: fmt.Errorf("ollama: scan: %w", err)})
			return
		}

		p.log.Info(ctx, "AGENT-CHAT: ollama stream complete",
			"elapsed", time.Since(start),
			"stop_for_tools", stopForTools,
			"tool_calls", toolNames)

		sent(llm.StreamEvent{
			Type:           llm.EventMessageComplete,
			StopForToolUse: stopForTools,
		})
	}()

	return ch, nil
}

// =========================================================================
// OpenAI-compatible request/response types
// =========================================================================

type openAIRequest struct {
	Model     string          `json:"model"`
	Messages  []openAIMessage `json:"messages"`
	Stream    bool            `json:"stream"`
	MaxTokens int             `json:"max_tokens,omitempty"`
	Tools     []openAITool    `json:"tools,omitempty"`
}

type openAIMessage struct {
	Role       string           `json:"role"`
	Content    string           `json:"content,omitempty"`
	ToolCalls  []openAIToolCall `json:"tool_calls,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
}

type openAITool struct {
	Type     string             `json:"type"`
	Function openAIToolFunction `json:"function"`
}

type openAIToolFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

type openAIToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type openAIChunk struct {
	Choices []openAIChunkChoice `json:"choices"`
}

type openAIChunkChoice struct {
	Delta        openAIChunkDelta `json:"delta"`
	FinishReason string           `json:"finish_reason"`
}

type openAIChunkDelta struct {
	Content   string           `json:"content"`
	ToolCalls []openAIToolCall `json:"tool_calls"`
}

// =========================================================================
// Conversion helpers
// =========================================================================

func toOpenAIMessages(systemPrompt string, msgs []llm.Message) []openAIMessage {
	out := make([]openAIMessage, 0, len(msgs)+1)

	if systemPrompt != "" {
		out = append(out, openAIMessage{
			Role:    "system",
			Content: systemPrompt,
		})
	}

	for _, m := range msgs {
		switch m.Role {
		case "user":
			if m.Content != "" {
				out = append(out, openAIMessage{
					Role:    "user",
					Content: m.Content,
				})
			}
			// Tool results become separate messages with role=tool.
			for _, tr := range m.ToolResults {
				content := tr.Content
				if tr.IsError {
					content = "ERROR: " + content
				}
				out = append(out, openAIMessage{
					Role:       "tool",
					Content:    content,
					ToolCallID: tr.ToolUseID,
				})
			}

		case "assistant":
			msg := openAIMessage{
				Role:    "assistant",
				Content: m.Content,
			}
			for _, tc := range m.ToolCalls {
				msg.ToolCalls = append(msg.ToolCalls, openAIToolCall{
					ID:   tc.ID,
					Type: "function",
					Function: struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					}{
						Name:      tc.Name,
						Arguments: string(tc.Input),
					},
				})
			}
			out = append(out, msg)
		}
	}

	return out
}

func toOpenAITools(defs []llm.ToolDef) []openAITool {
	out := make([]openAITool, len(defs))
	for i, d := range defs {
		out[i] = openAITool{
			Type: "function",
			Function: openAIToolFunction{
				Name:        d.Name,
				Description: d.Description,
				Parameters:  d.InputSchema,
			},
		}
	}
	return out
}
