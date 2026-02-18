package temporal

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

func Test_computeFieldChanges(t *testing.T) {
	tests := []struct {
		name   string
		before map[string]any
		after  map[string]any
		want   map[string]workflow.FieldChange
	}{
		{
			name:   "no changes",
			before: map[string]any{"name": "Alice", "age": float64(30)},
			after:  map[string]any{"name": "Alice", "age": float64(30)},
			want:   nil,
		},
		{
			name:   "single field changed",
			before: map[string]any{"name": "Alice", "age": float64(30)},
			after:  map[string]any{"name": "Bob", "age": float64(30)},
			want: map[string]workflow.FieldChange{
				"name": {OldValue: "Alice", NewValue: "Bob"},
			},
		},
		{
			name:   "multiple fields changed",
			before: map[string]any{"name": "Alice", "age": float64(30), "city": "NYC"},
			after:  map[string]any{"name": "Bob", "age": float64(31), "city": "NYC"},
			want: map[string]workflow.FieldChange{
				"name": {OldValue: "Alice", NewValue: "Bob"},
				"age":  {OldValue: float64(30), NewValue: float64(31)},
			},
		},
		{
			name:   "new field in after",
			before: map[string]any{"name": "Alice"},
			after:  map[string]any{"name": "Alice", "age": float64(30)},
			want: map[string]workflow.FieldChange{
				"age": {OldValue: nil, NewValue: float64(30)},
			},
		},
		{
			name:   "nil to value",
			before: map[string]any{"name": "Alice", "notes": nil},
			after:  map[string]any{"name": "Alice", "notes": "hello"},
			want: map[string]workflow.FieldChange{
				"notes": {OldValue: nil, NewValue: "hello"},
			},
		},
		{
			name:   "value to nil",
			before: map[string]any{"name": "Alice", "notes": "hello"},
			after:  map[string]any{"name": "Alice", "notes": nil},
			want: map[string]workflow.FieldChange{
				"notes": {OldValue: "hello", NewValue: nil},
			},
		},
		{
			name:   "empty maps",
			before: map[string]any{},
			after:  map[string]any{},
			want:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeFieldChanges(tt.before, tt.after)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_extractEntityData(t *testing.T) {
	t.Run("nil result", func(t *testing.T) {
		_, _, err := extractEntityData(nil)
		require.Error(t, err)
	})

	t.Run("struct with ID field", func(t *testing.T) {
		type entity struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		e := entity{ID: "550e8400-e29b-41d4-a716-446655440000", Name: "test"}
		id, data, err := extractEntityData(e)
		require.NoError(t, err)
		require.Equal(t, "550e8400-e29b-41d4-a716-446655440000", id.String())
		require.Equal(t, "test", data["name"])
	})

	t.Run("map input", func(t *testing.T) {
		m := map[string]any{"id": "550e8400-e29b-41d4-a716-446655440000", "status": "active"}
		id, data, err := extractEntityData(m)
		require.NoError(t, err)
		require.Equal(t, "550e8400-e29b-41d4-a716-446655440000", id.String())
		require.Equal(t, "active", data["status"])
	})
}
