package pageconfigbus

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// extractJSONPath converts a struct field path to a JSON path using reflection.
// Example: "PageConfigWithRelations.Contents[2].TableConfigID" → "contents[2].tableConfigId"
func extractJSONPath(structPath string, structType reflect.Type) string {
	// Handle simple root-level fields first
	if !strings.Contains(structPath, ".") {
		return structPath
	}

	// Split path and remove the root struct name
	parts := strings.Split(structPath, ".")
	if len(parts) > 1 && parts[0] == structType.Name() {
		parts = parts[1:] // Remove root struct name
	}

	var jsonPath strings.Builder
	currentType := structType

	for i, part := range parts {
		// Handle array indices (e.g., "Contents[2]" → "contents[2]")
		if idx := strings.Index(part, "["); idx != -1 {
			fieldName := part[:idx]
			arrayPart := part[idx:] // "[2]"

			// Get JSON name for the field
			if currentType.Kind() == reflect.Ptr {
				currentType = currentType.Elem()
			}

			if currentType.Kind() == reflect.Struct {
				if field, found := currentType.FieldByName(fieldName); found {
					jsonTag := field.Tag.Get("json")
					if jsonTag != "" {
						jsonName := strings.Split(jsonTag, ",")[0]
						if jsonName != "-" {
							if jsonPath.Len() > 0 {
								jsonPath.WriteString(".")
							}
							jsonPath.WriteString(jsonName)
							jsonPath.WriteString(arrayPart)

							// Update current type to array element type
							fieldType := field.Type
							if fieldType.Kind() == reflect.Slice || fieldType.Kind() == reflect.Array {
								currentType = fieldType.Elem()
							}
						}
					}
				}
			}
			continue
		}

		// Handle regular field names
		if currentType.Kind() == reflect.Ptr {
			currentType = currentType.Elem()
		}

		if currentType.Kind() == reflect.Struct {
			field, found := currentType.FieldByName(part)
			if found {
				jsonTag := field.Tag.Get("json")
				if jsonTag != "" {
					jsonName := strings.Split(jsonTag, ",")[0]
					if jsonName != "-" {
						if jsonPath.Len() > 0 && i > 0 {
							jsonPath.WriteString(".")
						}
						jsonPath.WriteString(jsonName)
						currentType = field.Type
					}
				} else {
					// No JSON tag, use lowercase field name
					if jsonPath.Len() > 0 && i > 0 {
						jsonPath.WriteString(".")
					}
					jsonPath.WriteString(strings.ToLower(part[:1]) + part[1:])
				}
			}
		}
	}

	result := jsonPath.String()
	if result == "" {
		// Fallback: return the last part of the path in camelCase
		lastPart := parts[len(parts)-1]
		return strings.ToLower(lastPart[:1]) + lastPart[1:]
	}

	return result
}

// transformFieldError converts a validator.FieldError to a ValidationError with a helpful message
func transformFieldError(err validator.FieldError, structType reflect.Type) ValidationError {
	path := extractJSONPath(err.Namespace(), structType)

	var message string
	switch err.Tag() {
	case "required":
		message = fmt.Sprintf("Field '%s' is required and cannot be empty", err.Field())
	case "uuid":
		message = fmt.Sprintf("Field '%s' must be a valid UUID (format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx)", err.Field())
	case "oneof":
		message = fmt.Sprintf("Field '%s' must be one of: %s", err.Field(), err.Param())
	case "min":
		message = fmt.Sprintf("Field '%s' must be at least %s", err.Field(), err.Param())
	case "max":
		message = fmt.Sprintf("Field '%s' must be at most %s", err.Field(), err.Param())
	case "validContentType":
		message = fmt.Sprintf("Content type '%v' is invalid. Valid types: table, form, chart, tabs, container, text", err.Value())
	default:
		message = fmt.Sprintf("Validation failed for field '%s': %s", err.Field(), err.Tag())
	}

	return ValidationError{
		Field:   path,
		Message: message,
		Code:    mapValidatorTagToErrorCode(err.Tag()),
	}
}

// mapValidatorTagToErrorCode maps validator tags to our error codes
func mapValidatorTagToErrorCode(tag string) string {
	switch tag {
	case "required":
		return ErrCodeRequiredField
	case "uuid":
		return "INVALID_FORMAT"
	case "oneof", "validContentType":
		return ErrCodeInvalidType
	case "min", "max":
		return "RANGE_ERROR"
	default:
		return "INVALID_FORMAT"
	}
}

// sanitizeValue prevents exposing sensitive data in error messages.
// It limits string lengths and redacts potential secrets.
func sanitizeValue(value any) any {
	if value == nil {
		return nil
	}

	// Handle strings
	if str, ok := value.(string); ok {
		// Limit length to prevent log spam
		if len(str) > 100 {
			return str[:97] + "..."
		}

		// Redact potential secrets (tokens, long strings without spaces)
		if len(str) > 20 && !strings.Contains(str, " ") {
			return "[REDACTED]"
		}

		return str
	}

	// Handle numbers (don't sanitize)
	switch value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return value
	case bool:
		return value
	}

	// For complex types, return type description
	return fmt.Sprintf("<%T>", value)
}

// validateSingleContent validates a single content item based on its type
func validateSingleContent(ctx context.Context, content PageContentExport, index int, basePath string) []ValidationError {
	var errors []ValidationError
	itemPath := fmt.Sprintf("%s[%d]", basePath, index)

	// Check for context cancellation
	select {
	case <-ctx.Done():
		return []ValidationError{{
			Field:   "validation",
			Message: "Validation cancelled or timed out",
			Code:    "VALIDATION_TIMEOUT",
		}}
	default:
	}

	// Validate content type is recognized
	validTypes := []string{"table", "form", "chart", "tabs", "container", "text"}
	if !contains(validTypes, content.ContentType) {
		errors = append(errors, ValidationError{
			Field:   fmt.Sprintf("%s.contentType", itemPath),
			Message: fmt.Sprintf("Invalid content type '%s'. Valid types: %s", content.ContentType, strings.Join(validTypes, ", ")),
			Code:    ErrCodeInvalidType,
		})
		return errors // Don't continue if type is invalid
	}

	// Type-specific validation
	switch content.ContentType {
	case "table":
		if content.TableConfigID == uuid.Nil {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("%s.tableConfigId", itemPath),
				Message: "Table content requires a valid tableConfigId",
				Code:    ErrCodeRequiredField,
			})
		}

	case "form":
		if content.FormID == uuid.Nil {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("%s.formId", itemPath),
				Message: "Form content requires a valid formId",
				Code:    ErrCodeRequiredField,
			})
		}

	case "chart":
		if content.TableConfigID == uuid.Nil {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("%s.tableConfigId", itemPath),
				Message: "Chart content requires a valid tableConfigId for data source",
				Code:    ErrCodeRequiredField,
			})
		}

	case "tabs":
		// Tabs should ideally have children, but we allow empty for progressive construction
		if content.Label == "" {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("%s.label", itemPath),
				Message: "Tabs content should have a descriptive label",
				Code:    ErrCodeRequiredField,
			})
		}

	case "container":
		// Container can be empty, but if it has layout, validate it
		if len(content.Layout) > 0 {
			errors = append(errors, validateLayoutBytes(content.Layout, itemPath)...)
		}

	case "text":
		// Text content should have label or some content
		if content.Label == "" {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("%s.label", itemPath),
				Message: "Text content should have a label",
				Code:    ErrCodeRequiredField,
			})
		}
	}

	// Validate layout if present
	if len(content.Layout) > 0 {
		errors = append(errors, validateLayoutBytes(content.Layout, itemPath)...)
	}

	return errors
}

// LayoutConfig represents the layout configuration structure
type LayoutConfig struct {
	ColSpan *ResponsiveValue `json:"colSpan,omitempty"`
	RowSpan *int             `json:"rowSpan,omitempty"`
	Gap     *string          `json:"gap,omitempty"`
}

// ResponsiveValue represents responsive breakpoint values
type ResponsiveValue struct {
	Xs *int `json:"xs,omitempty"`
	Sm *int `json:"sm,omitempty"`
	Md *int `json:"md,omitempty"`
	Lg *int `json:"lg,omitempty"`
	Xl *int `json:"xl,omitempty"`
}

// validateLayoutBytes parses layout JSON and validates it
func validateLayoutBytes(layoutBytes []byte, basePath string) []ValidationError {
	if len(layoutBytes) == 0 {
		return nil
	}

	var layout LayoutConfig
	if err := json.Unmarshal(layoutBytes, &layout); err != nil {
		return []ValidationError{{
			Field:   fmt.Sprintf("%s.layout", basePath),
			Message: fmt.Sprintf("Invalid layout JSON: %v", err),
			Code:    ErrCodeInvalidJSON,
		}}
	}

	return validateLayout(layout, basePath)
}

// validateLayout validates layout configuration (colSpan ranges, valid breakpoints)
func validateLayout(layout LayoutConfig, basePath string) []ValidationError {
	var errors []ValidationError

	if layout.ColSpan != nil {
		// Validate responsive breakpoint values (1-12 for colSpan)
		breakpoints := map[string]*int{
			"xs": layout.ColSpan.Xs,
			"sm": layout.ColSpan.Sm,
			"md": layout.ColSpan.Md,
			"lg": layout.ColSpan.Lg,
			"xl": layout.ColSpan.Xl,
		}

		for bp, val := range breakpoints {
			if val != nil {
				if *val < 1 || *val > 12 {
					errors = append(errors, ValidationError{
						Field:   fmt.Sprintf("%s.layout.colSpan.%s", basePath, bp),
						Message: fmt.Sprintf("Column span must be between 1 and 12 (got %d)", *val),
						Code:    "RANGE_ERROR",
					})
				}
			}
		}
	}

	// Validate rowSpan if present
	if layout.RowSpan != nil && *layout.RowSpan < 1 {
		errors = append(errors, ValidationError{
			Field:   fmt.Sprintf("%s.layout.rowSpan", basePath),
			Message: fmt.Sprintf("Row span must be at least 1 (got %d)", *layout.RowSpan),
			Code:    "RANGE_ERROR",
		})
	}

	return errors
}

// calculateNestingDepth recursively calculates the maximum nesting depth of content tree
func calculateNestingDepth(contents []PageContentExport, currentDepth int) int {
	if len(contents) == 0 {
		return currentDepth
	}

	maxDepth := currentDepth
	for _, content := range contents {
		// For now, contents don't have children in PageContentExport
		// This is a placeholder for future nested content support
		_ = content
	}

	return maxDepth
}

// detectCircularReferences checks for circular parent-child relationships
func detectCircularReferences(contents []PageContentExport) []ValidationError {
	// Build adjacency list (parent -> children map)
	graph := make(map[uuid.UUID][]uuid.UUID)
	idSet := make(map[uuid.UUID]bool)

	for _, content := range contents {
		if content.ID != uuid.Nil {
			idSet[content.ID] = true
			if content.ParentID != uuid.Nil {
				graph[content.ParentID] = append(graph[content.ParentID], content.ID)
			}
		}
	}

	// Perform DFS to detect cycles
	visited := make(map[uuid.UUID]bool)
	recursionStack := make(map[uuid.UUID]bool)

	var detectCycle func(nodeID uuid.UUID) []uuid.UUID
	detectCycle = func(nodeID uuid.UUID) []uuid.UUID {
		visited[nodeID] = true
		recursionStack[nodeID] = true

		for _, childID := range graph[nodeID] {
			if !visited[childID] {
				if cycle := detectCycle(childID); cycle != nil {
					return append([]uuid.UUID{nodeID}, cycle...)
				}
			} else if recursionStack[childID] {
				// Found a cycle
				return []uuid.UUID{nodeID, childID}
			}
		}

		recursionStack[nodeID] = false
		return nil
	}

	// Check all nodes for cycles
	for id := range idSet {
		if !visited[id] {
			if cycle := detectCycle(id); cycle != nil {
				cycleStr := make([]string, len(cycle))
				for i, id := range cycle {
					cycleStr[i] = id.String()
				}
				return []ValidationError{{
					Field:   "contents",
					Message: fmt.Sprintf("Circular reference detected: %s", strings.Join(cycleStr, " -> ")),
					Code:    ErrCodeCircularReference,
				}}
			}
		}
	}

	return nil
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
