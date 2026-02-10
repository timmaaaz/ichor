package temporal

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// =============================================================================
// Boundary Tests: String (exactly at limit vs just over)
// =============================================================================

func TestPayload_StringExactlyAtLimit(t *testing.T) {
	// MaxResultValueSize bytes: should NOT truncate (uses > not >=).
	result := map[string]any{
		"data": strings.Repeat("x", MaxResultValueSize),
	}
	sanitized, truncated := sanitizeResult(result)
	require.False(t, truncated, "exactly at limit should NOT truncate")
	require.Equal(t, result["data"], sanitized["data"])

	_, hasTruncatedFlag := sanitized["data_truncated"]
	require.False(t, hasTruncatedFlag)
}

func TestPayload_StringOneOverLimit(t *testing.T) {
	// MaxResultValueSize+1 bytes: should truncate.
	result := map[string]any{
		"data": strings.Repeat("x", MaxResultValueSize+1),
	}
	sanitized, truncated := sanitizeResult(result)
	require.True(t, truncated)

	truncatedStr := sanitized["data"].(string)
	require.True(t, strings.HasSuffix(truncatedStr, "...[TRUNCATED]"))
	require.Equal(t, MaxResultValueSize+len("...[TRUNCATED]"), len(truncatedStr))

	require.True(t, sanitized["data_truncated"].(bool))
}

func TestPayload_StringWellOverLimit(t *testing.T) {
	// 10x over limit: verify truncated string size is bounded.
	result := map[string]any{
		"data": strings.Repeat("x", MaxResultValueSize*10),
	}
	sanitized, truncated := sanitizeResult(result)
	require.True(t, truncated)

	truncatedStr := sanitized["data"].(string)
	maxExpected := MaxResultValueSize + len("...[TRUNCATED]")
	require.LessOrEqual(t, len(truncatedStr), maxExpected,
		"truncated string should be bounded to MaxResultValueSize + marker")
}

// =============================================================================
// Boundary Tests: Binary Data
// =============================================================================

func TestPayload_BinaryExactlyAtLimit(t *testing.T) {
	result := map[string]any{
		"data": make([]byte, MaxResultValueSize),
	}
	sanitized, truncated := sanitizeResult(result)
	require.False(t, truncated, "binary exactly at limit should NOT truncate")

	_, hasTruncatedFlag := sanitized["data_truncated"]
	require.False(t, hasTruncatedFlag)
}

func TestPayload_BinaryOneOverLimit(t *testing.T) {
	result := map[string]any{
		"data": make([]byte, MaxResultValueSize+1),
	}
	sanitized, truncated := sanitizeResult(result)
	require.True(t, truncated)
	require.Equal(t, "[BINARY_DATA_TRUNCATED]", sanitized["data"])
	require.True(t, sanitized["data_truncated"].(bool))
}

// =============================================================================
// Boundary Tests: Complex Objects (JSON serialization size)
// =============================================================================

func TestPayload_ObjectUnderLimit(t *testing.T) {
	// Build an object whose JSON serialization is under MaxResultValueSize.
	smallMap := make(map[string]any)
	for i := 0; i < 490; i++ {
		smallMap[fmt.Sprintf("k%03d", i)] = strings.Repeat("v", 90)
	}
	result := map[string]any{"data": smallMap}
	sanitized, truncated := sanitizeResult(result)
	require.False(t, truncated, "object under limit should NOT truncate")

	_, isMap := sanitized["data"].(map[string]any)
	require.True(t, isMap, "object should be preserved as map")
}

func TestPayload_ObjectOverLimit(t *testing.T) {
	// Build an object whose JSON serialization exceeds MaxResultValueSize.
	largeMap := make(map[string]any)
	for i := 0; i < 1000; i++ {
		largeMap[fmt.Sprintf("key_%04d", i)] = strings.Repeat("v", 100)
	}
	result := map[string]any{"data": largeMap}
	sanitized, truncated := sanitizeResult(result)
	require.True(t, truncated)
	require.Equal(t, "[LARGE_OBJECT_TRUNCATED]", sanitized["data"])
	require.True(t, sanitized["data_truncated"].(bool))
}

// =============================================================================
// Multiple Large Fields
// =============================================================================

func TestPayload_MultipleLargeFields(t *testing.T) {
	// Two fields both over limit — both should be independently truncated.
	result := map[string]any{
		"field1": strings.Repeat("a", MaxResultValueSize+1),
		"field2": strings.Repeat("b", MaxResultValueSize+1),
	}
	sanitized, truncated := sanitizeResult(result)
	require.True(t, truncated)
	require.True(t, sanitized["field1_truncated"].(bool))
	require.True(t, sanitized["field2_truncated"].(bool))

	require.Contains(t, sanitized["field1"].(string), "...[TRUNCATED]")
	require.Contains(t, sanitized["field2"].(string), "...[TRUNCATED]")
}

// =============================================================================
// MergeResult Integration with Truncation
// =============================================================================

func TestPayload_MergeResult_LargeValueTruncated(t *testing.T) {
	ctx := NewMergedContext(map[string]any{"trigger": "data"})
	largeResult := map[string]any{
		"data": strings.Repeat("x", MaxResultValueSize+1),
	}
	ctx.MergeResult("test_action", largeResult)

	storedResult := ctx.ActionResults["test_action"]
	require.Contains(t, storedResult["data"].(string), "...[TRUNCATED]")
	require.True(t, storedResult["data_truncated"].(bool))

	// Verify truncation flag propagated to Flattened map.
	require.True(t, ctx.Flattened["test_action.data_truncated"].(bool))
}

func TestPayload_MergeResult_SmallValuePreserved(t *testing.T) {
	ctx := NewMergedContext(map[string]any{"trigger": "data"})
	smallResult := map[string]any{
		"data": strings.Repeat("x", MaxResultValueSize), // Exactly at limit
	}
	ctx.MergeResult("test_action", smallResult)

	storedResult := ctx.ActionResults["test_action"]
	require.Equal(t, smallResult["data"], storedResult["data"])

	_, hasTruncatedFlag := storedResult["data_truncated"]
	require.False(t, hasTruncatedFlag)
}

// =============================================================================
// MergedContext Size Stress Test
// =============================================================================

func TestPayload_MergedContext_LargeContextStillFunctions(t *testing.T) {
	// Add many results to approach the 200KB warning threshold.
	ctx := NewMergedContext(map[string]any{"trigger": "data"})

	// Add 20 actions with ~9KB each ≈ 180KB total.
	for i := 0; i < 20; i++ {
		ctx.MergeResult(
			fmt.Sprintf("action_%d", i),
			map[string]any{"data": strings.Repeat("x", 9*1024)},
		)
	}

	require.Equal(t, 20, len(ctx.ActionResults))
	require.Equal(t, "data", ctx.Flattened["trigger"])

	require.NotEmpty(t, ctx.ActionResults["action_0"]["data"])
	require.NotEmpty(t, ctx.ActionResults["action_19"]["data"])

	require.NotNil(t, ctx.Flattened["action_0.data"])
	require.NotNil(t, ctx.Flattened["action_19.data"])
}
