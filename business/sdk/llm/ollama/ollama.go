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
	"github.com/timmaaaz/ichor/business/sdk/llm/openaicompat"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Provider calls Ollama's OpenAI-compatible endpoint with streaming and tool
// support.
type Provider struct {
	host            string
	model           string
	maxTokens       int
	thinkingEnabled bool // whether to allow model thinking (qwen3 thinks by default)
	log             *logger.Logger
	talkLog         *logger.Logger
	client          *http.Client
}

// NewProvider builds an Ollama provider from a host URL (e.g.
// "http://ollama-service:11434"), model name, token limit, and optional
// thinking effort level ("high", "medium", "low", "none", or "" to disable).
// It starts a background goroutine to pre-load the model so the first
// user request doesn't block on model loading.
func NewProvider(host, model string, maxTokens int, thinkingEffort string, log, talkLog *logger.Logger) *Provider {
	p := &Provider{
		host:            strings.TrimRight(host, "/"),
		model:           model,
		maxTokens:       maxTokens,
		thinkingEnabled: thinkingEffort != "" && thinkingEffort != "none",
		log:             log,
		talkLog:         talkLog,
		client:          &http.Client{},
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
	// When thinking is disabled, append qwen3's native /no_think toggle
	// to the system prompt so the model skips chain-of-thought output.
	sysPrompt := req.SystemPrompt
	if !p.thinkingEnabled && sysPrompt != "" {
		sysPrompt += " /no_think"
	}

	msgs := openaicompat.ToMessages(sysPrompt, req.Messages)
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
		return nil, fmt.Errorf("ollama: marshal request: %w", err)
	}

	p.log.Info(ctx, "AGENT-CHAT: ollama request",
		"model", p.model,
		"messages", len(msgs),
		"tools", len(tools),
		"max_tokens", maxTokens,
		"thinking", p.thinkingEnabled)

	// Stage 3: Log the full sent packet (untruncated).
	if p.talkLog != nil {
		p.talkLog.Info(ctx, "TALK-LOG: sent_packet",
			"stage", "sent_packet",
			"session_id", llm.SessionID(ctx),
			"provider", "ollama",
			"model", p.model,
			"payload", json.RawMessage(payload))
	}

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

		// thinkFilter strips <think>...</think> tags from the content
		// stream and routes thinking content to EventThinkingDelta.
		// Qwen3 outputs thinking by default inside the content field.
		tf := &thinkFilter{enabled: p.thinkingEnabled}

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
				sent(llm.StreamEvent{Type: llm.EventError, Err: fmt.Errorf("ollama: unmarshal chunk: %w", err)})
				return
			}

			if len(chunk.Choices) == 0 {
				continue
			}
			choice := chunk.Choices[0]
			delta := choice.Delta

			// Some Ollama versions separate reasoning into its own field.
			if delta.Reasoning != "" {
				if !sent(llm.StreamEvent{Type: llm.EventThinkingDelta, ThinkingText: delta.Reasoning}) {
					return
				}
			}

			// Text content delta — filter <think> tags from qwen3 output.
			if delta.Content != "" {
				thinking, content := tf.process(delta.Content)
				if thinking != "" {
					if !sent(llm.StreamEvent{Type: llm.EventThinkingDelta, ThinkingText: thinking}) {
						return
					}
				}
				if content != "" {
					accText += content
					if !sent(llm.StreamEvent{Type: llm.EventContentDelta, Text: content}) {
						return
					}
				}
			}

			// Tool calls.
			for _, tc := range delta.ToolCalls {
				if tc.Function.Name != "" {
					p.log.Info(ctx, "AGENT-CHAT: ollama tool_call received",
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

			// Check finish reason. Some Ollama models don't return
			// "tool_calls" as the finish reason even when tool calls
			// are present, so also check if we accumulated any.
			if choice.FinishReason == "tool_calls" || len(toolNames) > 0 {
				stopForTools = true
			}
		}

		// Flush any buffered content from the think filter.
		if remaining := tf.flush(); remaining != "" {
			accText += remaining
			sent(llm.StreamEvent{Type: llm.EventContentDelta, Text: remaining})
		}

		// Finalize the last tool call input.
		if len(accToolCalls) > 0 {
			accToolCalls[len(accToolCalls)-1].Input = currentInput
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

		// Stage 4: Log the full LLM return (untruncated).
		if p.talkLog != nil {
			p.talkLog.Info(ctx, "TALK-LOG: llm_return",
				"stage", "llm_return",
				"session_id", llm.SessionID(ctx),
				"provider", "ollama",
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

// =========================================================================
// Think tag filter
// =========================================================================

// thinkFilter is a streaming state machine that separates <think>...</think>
// content from regular content in qwen3 output. Qwen3 outputs thinking by
// default inside the content field as: <think>reasoning...</think>response.
//
// Thinking always comes first, before the visible response. Tags can be split
// across chunk boundaries (e.g. chunk 1: "<thi", chunk 2: "nk>hello").
type thinkFilter struct {
	enabled bool   // false = pass all content through as-is (no filtering)
	state   int    // 0=init, 1=thinking, 2=done (past </think>)
	buf     string // buffer for partial tag matching
}

const (
	tfInit     = 0
	tfThinking = 1
	tfDone     = 2
)

// process takes a content chunk and returns (thinking, content) text.
// Either or both may be empty for a given chunk.
func (f *thinkFilter) process(chunk string) (thinking, content string) {
	if !f.enabled || f.state == tfDone {
		return "", chunk
	}

	f.buf += chunk

	switch f.state {
	case tfInit:
		// Look for <think> at the start of the stream.
		const openTag = "<think>"
		if len(f.buf) < len(openTag) {
			// Could still be a partial <think> prefix.
			if strings.HasPrefix(openTag, f.buf) {
				return "", "" // buffer more
			}
			// Not a think block — flush buffer as content.
			out := f.buf
			f.buf = ""
			f.state = tfDone
			return "", out
		}

		if strings.HasPrefix(f.buf, openTag) {
			f.state = tfThinking
			f.buf = f.buf[len(openTag):]
			return f.drainThinking()
		}

		// Content doesn't start with <think>.
		out := f.buf
		f.buf = ""
		f.state = tfDone
		return "", out

	case tfThinking:
		return f.drainThinking()
	}

	return "", ""
}

// drainThinking extracts thinking content up to </think>, returning any
// content after the close tag. Keeps a small buffer to handle partial tags.
func (f *thinkFilter) drainThinking() (thinking, content string) {
	const closeTag = "</think>"

	idx := strings.Index(f.buf, closeTag)
	if idx >= 0 {
		// Found the close tag.
		thinking = f.buf[:idx]
		content = strings.TrimLeft(f.buf[idx+len(closeTag):], "\n")
		f.buf = ""
		f.state = tfDone
		return thinking, content
	}

	// No close tag yet. Emit everything except the last len(closeTag)-1
	// chars, which could be a partial </think>.
	safe := len(f.buf) - (len(closeTag) - 1)
	if safe > 0 {
		thinking = f.buf[:safe]
		f.buf = f.buf[safe:]
	}
	return thinking, ""
}

// flush returns any remaining buffered content when the stream ends.
func (f *thinkFilter) flush() string {
	if f.buf == "" {
		return ""
	}
	out := f.buf
	f.buf = ""
	return out
}
