package ruleapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/web"
)

// templateVarRegex is pre-compiled regex for matching {{variable}} placeholders
// Compiled once at package initialization for performance
var templateVarRegex = regexp.MustCompile(`\{\{([^}]+)\}\}`)

// ============================================================
// Simulation Request Types
// ============================================================

// TestRuleRequest is the request body for POST /rules/{id}/test
type TestRuleRequest struct {
	SampleData map[string]interface{} `json:"sample_data"`
}

// Decode implements web.Decoder
func (r *TestRuleRequest) Decode(data []byte) error {
	return json.Unmarshal(data, r)
}

// ============================================================
// Simulation Response Types
// ============================================================

// SimulationResult is the response for POST /rules/{id}/test
type SimulationResult struct {
	RuleID            uuid.UUID               `json:"rule_id"`
	RuleName          string                  `json:"rule_name"`
	WouldTrigger      bool                    `json:"would_trigger"`
	MatchedConditions []ConditionResult       `json:"matched_conditions"`
	ActionsToExecute  []ActionPreview         `json:"actions_to_execute"`
	TemplatePreview   map[string]string       `json:"template_preview"`
	ValidationErrors  []string                `json:"validation_errors"`
}

// Encode implements web.Encoder
func (s SimulationResult) Encode() ([]byte, string, error) {
	data, err := json.Marshal(s)
	return data, "application/json", err
}

// ConditionResult shows whether a single condition matched during simulation
type ConditionResult struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Expected interface{} `json:"expected"`
	Actual   interface{} `json:"actual"`
	Matched  bool        `json:"matched"`
}

// ActionPreview shows what an action would do without executing it
type ActionPreview struct {
	ActionID   uuid.UUID         `json:"action_id"`
	ActionName string            `json:"action_name"`
	ActionType string            `json:"action_type"`
	Order      int               `json:"order"`
	Preview    map[string]string `json:"preview"`
}

// ============================================================
// Simulation Handler
// ============================================================

// testRule handles POST /v1/workflow/rules/{id}/test
func (a *api) testRule(ctx context.Context, r *http.Request) web.Encoder {
	ruleID, err := uuid.Parse(web.Param(r, "id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	var req TestRuleRequest
	if err := web.Decode(r, &req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	// Get rule
	rule, err := a.workflowBus.QueryRuleByID(ctx, ruleID)
	if err != nil {
		if errors.Is(err, workflow.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "query rule: %s", err)
	}

	// Get actions for preview
	actions, err := a.workflowBus.QueryActionsByRule(ctx, ruleID)
	if err != nil {
		return errs.Newf(errs.Internal, "query actions: %s", err)
	}

	result := SimulationResult{
		RuleID:            ruleID,
		RuleName:          rule.Name,
		WouldTrigger:      true,
		MatchedConditions: []ConditionResult{},
		ActionsToExecute:  []ActionPreview{},
		TemplatePreview:   make(map[string]string),
		ValidationErrors:  []string{},
	}

	// Evaluate trigger conditions
	if rule.TriggerConditions != nil && len(*rule.TriggerConditions) > 0 {
		conditionResults, allMatched := evaluateConditions(*rule.TriggerConditions, req.SampleData)
		result.MatchedConditions = conditionResults
		result.WouldTrigger = allMatched
	}

	// Generate action previews
	for _, action := range actions {
		if !action.IsActive {
			continue
		}

		preview := ActionPreview{
			ActionID:   action.ID,
			ActionName: action.Name,
			ActionType: extractActionType(action.ActionConfig),
			Order:      action.ExecutionOrder,
			Preview:    make(map[string]string),
		}

		// Render template variables in action config
		preview.Preview = renderTemplatePreview(action.ActionConfig, req.SampleData)

		result.ActionsToExecute = append(result.ActionsToExecute, preview)
	}

	// Generate overall template preview with sample variables
	result.TemplatePreview = generateTemplatePreview(req.SampleData)

	return result
}

// ============================================================
// Condition Evaluation Helpers
// ============================================================

// TriggerConditions represents the structure of trigger_conditions JSON
type TriggerConditions struct {
	Type       string              `json:"type"`       // "simple", "and", "or"
	Conditions []TriggerConditions `json:"conditions"` // For "and"/"or" types
	Field      string              `json:"field"`      // For "simple" type
	Operator   string              `json:"operator"`   // For "simple" type
	Value      interface{}         `json:"value"`      // For "simple" type
}

// evaluateConditions parses and evaluates trigger conditions against sample data
func evaluateConditions(conditionsJSON json.RawMessage, sampleData map[string]interface{}) ([]ConditionResult, bool) {
	var conditions TriggerConditions
	if err := json.Unmarshal(conditionsJSON, &conditions); err != nil {
		return []ConditionResult{}, true // Can't parse, assume would trigger
	}

	results := []ConditionResult{}
	allMatched := evaluateConditionNode(conditions, sampleData, &results)
	return results, allMatched
}

// evaluateConditionNode recursively evaluates a condition node
func evaluateConditionNode(cond TriggerConditions, data map[string]interface{}, results *[]ConditionResult) bool {
	switch cond.Type {
	case "simple", "":
		// Simple condition - evaluate directly
		actual := getFieldValue(data, cond.Field)
		matched := evaluateOperator(actual, cond.Operator, cond.Value)
		*results = append(*results, ConditionResult{
			Field:    cond.Field,
			Operator: cond.Operator,
			Expected: cond.Value,
			Actual:   actual,
			Matched:  matched,
		})
		return matched

	case "and":
		// All conditions must match
		allMatch := true
		for _, child := range cond.Conditions {
			if !evaluateConditionNode(child, data, results) {
				allMatch = false
			}
		}
		return allMatch

	case "or":
		// At least one condition must match
		anyMatch := false
		for _, child := range cond.Conditions {
			if evaluateConditionNode(child, data, results) {
				anyMatch = true
			}
		}
		return anyMatch

	default:
		// Unknown type, assume match
		return true
	}
}

// getFieldValue extracts a field value from nested data using dot notation
func getFieldValue(data map[string]interface{}, field string) interface{} {
	if data == nil {
		return nil
	}

	parts := strings.Split(field, ".")
	var current interface{} = data

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			current = v[part]
		default:
			return nil
		}
	}

	return current
}

// evaluateOperator evaluates a comparison operator
func evaluateOperator(actual interface{}, operator string, expected interface{}) bool {
	switch operator {
	case "equals", "eq", "=", "==":
		return compareValues(actual, expected) == 0
	case "not_equals", "neq", "!=", "<>":
		return compareValues(actual, expected) != 0
	case "greater_than", "gt", ">":
		return compareValues(actual, expected) > 0
	case "greater_than_or_equals", "gte", ">=":
		return compareValues(actual, expected) >= 0
	case "less_than", "lt", "<":
		return compareValues(actual, expected) < 0
	case "less_than_or_equals", "lte", "<=":
		return compareValues(actual, expected) <= 0
	case "contains":
		return containsValue(actual, expected)
	case "not_contains":
		return !containsValue(actual, expected)
	case "starts_with":
		return startsWithValue(actual, expected)
	case "ends_with":
		return endsWithValue(actual, expected)
	case "is_empty":
		return isEmpty(actual)
	case "is_not_empty":
		return !isEmpty(actual)
	case "in":
		return inList(actual, expected)
	case "not_in":
		return !inList(actual, expected)
	default:
		// Unknown operator, assume match
		return true
	}
}

// compareValues compares two values, returning -1, 0, or 1
func compareValues(a, b interface{}) int {
	if a == nil && b == nil {
		return 0
	}
	if a == nil {
		return -1
	}
	if b == nil {
		return 1
	}

	// Convert to strings for comparison
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)

	// Try numeric comparison first
	var aFloat, bFloat float64
	if _, err := fmt.Sscanf(aStr, "%f", &aFloat); err == nil {
		if _, err := fmt.Sscanf(bStr, "%f", &bFloat); err == nil {
			switch {
			case aFloat < bFloat:
				return -1
			case aFloat > bFloat:
				return 1
			default:
				return 0
			}
		}
	}

	// Fall back to string comparison
	if aStr < bStr {
		return -1
	}
	if aStr > bStr {
		return 1
	}
	return 0
}

// containsValue checks if a string contains another string
func containsValue(actual, expected interface{}) bool {
	aStr := fmt.Sprintf("%v", actual)
	eStr := fmt.Sprintf("%v", expected)
	return strings.Contains(strings.ToLower(aStr), strings.ToLower(eStr))
}

// startsWithValue checks if a string starts with another string
func startsWithValue(actual, expected interface{}) bool {
	aStr := fmt.Sprintf("%v", actual)
	eStr := fmt.Sprintf("%v", expected)
	return strings.HasPrefix(strings.ToLower(aStr), strings.ToLower(eStr))
}

// endsWithValue checks if a string ends with another string
func endsWithValue(actual, expected interface{}) bool {
	aStr := fmt.Sprintf("%v", actual)
	eStr := fmt.Sprintf("%v", expected)
	return strings.HasSuffix(strings.ToLower(aStr), strings.ToLower(eStr))
}

// isEmpty checks if a value is empty/nil/zero
func isEmpty(v interface{}) bool {
	if v == nil {
		return true
	}
	switch val := v.(type) {
	case string:
		return val == ""
	case []interface{}:
		return len(val) == 0
	case map[string]interface{}:
		return len(val) == 0
	default:
		return reflect.ValueOf(v).IsZero()
	}
}

// inList checks if a value is in a list
func inList(actual, expected interface{}) bool {
	// Expected should be a list
	list, ok := expected.([]interface{})
	if !ok {
		return false
	}
	for _, item := range list {
		if compareValues(actual, item) == 0 {
			return true
		}
	}
	return false
}

// ============================================================
// Template Rendering Helpers
// ============================================================

// extractActionType extracts the action_type from action config
func extractActionType(config json.RawMessage) string {
	var cfg map[string]interface{}
	if err := json.Unmarshal(config, &cfg); err != nil {
		return "unknown"
	}
	if actionType, ok := cfg["action_type"].(string); ok {
		return actionType
	}
	if actionType, ok := cfg["type"].(string); ok {
		return actionType
	}
	return "unknown"
}

// renderTemplatePreview renders template variables in action config fields
func renderTemplatePreview(config json.RawMessage, sampleData map[string]interface{}) map[string]string {
	result := make(map[string]string)

	var cfg map[string]interface{}
	if err := json.Unmarshal(config, &cfg); err != nil {
		return result
	}

	// Find string fields that might contain template variables
	for key, value := range cfg {
		if strVal, ok := value.(string); ok {
			// Check if it contains template variables ({{...}})
			if strings.Contains(strVal, "{{") {
				rendered := renderTemplateString(strVal, sampleData)
				if rendered != strVal {
					result[key] = rendered
				}
			}
		}
	}

	return result
}

// renderTemplateString replaces {{variable}} placeholders with values from sample data
func renderTemplateString(template string, data map[string]interface{}) string {
	return templateVarRegex.ReplaceAllStringFunc(template, func(match string) string {
		// Extract variable name (remove {{ and }})
		varName := strings.TrimPrefix(strings.TrimSuffix(match, "}}"), "{{")
		varName = strings.TrimSpace(varName)

		// Handle filters (e.g., {{total | currency:USD}})
		parts := strings.Split(varName, "|")
		fieldPath := strings.TrimSpace(parts[0])

		// Get value from sample data
		value := getFieldValue(data, fieldPath)
		if value == nil {
			return match // Keep original if not found
		}

		return fmt.Sprintf("%v", value)
	})
}

// generateTemplatePreview creates a preview of available template variables
func generateTemplatePreview(sampleData map[string]interface{}) map[string]string {
	preview := make(map[string]string)
	flattenData("entity", sampleData, preview)
	return preview
}

// flattenData flattens nested data into dot-notation keys
func flattenData(prefix string, data map[string]interface{}, result map[string]string) {
	for key, value := range data {
		fullKey := prefix + "." + key
		switch v := value.(type) {
		case map[string]interface{}:
			flattenData(fullKey, v, result)
		default:
			result[fullKey] = fmt.Sprintf("%v", v)
		}
	}
}
