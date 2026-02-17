// Package gemini implements the llm.Provider interface using Google Gemini's
// OpenAI-compatible chat completions endpoint.
package gemini

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
	"github.com/timmaaaz/ichor/business/sdk/llm/openaicompat"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// endpoint is the Gemini OpenAI-compatible chat completions URL.
const endpoint = "https://generativelanguage.googleapis.com/v1beta/openai/chat/completions"

// DefaultModel is used when no model is specified.
const DefaultModel = "gemini-2.0-flash"

// Provider calls Gemini's OpenAI-compatible endpoint with streaming and tool
// support.
type Provider struct {
	apiKey    string
	model     string
	maxTokens int
	log       *logger.Logger
	talkLog   *logger.Logger
	client    *http.Client
}

// NewProvider builds a Gemini provider from an API key, model name, and token
// limit. If model is empty, DefaultModel is used.
func NewProvider(apiKey, model string, maxTokens int, log, talkLog *logger.Logger) *Provider {
	if model == "" {
		model = DefaultModel
	}
	return &Provider{
		apiKey:    apiKey,
		model:     model,
		maxTokens: maxTokens,
		log:       log,
		talkLog:   talkLog,
		client:    &http.Client{Timeout: 60 * time.Second},
	}
}

// StreamChat implements llm.Provider. It sends a streaming request to Gemini's
// OpenAI-compatible endpoint and translates SSE events into llm.StreamEvent
// values on the returned channel.
func (p *Provider) StreamChat(ctx context.Context, req llm.ChatRequest) (<-chan llm.StreamEvent, error) {
	msgs := openaicompat.ToMessages(req.SystemPrompt, req.Messages)
	tools := openaicompat.ToTools(req.Tools)

	maxTokens := p.maxTokens
	if req.MaxTokens > 0 {
		maxTokens = req.MaxTokens
	}

	body := openaicompat.Request{
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
		return nil, fmt.Errorf("gemini: marshal request: %w", err)
	}

	p.log.Info(ctx, "AGENT-CHAT: gemini request",
		"model", p.model,
		"messages", len(msgs),
		"tools", len(tools),
		"max_tokens", maxTokens)

	// Stage 3: Log the full sent packet (untruncated).
	if p.talkLog != nil {
		p.talkLog.Info(ctx, "TALK-LOG: sent_packet",
			"stage", "sent_packet",
			"session_id", llm.SessionID(ctx),
			"provider", "gemini",
			"model", p.model,
			"payload", json.RawMessage(payload))
	}

	start := time.Now()

	// Use a context without the parent's deadline so streaming isn't
	// killed by the HTTP server's write timeout. Cancellation still
	// propagates on client disconnect.
	llmCtx, llmCancel := context.WithCancel(context.Background())
	go func() {
		<-ctx.Done()
		llmCancel()
	}()

	httpReq, err := http.NewRequestWithContext(llmCtx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		llmCancel()
		return nil, fmt.Errorf("gemini: new request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		p.log.Error(ctx, "AGENT-CHAT: gemini request failed", "elapsed", time.Since(start), "error", err)
		return nil, fmt.Errorf("gemini: do request: %w", err)
	}

	p.log.Info(ctx, "AGENT-CHAT: gemini stream started", "status", resp.StatusCode, "elapsed", time.Since(start))

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gemini: HTTP %d: %s", resp.StatusCode, string(b))
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

		// Accumulators for talk-log stage 4.
		var accText string
		type talkToolCall struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			Input string `json:"input"`
		}
		var accToolCalls []talkToolCall
		var currentInput string

		for scanner.Scan() {
			line := scanner.Text()

			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")

			if data == "[DONE]" {
				break
			}

			var chunk openaicompat.Chunk
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				sent(llm.StreamEvent{Type: llm.EventError, Err: fmt.Errorf("gemini: unmarshal chunk: %w", err)})
				return
			}

			if len(chunk.Choices) == 0 {
				continue
			}
			choice := chunk.Choices[0]
			delta := choice.Delta

			// Text content delta.
			if delta.Content != "" {
				accText += delta.Content
				if !sent(llm.StreamEvent{Type: llm.EventContentDelta, Text: delta.Content}) {
					return
				}
			}

			// Tool calls.
			for _, tc := range delta.ToolCalls {
				if tc.Function.Name != "" {
					p.log.Info(ctx, "AGENT-CHAT: gemini tool_call received",
						"name", tc.Function.Name,
						"tool_call_id", tc.ID)
					toolNames = append(toolNames, tc.Function.Name)

					// Finalize any previous tool call accumulation.
					if len(accToolCalls) > 0 {
						accToolCalls[len(accToolCalls)-1].Input = currentInput
					}
					accToolCalls = append(accToolCalls, talkToolCall{ID: tc.ID, Name: tc.Function.Name})
					currentInput = ""

					if !sent(llm.StreamEvent{
						Type:         llm.EventToolUseStart,
						ToolCallID:   tc.ID,
						ToolCallName: tc.Function.Name,
					}) {
						return
					}
				}
				if tc.Function.Arguments != "" {
					currentInput += tc.Function.Arguments
					if !sent(llm.StreamEvent{
						Type:         llm.EventToolUseInput,
						PartialInput: tc.Function.Arguments,
					}) {
						return
					}
				}
			}

			if choice.FinishReason == "tool_calls" || (choice.FinishReason == "stop" && len(toolNames) > 0) {
				stopForTools = true
			}
		}

		// Finalize the last tool call input.
		if len(accToolCalls) > 0 {
			accToolCalls[len(accToolCalls)-1].Input = currentInput
		}

		if err := scanner.Err(); err != nil {
			p.log.Error(ctx, "AGENT-CHAT: gemini scan error", "elapsed", time.Since(start), "error", err)
			sent(llm.StreamEvent{Type: llm.EventError, Err: fmt.Errorf("gemini: scan: %w", err)})
			return
		}

		p.log.Info(ctx, "AGENT-CHAT: gemini stream complete",
			"elapsed", time.Since(start),
			"stop_for_tools", stopForTools,
			"tool_calls", toolNames)

		// Stage 4: Log the full LLM return (untruncated).
		if p.talkLog != nil {
			p.talkLog.Info(ctx, "TALK-LOG: llm_return",
				"stage", "llm_return",
				"session_id", llm.SessionID(ctx),
				"provider", "gemini",
				"assistant_text", accText,
				"tool_calls", accToolCalls,
				"stop_for_tools", stopForTools,
				"elapsed", time.Since(start))
		}

		sent(llm.StreamEvent{
			Type:           llm.EventMessageComplete,
			StopForToolUse: stopForTools,
		})
	}()

	return ch, nil
}
