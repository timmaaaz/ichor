package chatapi

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestIsNewWorkflow(t *testing.T) {
	tests := []struct {
		name string
		ctx  json.RawMessage
		want bool
	}{
		{
			name: "nil context",
			ctx:  nil,
			want: true,
		},
		{
			name: "empty bytes",
			ctx:  json.RawMessage{},
			want: true,
		},
		{
			name: "null JSON",
			ctx:  json.RawMessage(`null`),
			want: true,
		},
		{
			name: "explicit is_new true",
			ctx:  json.RawMessage(`{"is_new":true}`),
			want: true,
		},
		{
			name: "explicit is_new true with empty workflow_id",
			ctx:  json.RawMessage(`{"is_new":true,"workflow_id":"","nodes":[]}`),
			want: true,
		},
		{
			name: "empty workflow_id and no nodes",
			ctx:  json.RawMessage(`{"workflow_id":"","nodes":[]}`),
			want: true,
		},
		{
			name: "empty workflow_id and null nodes",
			ctx:  json.RawMessage(`{"workflow_id":""}`),
			want: true,
		},
		{
			name: "has workflow_id",
			ctx:  json.RawMessage(`{"workflow_id":"550e8400-e29b-41d4-a716-446655440000","nodes":[]}`),
			want: false,
		},
		{
			name: "has nodes",
			ctx:  json.RawMessage(`{"workflow_id":"","nodes":[{"data":{"id":"abc"}}]}`),
			want: false,
		},
		{
			name: "has both workflow_id and nodes",
			ctx:  json.RawMessage(`{"workflow_id":"550e8400-e29b-41d4-a716-446655440000","nodes":[{"data":{}}]}`),
			want: false,
		},
		{
			name: "is_new false with workflow_id",
			ctx:  json.RawMessage(`{"is_new":false,"workflow_id":"550e8400-e29b-41d4-a716-446655440000"}`),
			want: false,
		},
		{
			name: "is_new false with no workflow_id — explicit false wins over inference",
			ctx:  json.RawMessage(`{"is_new":false}`),
			want: false,
		},
		{
			name: "is_new false with empty nodes — explicit false wins over inference",
			ctx:  json.RawMessage(`{"is_new":false,"workflow_id":"","nodes":[]}`),
			want: false,
		},
		{
			name: "invalid JSON",
			ctx:  json.RawMessage(`{not valid`),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isNewWorkflow(tt.ctx)
			if got != tt.want {
				t.Errorf("isNewWorkflow(%s) = %v, want %v", string(tt.ctx), got, tt.want)
			}
		})
	}
}

func TestBuildSystemPrompt_GuidedCreation(t *testing.T) {
	t.Run("new workflow includes guided prompt", func(t *testing.T) {
		ctx := json.RawMessage(`{"is_new":true}`)
		prompt := buildSystemPrompt("workflow", ctx)

		if !strings.Contains(prompt, "Guided Workflow Creation") {
			t.Error("expected guided creation prompt in system prompt for new workflow")
		}
		// Should still include the standard role block.
		if !strings.Contains(prompt, "workflow automation assistant") {
			t.Error("expected standard role block to be preserved")
		}
		// Should still include draft builder guidance.
		if !strings.Contains(prompt, "Draft Builder") {
			t.Error("expected draft builder guidance to be preserved")
		}
	})

	t.Run("existing workflow excludes guided prompt", func(t *testing.T) {
		ctx := json.RawMessage(`{"workflow_id":"550e8400-e29b-41d4-a716-446655440000","nodes":[{"data":{}}]}`)
		prompt := buildSystemPrompt("workflow", ctx)

		if strings.Contains(prompt, "Guided Workflow Creation") {
			t.Error("did not expect guided creation prompt for existing workflow")
		}
		if !strings.Contains(prompt, "workflow automation assistant") {
			t.Error("expected standard role block")
		}
	})

	t.Run("tables context never gets guided prompt", func(t *testing.T) {
		ctx := json.RawMessage(`{}`)
		prompt := buildSystemPrompt("tables", ctx)

		if strings.Contains(prompt, "Guided Workflow Creation") {
			t.Error("did not expect guided creation prompt for tables context")
		}
		if !strings.Contains(prompt, "UI configuration assistant") {
			t.Error("expected tables role block")
		}
	})

	t.Run("nil context gets guided prompt", func(t *testing.T) {
		prompt := buildSystemPrompt("workflow", nil)

		if !strings.Contains(prompt, "Guided Workflow Creation") {
			t.Error("expected guided creation prompt when no context provided")
		}
	})

	t.Run("new workflow does not include context preamble", func(t *testing.T) {
		ctx := json.RawMessage(`{"is_new":true}`)
		prompt := buildSystemPrompt("workflow", ctx)

		if strings.Contains(prompt, "complete workflow state is provided below") {
			t.Error("context preamble should not appear for new workflows")
		}
	})

	t.Run("existing workflow includes context preamble", func(t *testing.T) {
		ctx := json.RawMessage(`{"workflow_id":"550e8400-e29b-41d4-a716-446655440000","nodes":[{"data":{}}]}`)
		prompt := buildSystemPrompt("workflow", ctx)

		if !strings.Contains(prompt, "complete workflow state is provided below") {
			t.Error("expected context preamble for existing workflow")
		}
	})
}
