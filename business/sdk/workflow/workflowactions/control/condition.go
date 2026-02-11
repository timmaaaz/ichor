// Package control provides control flow actions for workflow execution.
package control

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// FieldCondition represents a condition for field evaluation.
// This mirrors the structure in trigger.go for consistency.
type FieldCondition struct {
	FieldName     string      `json:"field_name"`
	Operator      string      `json:"operator"`
	Value         interface{} `json:"value,omitempty"`
	PreviousValue interface{} `json:"previous_value,omitempty"`
}

// ConditionConfig defines the configuration for evaluate_condition action.
type ConditionConfig struct {
	Conditions []FieldCondition `json:"conditions"`
	LogicType  string           `json:"logic_type"` // "and" (default) or "or"
}

// EvaluateConditionHandler evaluates conditions and returns branch direction.
// This is used for implementing condition nodes in workflow graphs.
type EvaluateConditionHandler struct {
	log *logger.Logger
}

// NewEvaluateConditionHandler creates a new condition evaluation handler.
func NewEvaluateConditionHandler(log *logger.Logger) *EvaluateConditionHandler {
	return &EvaluateConditionHandler{log: log}
}

// GetType returns the action type this handler supports.
func (h *EvaluateConditionHandler) GetType() string {
	return "evaluate_condition"
}

// GetDescription returns a human-readable description for discovery APIs.
func (h *EvaluateConditionHandler) GetDescription() string {
	return "Evaluates conditions against entity data and determines branch direction for workflow execution"
}

// SupportsManualExecution returns false - conditions only make sense in workflow context.
func (h *EvaluateConditionHandler) SupportsManualExecution() bool {
	return false
}

// IsAsync returns false - condition evaluation completes inline.
func (h *EvaluateConditionHandler) IsAsync() bool {
	return false
}

// Validate validates the action configuration before execution.
func (h *EvaluateConditionHandler) Validate(config json.RawMessage) error {
	var cfg ConditionConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("invalid condition config: %w", err)
	}

	if len(cfg.Conditions) == 0 {
		return fmt.Errorf("at least one condition is required")
	}

	// Validate logic type if provided
	if cfg.LogicType != "" && cfg.LogicType != "and" && cfg.LogicType != "or" {
		return fmt.Errorf("invalid logic_type: must be 'and' or 'or'")
	}

	// Validate each condition has required fields
	validOperators := map[string]bool{
		"equals":       true,
		"not_equals":   true,
		"changed_from": true,
		"changed_to":   true,
		"greater_than": true,
		"less_than":    true,
		"contains":     true,
		"in":           true,
		"is_null":      true,
		"is_not_null":  true,
	}

	for i, cond := range cfg.Conditions {
		if cond.FieldName == "" {
			return fmt.Errorf("condition %d: field_name is required", i)
		}
		if cond.Operator == "" {
			return fmt.Errorf("condition %d: operator is required", i)
		}
		if !validOperators[cond.Operator] {
			return fmt.Errorf("condition %d: invalid operator '%s'", i, cond.Operator)
		}
	}

	return nil
}

// GetOutputPorts implements workflow.OutputPortProvider.
func (h *EvaluateConditionHandler) GetOutputPorts() []workflow.OutputPort {
	return []workflow.OutputPort{
		{Name: "true", Description: "Condition evaluated to true", IsDefault: true},
		{Name: "false", Description: "Condition evaluated to false"},
	}
}

// Execute evaluates conditions against the execution context and returns the branch direction.
func (h *EvaluateConditionHandler) Execute(ctx context.Context, config json.RawMessage, execCtx workflow.ActionExecutionContext) (interface{}, error) {
	var cfg ConditionConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse condition config: %w", err)
	}

	// Default to AND logic
	logicType := cfg.LogicType
	if logicType == "" {
		logicType = "and"
	}

	// Evaluate conditions against RawData and FieldChanges
	result := h.evaluateConditions(cfg.Conditions, execCtx.RawData, execCtx.FieldChanges, execCtx.EventType, logicType)

	output := "false"
	if result {
		output = "true"
	}

	h.log.Info(ctx, "evaluate_condition action executed",
		"conditions_count", len(cfg.Conditions),
		"logic_type", logicType,
		"result", result,
		"output", output,
		"entity_name", execCtx.EntityName,
		"entity_id", execCtx.EntityID)

	return map[string]any{
		"evaluated": true,
		"result":    result,
		"output":    output,
	}, nil
}

// evaluateConditions evaluates all conditions and returns the combined result.
func (h *EvaluateConditionHandler) evaluateConditions(
	conditions []FieldCondition,
	data map[string]interface{},
	fieldChanges map[string]workflow.FieldChange,
	eventType string,
	logicType string,
) bool {
	if len(conditions) == 0 {
		return true // No conditions = always true
	}

	for _, cond := range conditions {
		matched := h.evaluateSingleCondition(cond, data, fieldChanges, eventType)

		if logicType == "or" && matched {
			return true // OR: any true means true
		}
		if logicType == "and" && !matched {
			return false // AND: any false means false
		}
	}

	// AND: all true means true, OR: all false means false
	return logicType == "and"
}

// evaluateSingleCondition evaluates a single field condition.
func (h *EvaluateConditionHandler) evaluateSingleCondition(
	cond FieldCondition,
	data map[string]interface{},
	fieldChanges map[string]workflow.FieldChange,
	eventType string,
) bool {
	fieldName := cond.FieldName
	var currentValue, previousValue interface{}

	// Get current and previous values
	if eventType == "on_update" && fieldChanges != nil {
		if fieldChange, exists := fieldChanges[fieldName]; exists {
			currentValue = fieldChange.NewValue
			previousValue = fieldChange.OldValue
		} else {
			// Field wasn't changed in this update
			if data != nil {
				currentValue = data[fieldName]
			}
			previousValue = currentValue
		}
	} else {
		// For create/delete/manual, use raw data
		if data != nil {
			currentValue = data[fieldName]
		}
		previousValue = nil
	}

	// Evaluate condition based on operator
	switch cond.Operator {
	case "equals":
		return h.compareValues(currentValue, cond.Value, "==")

	case "not_equals":
		return !h.compareValues(currentValue, cond.Value, "==")

	case "changed_from":
		return eventType == "on_update" &&
			h.compareValues(previousValue, cond.PreviousValue, "==")

	case "changed_to":
		return eventType == "on_update" &&
			h.compareValues(currentValue, cond.Value, "==") &&
			!h.compareValues(previousValue, cond.Value, "==")

	case "greater_than":
		return h.compareValues(currentValue, cond.Value, ">")

	case "less_than":
		return h.compareValues(currentValue, cond.Value, "<")

	case "contains":
		if strVal, ok := currentValue.(string); ok {
			if searchVal, ok := cond.Value.(string); ok {
				return strings.Contains(strVal, searchVal)
			}
		}
		return false

	case "in":
		if values, ok := cond.Value.([]interface{}); ok {
			for _, v := range values {
				if h.compareValues(currentValue, v, "==") {
					return true
				}
			}
		}
		return false

	case "is_null":
		return currentValue == nil

	case "is_not_null":
		return currentValue != nil

	default:
		return false
	}
}

// compareValues compares two values based on the operator.
func (h *EvaluateConditionHandler) compareValues(a, b interface{}, op string) bool {
	// Handle nil cases
	if a == nil || b == nil {
		if op == "==" {
			return a == b
		}
		return false
	}

	switch op {
	case "==":
		return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
	case ">":
		aFloat, aOk := h.toFloat64(a)
		bFloat, bOk := h.toFloat64(b)
		if aOk && bOk {
			return aFloat > bFloat
		}
		// Fall back to string comparison
		return fmt.Sprintf("%v", a) > fmt.Sprintf("%v", b)
	case "<":
		aFloat, aOk := h.toFloat64(a)
		bFloat, bOk := h.toFloat64(b)
		if aOk && bOk {
			return aFloat < bFloat
		}
		return fmt.Sprintf("%v", a) < fmt.Sprintf("%v", b)
	default:
		return false
	}
}

// toFloat64 attempts to convert a value to float64.
func (h *EvaluateConditionHandler) toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int32:
		return float64(val), true
	case int64:
		return float64(val), true
	case json.Number:
		f, err := val.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}
