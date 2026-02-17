// Package claude implements the llm.Provider interface using the Anthropic SDK.
package claude

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/timmaaaz/ichor/business/sdk/llm"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Provider calls the Claude Messages API with streaming and tool support.
type Provider struct {
	client    anthropic.Client
	model     anthropic.Model
	maxTokens int64
	log       *logger.Logger
	talkLog   *logger.Logger
}

// NewProvider builds a Claude provider from an API key, model name, and
// token limit.
func NewProvider(apiKey string, model string, maxTokens int, log, talkLog *logger.Logger) *Provider {
	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &Provider{
		client:    client,
		model:     anthropic.Model(model),
		maxTokens: int64(maxTokens),
		log:       log,
		talkLog:   talkLog,
	}
}

// StreamChat implements llm.Provider. It opens a streaming request to Claude
// and translates SDK events into llm.StreamEvent values on the returned
// channel.
func (p *Provider) StreamChat(ctx context.Context, req llm.ChatRequest) (<-chan llm.StreamEvent, error) {
	params := anthropic.MessageNewParams{
		Model:     p.model,
		MaxTokens: p.maxTokens,
		Messages:  toAnthropicMessages(req.Messages),
		Tools:     toAnthropicTools(req.Tools),
	}

	if req.SystemPrompt != "" {
		params.System = []anthropic.TextBlockParam{
			{Text: req.SystemPrompt},
		}
	}

	// Stage 3: Log the full sent packet (untruncated).
	if p.talkLog != nil {
		if paramsJSON, err := json.Marshal(params); err == nil {
			p.talkLog.Info(ctx, "TALK-LOG: sent_packet",
				"stage", "sent_packet",
				"session_id", llm.SessionID(ctx),
				"provider", "claude",
				"model", string(p.model),
				"payload", json.RawMessage(paramsJSON))
		}
	}

	start := time.Now()
	stream := p.client.Messages.NewStreaming(ctx, params)

	ch := make(chan llm.StreamEvent, 64)

	go func() {
		defer close(ch)
		defer stream.Close()

		sent := func(e llm.StreamEvent) bool {
			select {
			case ch <- e:
				return true
			case <-ctx.Done():
				return false
			}
		}

		sent(llm.StreamEvent{Type: llm.EventMessageStart})

		msg := anthropic.Message{}
		for stream.Next() {
			event := stream.Current()
			if err := msg.Accumulate(event); err != nil {
				sent(llm.StreamEvent{Type: llm.EventError, Err: fmt.Errorf("accumulate: %w", err)})
				return
			}

			switch ev := event.AsAny().(type) {
			case anthropic.ContentBlockStartEvent:
				if ev.ContentBlock.Name != "" {
					if !sent(llm.StreamEvent{
						Type:         llm.EventToolUseStart,
						ToolCallID:   ev.ContentBlock.ID,
						ToolCallName: ev.ContentBlock.Name,
					}) {
						return
					}
				}

			case anthropic.ContentBlockDeltaEvent:
				if ev.Delta.Text != "" {
					if !sent(llm.StreamEvent{Type: llm.EventContentDelta, Text: ev.Delta.Text}) {
						return
					}
				}
				if ev.Delta.PartialJSON != "" {
					if !sent(llm.StreamEvent{Type: llm.EventToolUseInput, PartialInput: ev.Delta.PartialJSON}) {
						return
					}
				}
			}
		}

		if err := stream.Err(); err != nil {
			sent(llm.StreamEvent{Type: llm.EventError, Err: err})
			return
		}

		// Determine whether the LLM stopped to invoke tools.
		stopForTools := msg.StopReason == anthropic.StopReasonToolUse

		// Stage 4: Log the full LLM return (untruncated).
		if p.talkLog != nil {
			// Extract text and tool calls from accumulated message.
			var assistantText string
			type talkToolCall struct {
				ID    string `json:"id"`
				Name  string `json:"name"`
				Input string `json:"input"`
			}
			var toolCalls []talkToolCall

			for _, block := range msg.Content {
				switch v := block.AsAny().(type) {
				case anthropic.TextBlock:
					assistantText += v.Text
				case anthropic.ToolUseBlock:
					toolCalls = append(toolCalls, talkToolCall{
						ID:    block.ID,
						Name:  block.Name,
						Input: v.JSON.Input.Raw(),
					})
				}
			}

			p.talkLog.Info(ctx, "TALK-LOG: llm_return",
				"stage", "llm_return",
				"session_id", llm.SessionID(ctx),
				"provider", "claude",
				"assistant_text", assistantText,
				"tool_calls", toolCalls,
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

// AccumulatedToolCalls extracts completed tool-use blocks from a fully
// accumulated anthropic.Message. The chatapi handler calls this after
// draining the stream to learn which tools the LLM requested.
func AccumulatedToolCalls(msg *anthropic.Message) []llm.ToolCall {
	var calls []llm.ToolCall
	for _, block := range msg.Content {
		if variant, ok := block.AsAny().(anthropic.ToolUseBlock); ok {
			raw := json.RawMessage(variant.JSON.Input.Raw())
			calls = append(calls, llm.ToolCall{
				ID:    block.ID,
				Name:  block.Name,
				Input: raw,
			})
		}
	}
	return calls
}

// =========================================================================
// Conversion helpers
// =========================================================================

func toAnthropicMessages(msgs []llm.Message) []anthropic.MessageParam {
	out := make([]anthropic.MessageParam, 0, len(msgs))
	for _, m := range msgs {
		switch m.Role {
		case "user":
			blocks := make([]anthropic.ContentBlockParamUnion, 0)
			if m.Content != "" {
				blocks = append(blocks, anthropic.NewTextBlock(m.Content))
			}
			for _, tr := range m.ToolResults {
				blocks = append(blocks, anthropic.NewToolResultBlock(
					tr.ToolUseID,
					tr.Content,
					tr.IsError,
				))
			}
			out = append(out, anthropic.NewUserMessage(blocks...))

		case "assistant":
			blocks := make([]anthropic.ContentBlockParamUnion, 0)
			if m.Content != "" {
				blocks = append(blocks, anthropic.NewTextBlock(m.Content))
			}
			for _, tc := range m.ToolCalls {
				var input any
				if err := json.Unmarshal(tc.Input, &input); err != nil {
					input = map[string]any{}
				}
				blocks = append(blocks, anthropic.NewToolUseBlock(tc.ID, input, tc.Name))
			}
			out = append(out, anthropic.NewAssistantMessage(blocks...))
		}
	}
	return out
}

func toAnthropicTools(defs []llm.ToolDef) []anthropic.ToolUnionParam {
	out := make([]anthropic.ToolUnionParam, len(defs))
	for i, d := range defs {
		var props map[string]any
		if err := json.Unmarshal(d.InputSchema, &props); err != nil {
			props = map[string]any{}
		}

		out[i] = anthropic.ToolUnionParam{
			OfTool: &anthropic.ToolParam{
				Name:        d.Name,
				Description: anthropic.String(d.Description),
				InputSchema: anthropic.ToolInputSchemaParam{
					Properties: props,
				},
			},
		}
	}
	return out
}
