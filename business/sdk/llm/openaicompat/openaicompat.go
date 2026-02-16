// Package openaicompat provides shared types and helpers for LLM providers
// that use the OpenAI-compatible chat completions API (e.g. Ollama, Gemini).
package openaicompat

import (
	"encoding/json"

	"github.com/timmaaaz/ichor/business/sdk/llm"
)

// =========================================================================
// Request/response types
// =========================================================================

// Request is the JSON body sent to an OpenAI-compatible chat completions
// endpoint.
type Request struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	Stream    bool      `json:"stream"`
	MaxTokens int       `json:"max_tokens,omitempty"`
	Tools     []Tool    `json:"tools,omitempty"`
}

// Message is a single message in the OpenAI chat format.
type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

// Tool wraps a function definition for the OpenAI tools array.
type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

// ToolFunction describes a callable function within a Tool.
type ToolFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

// ToolCall represents a tool invocation in a streaming delta or message.
type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// Chunk is a single SSE chunk from a streaming chat completions response.
type Chunk struct {
	Choices []ChunkChoice `json:"choices"`
}

// ChunkChoice is one choice within a streaming chunk.
type ChunkChoice struct {
	Delta        ChunkDelta `json:"delta"`
	FinishReason string     `json:"finish_reason"`
}

// ChunkDelta holds incremental content and tool call data.
type ChunkDelta struct {
	Content   string     `json:"content"`
	Reasoning string     `json:"reasoning"`
	ToolCalls []ToolCall `json:"tool_calls"`
}

// =========================================================================
// Conversion helpers
// =========================================================================

// ToMessages converts the internal llm.Message slice and an optional system
// prompt into the OpenAI message format.
func ToMessages(systemPrompt string, msgs []llm.Message) []Message {
	out := make([]Message, 0, len(msgs)+1)

	if systemPrompt != "" {
		out = append(out, Message{
			Role:    "system",
			Content: systemPrompt,
		})
	}

	for _, m := range msgs {
		switch m.Role {
		case "user":
			if m.Content != "" {
				out = append(out, Message{
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
				out = append(out, Message{
					Role:       "tool",
					Content:    content,
					ToolCallID: tr.ToolUseID,
				})
			}

		case "assistant":
			msg := Message{
				Role:    "assistant",
				Content: m.Content,
			}
			for _, tc := range m.ToolCalls {
				msg.ToolCalls = append(msg.ToolCalls, ToolCall{
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

// ToTools converts the internal llm.ToolDef slice into the OpenAI tools
// format.
func ToTools(defs []llm.ToolDef) []Tool {
	out := make([]Tool, len(defs))
	for i, d := range defs {
		out[i] = Tool{
			Type: "function",
			Function: ToolFunction{
				Name:        d.Name,
				Description: d.Description,
				Parameters:  d.InputSchema,
			},
		}
	}
	return out
}
