package workflow

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// TemplateProcessor handles variable substitution in templates
type TemplateProcessor struct {
	strictMode    bool                  // If true, fail on missing variables
	defaultValue  string                // Default value for missing variables
	preserveTypes bool                  // Keep non-string types intact
	allowNested   bool                  // Allow nested object access like user.profile.name
	customFilters map[string]FilterFunc // Custom filter functions
	variableRegex *regexp.Regexp
}

// FilterFunc defines a function that can transform values in templates
type FilterFunc func(value interface{}, args ...string) (interface{}, error)

// TemplateContext provides the data context for template processing
type TemplateContext map[string]interface{}

// TemplateVariable represents a variable found and processed in a template
type TemplateVariable struct {
	Name     string      `json:"name"`
	Path     string      `json:"path"`
	Value    interface{} `json:"value"`
	Source   string      `json:"source"` // "context", "computed", "default"
	RawMatch string      `json:"raw_match"`
}

// TemplateProcessingResult contains the result of template processing
type TemplateProcessingResult struct {
	Processed     interface{}        `json:"processed"`
	VariablesUsed []TemplateVariable `json:"variables_used"`
	Warnings      []string           `json:"warnings"`
	Errors        []string           `json:"errors"`
}

// TemplateProcessingOptions configures template processing behavior
type TemplateProcessingOptions struct {
	StrictMode    bool                  `json:"strict_mode"`
	DefaultValue  string                `json:"default_value"`
	PreserveTypes bool                  `json:"preserve_types"`
	AllowNested   bool                  `json:"allow_nested"`
	CustomFilters map[string]FilterFunc `json:"-"`
}

// DefaultTemplateProcessingOptions returns default options for template processing
func DefaultTemplateProcessingOptions() TemplateProcessingOptions {
	return TemplateProcessingOptions{
		StrictMode:    false,
		DefaultValue:  "",
		PreserveTypes: true,
		AllowNested:   true,
		CustomFilters: make(map[string]FilterFunc),
	}
}

// NewTemplateProcessor creates a new template processor with the given options
func NewTemplateProcessor(opts TemplateProcessingOptions) *TemplateProcessor {
	return &TemplateProcessor{
		strictMode:    opts.StrictMode,
		defaultValue:  opts.DefaultValue,
		preserveTypes: opts.PreserveTypes,
		allowNested:   opts.AllowNested,
		customFilters: opts.CustomFilters,
		variableRegex: regexp.MustCompile(`\{\{([^}]+)\}\}`),
	}
}

// ProcessTemplate processes template variables in a string
func (tp *TemplateProcessor) ProcessTemplate(template string, context TemplateContext) TemplateProcessingResult {
	result := TemplateProcessingResult{
		VariablesUsed: make([]TemplateVariable, 0),
		Warnings:      make([]string, 0),
		Errors:        make([]string, 0),
	}

	processed := template
	matches := tp.variableRegex.FindAllStringSubmatch(template, -1)

	for _, match := range matches {
		fullMatch := match[0]                       // "{{variable_name}}"
		variablePath := strings.TrimSpace(match[1]) // "variable_name"

		// Validate syntax
		if err := tp.validateVariable(variablePath); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Invalid variable syntax: %s - %v", variablePath, err))
			continue
		}

		// Resolve variable value
		resolution := tp.resolve(variablePath, context)

		if resolution.Found {
			// Convert value to string for replacement
			valueStr := tp.valueToString(resolution.Value)
			processed = strings.ReplaceAll(processed, fullMatch, valueStr)

			result.VariablesUsed = append(result.VariablesUsed, TemplateVariable{
				Name:     variablePath,
				Path:     resolution.Path,
				Value:    resolution.Value,
				Source:   resolution.Source,
				RawMatch: fullMatch,
			})
		} else {
			// Handle missing variable
			if tp.strictMode {
				result.Errors = append(result.Errors, fmt.Sprintf("Missing variable: %s", variablePath))
			} else {
				result.Warnings = append(result.Warnings, fmt.Sprintf("Missing variable: %s, using default", variablePath))
				processed = strings.ReplaceAll(processed, fullMatch, tp.defaultValue)
			}
		}
	}

	result.Processed = processed
	return result
}

// ProcessTemplateObject processes template variables in any object (recursive)
func (tp *TemplateProcessor) ProcessTemplateObject(obj interface{}, context TemplateContext) TemplateProcessingResult {
	result := TemplateProcessingResult{
		VariablesUsed: make([]TemplateVariable, 0),
		Warnings:      make([]string, 0),
		Errors:        make([]string, 0),
	}

	processed := tp.processValue(obj, context, &result)
	result.Processed = processed
	return result
}

// processValue recursively processes a value
func (tp *TemplateProcessor) processValue(value interface{}, context TemplateContext, result *TemplateProcessingResult) interface{} {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case string:
		// Process string templates
		subResult := tp.ProcessTemplate(v, context)
		result.VariablesUsed = append(result.VariablesUsed, subResult.VariablesUsed...)
		result.Warnings = append(result.Warnings, subResult.Warnings...)
		result.Errors = append(result.Errors, subResult.Errors...)
		return subResult.Processed

	case map[string]interface{}:
		// Process maps recursively
		processed := make(map[string]interface{})
		for key, val := range v {
			processed[key] = tp.processValue(val, context, result)
		}
		return processed

	case []interface{}:
		// Process slices recursively
		processed := make([]interface{}, len(v))
		for i, val := range v {
			processed[i] = tp.processValue(val, context, result)
		}
		return processed

	case json.RawMessage:
		// Handle JSON raw messages
		var decoded interface{}
		if err := json.Unmarshal(v, &decoded); err == nil {
			return tp.processValue(decoded, context, result)
		}
		return v

	default:
		// Use reflection for other slice types
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Slice {
			length := rv.Len()
			processed := make([]interface{}, length)
			for i := 0; i < length; i++ {
				processed[i] = tp.processValue(rv.Index(i).Interface(), context, result)
			}
			return processed
		}

		// Return primitive values unchanged
		return value
	}
}

// ResolutionResult contains the result of resolving a variable
type ResolutionResult struct {
	Found  bool
	Value  interface{}
	Path   string
	Source string // "context", "computed", "default"
}

// resolve resolves a variable path with optional filters
func (tp *TemplateProcessor) resolve(variablePath string, context TemplateContext) ResolutionResult {
	// Parse variable path and filters: "user.name | uppercase"
	path, filters := tp.parseVariablePath(variablePath)

	// Resolve the base value
	value := tp.resolveNestedPath(path, context)
	found := value != nil
	source := "context"

	if !found {
		value = tp.defaultValue
		source = "default"
	}

	// Apply filters if any
	if found && len(filters) > 0 {
		var err error
		value, err = tp.applyFilters(value, filters)
		if err == nil {
			source = "computed"
		}
		// If filter application fails, keep original value
	}

	return ResolutionResult{
		Found:  found,
		Value:  value,
		Path:   path,
		Source: source,
	}
}

// parseVariablePath parses a variable path into base path and filters
func (tp *TemplateProcessor) parseVariablePath(variablePath string) (string, []filterSpec) {
	parts := strings.Split(variablePath, "|")
	path := strings.TrimSpace(parts[0])

	filters := make([]filterSpec, 0, len(parts)-1)
	for i := 1; i < len(parts); i++ {
		filterStr := strings.TrimSpace(parts[i])
		filterParts := strings.Split(filterStr, ":")

		filter := filterSpec{
			name: strings.TrimSpace(filterParts[0]),
			args: make([]string, 0),
		}

		for j := 1; j < len(filterParts); j++ {
			filter.args = append(filter.args, strings.TrimSpace(filterParts[j]))
		}

		filters = append(filters, filter)
	}

	return path, filters
}

type filterSpec struct {
	name string
	args []string
}

// resolveNestedPath resolves a nested path in the context
func (tp *TemplateProcessor) resolveNestedPath(path string, context TemplateContext) interface{} {
	if !tp.allowNested {
		return context[path]
	}

	segments := strings.Split(path, ".")
	var current interface{} = context

	for _, segment := range segments {
		switch v := current.(type) {
		case map[string]interface{}:
			val, ok := v[segment]
			if !ok {
				return nil
			}
			current = val
		case TemplateContext:
			val, ok := v[segment]
			if !ok {
				return nil
			}
			current = val
		default:
			// Try reflection for struct fields
			rv := reflect.ValueOf(current)
			if rv.Kind() == reflect.Ptr {
				rv = rv.Elem()
			}
			if rv.Kind() != reflect.Struct {
				return nil
			}

			field := rv.FieldByName(segment)
			if !field.IsValid() {
				return nil
			}
			current = field.Interface()
		}
	}

	return current
}

// applyFilters applies a series of filters to a value
func (tp *TemplateProcessor) applyFilters(value interface{}, filters []filterSpec) (interface{}, error) {
	current := value

	for _, filter := range filters {
		// Check built-in filters first
		if fn, ok := builtInFilters[filter.name]; ok {
			var err error
			current, err = fn(current, filter.args...)
			if err != nil {
				return value, fmt.Errorf("filter %s failed: %w", filter.name, err)
			}
		} else if fn, ok := tp.customFilters[filter.name]; ok {
			// Then check custom filters
			var err error
			current, err = fn(current, filter.args...)
			if err != nil {
				return value, fmt.Errorf("filter %s failed: %w", filter.name, err)
			}
		} else {
			return value, fmt.Errorf("unknown filter: %s", filter.name)
		}
	}

	return current, nil
}

// validateVariable validates a variable path syntax
func (tp *TemplateProcessor) validateVariable(variablePath string) error {
	if variablePath == "" {
		return fmt.Errorf("empty variable name")
	}

	// Check for invalid characters
	invalidChars := regexp.MustCompile(`[^a-zA-Z0-9._|:\s]`)
	if invalidChars.MatchString(variablePath) {
		return fmt.Errorf("invalid characters in variable")
	}

	// Parse and validate variable and filters
	parts := strings.Split(variablePath, "|")
	variable := strings.TrimSpace(parts[0])

	// Validate variable name
	validVar := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9._]*$`)
	if !validVar.MatchString(variable) {
		return fmt.Errorf("invalid variable name: %s", variable)
	}

	// Validate filters
	for i := 1; i < len(parts); i++ {
		filterParts := strings.Split(strings.TrimSpace(parts[i]), ":")
		filterName := strings.TrimSpace(filterParts[0])

		validFilter := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)
		if !validFilter.MatchString(filterName) {
			return fmt.Errorf("invalid filter name: %s", filterName)
		}
	}

	return nil
}

// valueToString converts a value to string representation
func (tp *TemplateProcessor) valueToString(value interface{}) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	case bool:
		return strconv.FormatBool(v)
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%g", v)
	case time.Time:
		return v.Format(time.RFC3339)
	case []byte:
		return string(v)
	default:
		// Try JSON encoding for complex types
		if data, err := json.Marshal(value); err == nil {
			return string(data)
		}
		return fmt.Sprintf("%v", value)
	}
}

// builtInFilters contains all built-in filter functions
var builtInFilters = map[string]FilterFunc{
	// String filters
	"uppercase": func(value interface{}, args ...string) (interface{}, error) {
		return strings.ToUpper(fmt.Sprintf("%v", value)), nil
	},
	"lowercase": func(value interface{}, args ...string) (interface{}, error) {
		return strings.ToLower(fmt.Sprintf("%v", value)), nil
	},
	"capitalize": func(value interface{}, args ...string) (interface{}, error) {
		str := fmt.Sprintf("%v", value)
		if len(str) == 0 {
			return str, nil
		}
		return strings.ToUpper(str[:1]) + strings.ToLower(str[1:]), nil
	},
	"trim": func(value interface{}, args ...string) (interface{}, error) {
		return strings.TrimSpace(fmt.Sprintf("%v", value)), nil
	},
	"truncate": func(value interface{}, args ...string) (interface{}, error) {
		str := fmt.Sprintf("%v", value)
		length := 50
		if len(args) > 0 {
			if l, err := strconv.Atoi(args[0]); err == nil {
				length = l
			}
		}
		if len(str) > length {
			return str[:length] + "...", nil
		}
		return str, nil
	},

	// Number filters
	"currency": func(value interface{}, args ...string) (interface{}, error) {
		// Simple currency formatting - could be enhanced
		num, err := toFloat64(value)
		if err != nil {
			return value, err
		}
		currency := "USD"
		if len(args) > 0 {
			currency = args[0]
		}
		symbol := "$"
		switch currency {
		case "EUR":
			symbol = "€"
		case "GBP":
			symbol = "£"
		case "JPY":
			symbol = "¥"
		}
		return fmt.Sprintf("%s%.2f", symbol, num), nil
	},
	"round": func(value interface{}, args ...string) (interface{}, error) {
		num, err := toFloat64(value)
		if err != nil {
			return value, err
		}
		precision := 0
		if len(args) > 0 {
			if p, err := strconv.Atoi(args[0]); err == nil {
				precision = p
			}
		}
		multiplier := 1.0
		for i := 0; i < precision; i++ {
			multiplier *= 10
		}
		return float64(int(num*multiplier+0.5)) / multiplier, nil
	},

	// Date filters
	"formatDate": func(value interface{}, args ...string) (interface{}, error) {
		var t time.Time
		var err error

		switch v := value.(type) {
		case time.Time:
			t = v
		case string:
			t, err = time.Parse(time.RFC3339, v)
			if err != nil {
				// Try other common formats
				for _, format := range []string{
					"2006-01-02",
					"2006-01-02 15:04:05",
					time.RFC822,
					time.RFC850,
					time.RFC1123,
				} {
					if t, err = time.Parse(format, v); err == nil {
						break
					}
				}
			}
		default:
			return value, fmt.Errorf("cannot format non-date value")
		}

		if err != nil {
			return value, err
		}

		format := "2006-01-02"
		if len(args) > 0 {
			switch args[0] {
			case "short":
				format = "Jan 2, 2006"
			case "long":
				format = "Monday, January 2, 2006"
			case "time":
				format = "15:04:05"
			case "datetime":
				format = "2006-01-02 15:04:05"
			default:
				format = args[0]
			}
		}

		return t.Format(format), nil
	},

	// Array/slice filters
	"join": func(value interface{}, args ...string) (interface{}, error) {
		separator := ", "
		if len(args) > 0 {
			separator = args[0]
		}

		rv := reflect.ValueOf(value)
		if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
			return fmt.Sprintf("%v", value), nil
		}

		parts := make([]string, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			parts[i] = fmt.Sprintf("%v", rv.Index(i).Interface())
		}

		return strings.Join(parts, separator), nil
	},
	"first": func(value interface{}, args ...string) (interface{}, error) {
		rv := reflect.ValueOf(value)
		if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
			return value, nil
		}
		if rv.Len() == 0 {
			return nil, nil
		}
		return rv.Index(0).Interface(), nil
	},
	"last": func(value interface{}, args ...string) (interface{}, error) {
		rv := reflect.ValueOf(value)
		if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
			return value, nil
		}
		if rv.Len() == 0 {
			return nil, nil
		}
		return rv.Index(rv.Len() - 1).Interface(), nil
	},

	// Utility filters
	"default": func(value interface{}, args ...string) (interface{}, error) {
		if value == nil || value == "" {
			if len(args) > 0 {
				return args[0], nil
			}
			return "", nil
		}
		return value, nil
	},
}

// toFloat64 converts various numeric types to float64
func toFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", value)
	}
}
