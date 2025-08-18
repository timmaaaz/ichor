package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/sdk/workflow/stores/workflowdb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// FieldCondition represents a condition for field evaluation
type FieldCondition struct {
	FieldName     string      `json:"field_name"`
	Operator      string      `json:"operator"`
	Value         interface{} `json:"value,omitempty"`
	PreviousValue interface{} `json:"previous_value,omitempty"`
}

// TriggerConditions represents the conditions for triggering a rule
type TriggerConditions struct {
	FieldConditions []FieldCondition `json:"field_conditions,omitempty"`
}

// ConditionEvaluationResult represents the result of evaluating a condition
type ConditionEvaluationResult struct {
	Condition     FieldCondition `json:"condition"`
	Matched       bool           `json:"matched"`
	ActualValue   interface{}    `json:"actual_value,omitempty"`
	PreviousValue interface{}    `json:"previous_value,omitempty"`
	Error         string         `json:"error,omitempty"`
}

// RuleMatchResult represents the result of matching a rule against an event
type RuleMatchResult struct {
	Rule             workflowdb.AutomationRulesView `json:"rule"`
	Matched          bool                           `json:"matched"`
	TriggerEvent     TriggerEvent                   `json:"trigger_event"`
	ConditionResults []ConditionEvaluationResult    `json:"condition_results"`
	MatchReason      string                         `json:"match_reason,omitempty"`
	ExecutionContext map[string]interface{}         `json:"execution_context"`
}

// ProcessingResult represents the result of processing a trigger event
type ProcessingResult struct {
	TriggerEvent        TriggerEvent      `json:"trigger_event"`
	TotalRulesEvaluated int               `json:"total_rules_evaluated"`
	MatchedRules        []RuleMatchResult `json:"matched_rules"`
	ProcessingTime      time.Duration     `json:"processing_time_ms"`
	Errors              []string          `json:"errors"`
}

// EventValidationResult represents the result of event validation
type EventValidationResult struct {
	IsValid  bool     `json:"is_valid"`
	Errors   []string `json:"errors"`
	Warnings []string `json:"warnings"`
}

// TriggerProcessor processes trigger events and determines which rules should execute
type TriggerProcessor struct {
	log *logger.Logger
	db  *sqlx.DB

	// Cached data
	activeRules  []workflowdb.AutomationRulesView
	lastLoadTime time.Time
	cacheTimeout time.Duration
}

// NewTriggerProcessor creates a new trigger processor
func NewTriggerProcessor(log *logger.Logger, db *sqlx.DB) *TriggerProcessor {
	return &TriggerProcessor{
		log:          log,
		db:           db,
		cacheTimeout: 5 * time.Minute,
	}
}

// Initialize loads metadata and prepares the processor
func (tp *TriggerProcessor) Initialize(ctx context.Context) error {
	tp.log.Info(ctx, "Initializing trigger processor...")

	if err := tp.loadMetadata(ctx); err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}

	tp.log.Info(ctx, "Trigger processor initialized successfully")
	return nil
}

// loadMetadata loads active rules and related metadata
func (tp *TriggerProcessor) loadMetadata(ctx context.Context) error {
	// Check if cache is still valid
	if time.Since(tp.lastLoadTime) < tp.cacheTimeout && len(tp.activeRules) > 0 {
		return nil
	}

	query := `
        SELECT 
            ar.id,
            ar.name,
            ar.description,
            ar.entity_name,
            ar.trigger_conditions,
            ar.is_active,
            ar.created_date,
            ar.updated_date,
            ar.created_by,
            ar.updated_by,
            tt.id as trigger_type_id,
            tt.name as trigger_type_name,
            tt.description as trigger_type_description,
            et.id as entity_type_id,
            et.name as entity_type_name,
            et.description as entity_type_description,
            e.name as entity_name,
            e.id as entity_id
        FROM automation_rules ar
        LEFT JOIN trigger_types tt ON ar.trigger_type_id = tt.id
        LEFT JOIN entities e ON ar.entity_name = e.name
        LEFT JOIN entity_types et ON e.entity_type_id = et.id
        WHERE ar.is_active = true
    `

	var rules []workflowdb.AutomationRulesView
	if err := tp.db.SelectContext(ctx, &rules, query); err != nil {
		return fmt.Errorf("failed to load active rules: %w", err)
	}

	tp.activeRules = rules
	tp.lastLoadTime = time.Now()

	tp.log.Info(ctx, "Loaded active rules", "count", len(rules))
	return nil
}

// ProcessEvent processes a trigger event and returns matching rules
func (tp *TriggerProcessor) ProcessEvent(ctx context.Context, event TriggerEvent) (*ProcessingResult, error) {
	startTime := time.Now()

	// Reload metadata if cache expired
	if err := tp.loadMetadata(ctx); err != nil {
		return nil, fmt.Errorf("failed to refresh metadata: %w", err)
	}

	// Validate event
	validation := tp.validateEvent(event)
	if !validation.IsValid {
		return nil, fmt.Errorf("invalid trigger event: %s", strings.Join(validation.Errors, ", "))
	}

	// Get rules for this entity
	entityRules := tp.getRulesForEntity(event.EntityName)

	// Evaluate each rule
	matchResults := make([]RuleMatchResult, 0)
	errors := make([]string, 0)

	for _, rule := range entityRules {
		matchResult := tp.checkRuleMatch(rule, event)
		if matchResult.Matched {
			matchResults = append(matchResults, matchResult)
		}
	}

	return &ProcessingResult{
		TriggerEvent:        event,
		TotalRulesEvaluated: len(entityRules),
		MatchedRules:        matchResults,
		ProcessingTime:      time.Since(startTime),
		Errors:              errors,
	}, nil
}

// validateEvent validates a trigger event
func (tp *TriggerProcessor) validateEvent(event TriggerEvent) EventValidationResult {
	errors := make([]string, 0)
	warnings := make([]string, 0)

	// Basic validation
	if event.EventType == "" {
		errors = append(errors, "Event type is required")
	} else if !tp.isSupportedEventType(event.EventType) {
		errors = append(errors, fmt.Sprintf("Unsupported event type: %s", event.EventType))
	}

	if event.EntityName == "" {
		errors = append(errors, "Entity name is required")
	}

	if event.Timestamp.IsZero() {
		errors = append(errors, "Timestamp is required")
	}

	// Update-specific validation
	if event.EventType == "on_update" {
		if len(event.FieldChanges) == 0 {
			warnings = append(warnings, "Update event has no field changes specified")
		}
	}

	// Entity existence check
	if event.EntityName != "" && !tp.hasRulesForEntity(event.EntityName) {
		warnings = append(warnings, fmt.Sprintf("No active rules found for entity: %s", event.EntityName))
	}

	return EventValidationResult{
		IsValid:  len(errors) == 0,
		Errors:   errors,
		Warnings: warnings,
	}
}

// checkRuleMatch checks if a rule matches the given event
func (tp *TriggerProcessor) checkRuleMatch(rule workflowdb.AutomationRulesView, event TriggerEvent) RuleMatchResult {
	result := RuleMatchResult{
		Rule:             rule,
		Matched:          false,
		TriggerEvent:     event,
		ConditionResults: make([]ConditionEvaluationResult, 0),
		ExecutionContext: make(map[string]interface{}),
	}

	// Check if rule applies to this entity
	if rule.EntityName.Valid && rule.EntityName.String != event.EntityName {
		result.MatchReason = fmt.Sprintf("Entity mismatch: rule for %s, event for %s",
			rule.EntityName.String, event.EntityName)
		return result
	}

	// Check if rule applies to this trigger type
	if rule.TriggerTypeName.Valid {
		expectedTriggerType := rule.TriggerTypeName.String
		if expectedTriggerType != event.EventType {
			result.MatchReason = fmt.Sprintf("Trigger type mismatch: rule for %s, event for %s",
				expectedTriggerType, event.EventType)
			return result
		}
	}

	// Evaluate field conditions
	result.ConditionResults = tp.evaluateRuleConditions(rule, event)

	// Rule matches if all conditions pass (AND logic)
	hasConditions := len(result.ConditionResults) > 0
	allConditionsPass := true

	if hasConditions {
		for _, cr := range result.ConditionResults {
			if !cr.Matched || cr.Error != "" {
				allConditionsPass = false
				break
			}
		}
	}

	result.Matched = allConditionsPass

	if result.Matched {
		if hasConditions {
			result.MatchReason = "All field conditions satisfied"
		} else {
			result.MatchReason = "No conditions specified (auto-match)"
		}

		// Build execution context
		result.ExecutionContext = map[string]interface{}{
			"entity_id":     event.EntityID,
			"entity_name":   event.EntityName,
			"event_type":    event.EventType,
			"field_changes": event.FieldChanges,
			"raw_data":      event.RawData,
			"timestamp":     event.Timestamp,
		}
	} else {
		failedConditions := make([]string, 0)
		for _, cr := range result.ConditionResults {
			if !cr.Matched || cr.Error != "" {
				failedConditions = append(failedConditions, cr.Condition.FieldName)
			}
		}
		result.MatchReason = fmt.Sprintf("Failed conditions: %s", strings.Join(failedConditions, ", "))
	}

	return result
}

// evaluateRuleConditions evaluates all conditions for a rule
func (tp *TriggerProcessor) evaluateRuleConditions(rule workflowdb.AutomationRulesView, event TriggerEvent) []ConditionEvaluationResult {
	// TODO: Write validation function for this
	// if !rule.TriggerConditions.Valid {
	// 	return []ConditionEvaluationResult{}
	// }

	var conditions TriggerConditions
	if err := json.Unmarshal(rule.TriggerConditions, &conditions); err != nil {
		// TODO: Check on use of context.Background here
		tp.log.Error(context.Background(), "Failed to unmarshal trigger conditions",
			"rule", rule.ID,
			"error", err)
		return []ConditionEvaluationResult{}
	}

	if len(conditions.FieldConditions) == 0 {
		return []ConditionEvaluationResult{}
	}

	results := make([]ConditionEvaluationResult, 0, len(conditions.FieldConditions))
	for _, condition := range conditions.FieldConditions {
		result := tp.evaluateFieldCondition(condition, event)
		results = append(results, result)
	}

	return results
}

// evaluateFieldCondition evaluates a single field condition
func (tp *TriggerProcessor) evaluateFieldCondition(condition FieldCondition, event TriggerEvent) ConditionEvaluationResult {
	result := ConditionEvaluationResult{
		Condition: condition,
		Matched:   false,
	}

	fieldName := condition.FieldName
	var currentValue, previousValue interface{}

	// Get current and previous values
	if event.EventType == "on_update" && event.FieldChanges != nil {
		if fieldChange, exists := event.FieldChanges[fieldName]; exists {
			currentValue = fieldChange.NewValue
			previousValue = fieldChange.OldValue
		} else {
			// Field wasn't changed in this update
			if event.RawData != nil {
				currentValue = event.RawData[fieldName]
			}
			previousValue = currentValue
		}
	} else {
		// For create/delete, use raw data
		if event.RawData != nil {
			currentValue = event.RawData[fieldName]
		}
		previousValue = nil
	}

	result.ActualValue = currentValue
	result.PreviousValue = previousValue

	// Evaluate condition based on operator
	switch condition.Operator {
	case "equals":
		result.Matched = tp.compareValues(currentValue, condition.Value, "==")

	case "not_equals":
		result.Matched = !tp.compareValues(currentValue, condition.Value, "==")

	case "changed_from":
		result.Matched = event.EventType == "on_update" &&
			tp.compareValues(previousValue, condition.PreviousValue, "==")

	case "changed_to":
		result.Matched = event.EventType == "on_update" &&
			tp.compareValues(currentValue, condition.Value, "==") &&
			!tp.compareValues(previousValue, condition.Value, "==")

	case "greater_than":
		result.Matched = tp.compareValues(currentValue, condition.Value, ">")

	case "less_than":
		result.Matched = tp.compareValues(currentValue, condition.Value, "<")

	case "contains":
		if strVal, ok := currentValue.(string); ok {
			if searchVal, ok := condition.Value.(string); ok {
				result.Matched = strings.Contains(strVal, searchVal)
			}
		}

	case "in":
		if values, ok := condition.Value.([]interface{}); ok {
			for _, v := range values {
				if tp.compareValues(currentValue, v, "==") {
					result.Matched = true
					break
				}
			}
		}

	default:
		result.Error = fmt.Sprintf("Unknown operator: %s", condition.Operator)
		result.Matched = false
	}

	return result
}

// compareValues compares two values based on the operator
func (tp *TriggerProcessor) compareValues(a, b interface{}, op string) bool {
	// Handle nil cases
	if a == nil || b == nil {
		if op == "==" {
			return a == b
		}
		return false
	}

	// Try to convert to comparable types
	switch op {
	case "==":
		return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
	case ">":
		// Try numeric comparison
		aFloat, aOk := tp.toFloat64(a)
		bFloat, bOk := tp.toFloat64(b)
		if aOk && bOk {
			return aFloat > bFloat
		}
		// Fall back to string comparison
		return fmt.Sprintf("%v", a) > fmt.Sprintf("%v", b)
	case "<":
		aFloat, aOk := tp.toFloat64(a)
		bFloat, bOk := tp.toFloat64(b)
		if aOk && bOk {
			return aFloat < bFloat
		}
		return fmt.Sprintf("%v", a) < fmt.Sprintf("%v", b)
	default:
		return false
	}
}

// toFloat64 attempts to convert a value to float64
func (tp *TriggerProcessor) toFloat64(v interface{}) (float64, bool) {
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

// Helper methods

func (tp *TriggerProcessor) isSupportedEventType(eventType string) bool {
	supportedTypes := []string{"on_create", "on_update", "on_delete", "scheduled"}
	for _, t := range supportedTypes {
		if t == eventType {
			return true
		}
	}
	return false
}

func (tp *TriggerProcessor) hasRulesForEntity(entityName string) bool {
	for _, rule := range tp.activeRules {
		if rule.EntityName.Valid && rule.EntityName.String == entityName {
			return true
		}
	}
	return false
}

func (tp *TriggerProcessor) getRulesForEntity(entityName string) []workflowdb.AutomationRulesView {
	rules := make([]workflowdb.AutomationRulesView, 0)
	for _, rule := range tp.activeRules {
		if rule.EntityName.Valid && rule.EntityName.String == entityName && rule.IsActive {
			rules = append(rules, rule)
		}
	}
	return rules
}

// GetMatchedRulesForEntity returns rules that would match for a given entity and event type
func (tp *TriggerProcessor) GetMatchedRulesForEntity(entityName string, eventType string) []workflowdb.AutomationRulesView {
	matched := make([]workflowdb.AutomationRulesView, 0)
	expectedTriggerType := eventType

	for _, rule := range tp.activeRules {
		if rule.EntityName.Valid && rule.EntityName.String == entityName {
			if rule.TriggerTypeName.Valid && rule.TriggerTypeName.String == expectedTriggerType {
				matched = append(matched, rule)
			}
		}
	}

	return matched
}

// RefreshRules forces a reload of rules from the database
func (tp *TriggerProcessor) RefreshRules(ctx context.Context) error {
	tp.lastLoadTime = time.Time{} // Reset cache time to force reload
	return tp.loadMetadata(ctx)
}
