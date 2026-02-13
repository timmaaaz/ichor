// Package llm defines a vendor-neutral interface for large language model
// providers. Implementations live in sub-packages (claude/, mock/, etc.).
package llm

import (
	"context"
	"encoding/json"
)

// Provider abstracts LLM vendors so the chat endpoint can swap
// Claude → GPT-4 → local models without touching the API layer.
type Provider interface {
	// StreamChat sends a request and returns a channel of streaming events.
	// The channel is closed when the response is complete or an error occurs.
	// The caller must drain the channel.
	StreamChat(ctx context.Context, req ChatRequest) (<-chan StreamEvent, error)
}

// ChatRequest is a single conversation turn sent to the provider.
type ChatRequest struct {
	SystemPrompt string
	Messages     []Message
	Tools        []ToolDef
	MaxTokens    int
}

// Message is one message in the conversation history.
type Message struct {
	Role    string // "user" | "assistant"
	Content string // text content (may be empty for tool-result messages)

	// ToolCalls is populated for assistant messages that requested tools.
	ToolCalls []ToolCall

	// ToolResults is populated for user messages that supply tool outputs.
	ToolResults []ToolResult
}

// ToolDef describes a callable tool exposed to the LLM.
type ToolDef struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"` // JSON Schema
}

// ToolCall represents the LLM requesting a tool invocation.
type ToolCall struct {
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

// ToolResult carries the output of one executed tool call.
type ToolResult struct {
	ToolUseID string `json:"tool_use_id"`
	Content   string `json:"content"`
	IsError   bool   `json:"is_error"`
}

// EventType discriminates StreamEvent variants.
type EventType string

const (
	EventMessageStart    EventType = "message_start"
	EventContentDelta    EventType = "content_delta"
	EventToolUseStart    EventType = "tool_use_start"
	EventToolUseInput    EventType = "tool_use_input"
	EventMessageComplete EventType = "message_complete"
	EventError           EventType = "error"
)

// StreamEvent is a single event emitted by the provider during streaming.
type StreamEvent struct {
	Type EventType

	// Text delta (EventContentDelta).
	Text string

	// Tool use metadata (EventToolUseStart).
	ToolCallID   string
	ToolCallName string

	// Partial JSON input for the current tool (EventToolUseInput).
	PartialInput string

	// True when the LLM's turn ended with a tool_use stop reason,
	// meaning the caller should execute tools and continue the loop.
	StopForToolUse bool

	// Non-nil for EventError.
	Err error
}
