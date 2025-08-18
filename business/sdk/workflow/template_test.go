package workflow_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

func TestTemplateProcessor_SimpleString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		template string
		context  workflow.TemplateContext
		opts     workflow.TemplateProcessingOptions
		want     string
		wantVars int
		wantErr  bool
	}{
		{
			name:     "basic variable substitution",
			template: "Hello {{name}}, welcome to {{company}}!",
			context: workflow.TemplateContext{
				"name":    "John",
				"company": "Acme Corp",
			},
			opts:     workflow.DefaultTemplateProcessingOptions(),
			want:     "Hello John, welcome to Acme Corp!",
			wantVars: 2,
		},
		{
			name:     "missing variable with default",
			template: "Hello {{name}}, your ID is {{id}}",
			context: workflow.TemplateContext{
				"name": "Jane",
			},
			opts: workflow.TemplateProcessingOptions{
				StrictMode:   false,
				DefaultValue: "N/A",
			},
			want:     "Hello Jane, your ID is N/A",
			wantVars: 1,
		},
		{
			name:     "missing variable strict mode",
			template: "Hello {{name}}, your ID is {{id}}",
			context: workflow.TemplateContext{
				"name": "Jane",
			},
			opts: workflow.TemplateProcessingOptions{
				StrictMode: true,
			},
			wantErr: true,
		},
		{
			name:     "nested path access",
			template: "User {{user.profile.name}} from {{user.profile.city}}",
			context: workflow.TemplateContext{
				"user": map[string]any{
					"profile": map[string]any{
						"name": "Bob",
						"city": "New York",
					},
				},
			},
			opts: workflow.TemplateProcessingOptions{
				AllowNested: true,
			},
			want:     "User Bob from New York",
			wantVars: 2,
		},
		{
			name:     "nested path disabled",
			template: "User {{user.profile.name}}",
			context: workflow.TemplateContext{
				"user": map[string]any{
					"profile": map[string]any{
						"name": "Bob",
					},
				},
			},
			opts: workflow.TemplateProcessingOptions{
				AllowNested:  false,
				DefaultValue: "unknown",
			},
			want:     "User unknown",
			wantVars: 0,
		},
		{
			name:     "no variables",
			template: "This is a plain text message",
			context:  workflow.TemplateContext{},
			opts:     workflow.DefaultTemplateProcessingOptions(),
			want:     "This is a plain text message",
			wantVars: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := workflow.NewTemplateProcessor(tt.opts)
			result := processor.ProcessTemplate(tt.template, tt.context)

			if tt.wantErr {
				if len(result.Errors) == 0 {
					t.Errorf("expected errors but got none")
				}
				return
			}

			if len(result.Errors) > 0 {
				t.Errorf("unexpected errors: %v", result.Errors)
			}

			got := result.Processed.(string)
			if got != tt.want {
				t.Errorf("ProcessTemplate() = %q, want %q", got, tt.want)
			}

			if len(result.VariablesUsed) != tt.wantVars {
				t.Errorf("variables used = %d, want %d", len(result.VariablesUsed), tt.wantVars)
			}
		})
	}
}

func TestTemplateProcessor_Filters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		template string
		context  workflow.TemplateContext
		want     string
	}{
		{
			name:     "uppercase filter",
			template: "Hello {{name | uppercase}}!",
			context: workflow.TemplateContext{
				"name": "john",
			},
			want: "Hello JOHN!",
		},
		{
			name:     "lowercase filter",
			template: "Hello {{name | lowercase}}!",
			context: workflow.TemplateContext{
				"name": "JANE",
			},
			want: "Hello jane!",
		},
		{
			name:     "capitalize filter",
			template: "Hello {{name | capitalize}}!",
			context: workflow.TemplateContext{
				"name": "alice smith",
			},
			want: "Hello Alice smith!",
		},
		{
			name:     "truncate filter",
			template: "{{message | truncate:10}}",
			context: workflow.TemplateContext{
				"message": "This is a very long message that needs truncation",
			},
			want: "This is a ...",
		},
		{
			name:     "currency filter USD",
			template: "Total: {{amount | currency}}",
			context: workflow.TemplateContext{
				"amount": 123.45,
			},
			want: "Total: $123.45",
		},
		{
			name:     "currency filter EUR",
			template: "Total: {{amount | currency:EUR}}",
			context: workflow.TemplateContext{
				"amount": 123.45,
			},
			want: "Total: €123.45",
		},
		{
			name:     "round filter",
			template: "Value: {{value | round:2}}",
			context: workflow.TemplateContext{
				"value": 3.14159,
			},
			want: "Value: 3.14",
		},
		{
			name:     "date format short",
			template: "Date: {{date | formatDate:short}}",
			context: workflow.TemplateContext{
				"date": time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
			},
			want: "Date: Mar 15, 2024",
		},
		{
			name:     "join filter",
			template: "Items: {{items | join:, }}",
			context: workflow.TemplateContext{
				"items": []string{"apple", "banana", "orange"},
			},
			want: "Items: apple, banana, orange",
		},
		{
			name:     "first filter",
			template: "First: {{items | first}}",
			context: workflow.TemplateContext{
				"items": []string{"alpha", "beta", "gamma"},
			},
			want: "First: alpha",
		},
		{
			name:     "last filter",
			template: "Last: {{items | last}}",
			context: workflow.TemplateContext{
				"items": []string{"alpha", "beta", "gamma"},
			},
			want: "Last: gamma",
		},
		{
			name:     "default filter",
			template: "Name: {{name | default:Anonymous}}",
			context: workflow.TemplateContext{
				"name": nil,
			},
			want: "Name: Anonymous",
		},
		{
			name:     "chained filters",
			template: "{{message | uppercase | truncate:5}}",
			context: workflow.TemplateContext{
				"message": "hello world",
			},
			want: "HELLO...",
		},
	}

	opts := workflow.DefaultTemplateProcessingOptions()
	processor := workflow.NewTemplateProcessor(opts)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processor.ProcessTemplate(tt.template, tt.context)

			if len(result.Errors) > 0 {
				t.Errorf("unexpected errors: %v", result.Errors)
			}

			got := result.Processed.(string)
			if got != tt.want {
				t.Errorf("ProcessTemplate() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTemplateProcessor_ComplexObject(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name    string
		input   any
		context workflow.TemplateContext
		want    any
	}

	tests := []testCase{
		{
			name: "nested map with templates",
			input: map[string]any{
				"greeting": "Hello {{user.name}}",
				"details": map[string]any{
					"age":     "{{user.age}}",
					"city":    "{{user.address.city}}",
					"country": "{{user.address.country | uppercase}}",
				},
				"items": []any{
					"Item for {{user.name}}",
					"Located in {{user.address.city}}",
				},
			},
			context: workflow.TemplateContext{
				"user": map[string]any{
					"name": "Alice",
					"age":  30,
					"address": map[string]any{
						"city":    "Seattle",
						"country": "usa",
					},
				},
			},
			want: map[string]any{
				"greeting": "Hello Alice",
				"details": map[string]any{
					"age":     "30",
					"city":    "Seattle",
					"country": "USA",
				},
				"items": []any{
					"Item for Alice",
					"Located in Seattle",
				},
			},
		},
		{
			name: "action config with templates",
			input: map[string]any{
				"type": "send_email",
				"config": map[string]any{
					"recipients": []any{"{{customer.email}}"},
					"subject":    "Order {{order.number}} - {{order.status | uppercase}}",
					"template_data": map[string]any{
						"customer_name": "{{customer.name | capitalize}}",
						"order_total":   "{{order.total | currency}}",
						"ship_date":     "{{order.ship_date | formatDate:short}}",
					},
				},
			},
			context: workflow.TemplateContext{
				"customer": map[string]any{
					"name":  "john doe",
					"email": "john@example.com",
				},
				"order": map[string]any{
					"number":    "ORD-12345",
					"status":    "processing",
					"total":     199.99,
					"ship_date": time.Date(2024, 3, 20, 0, 0, 0, 0, time.UTC),
				},
			},
			want: map[string]any{
				"type": "send_email",
				"config": map[string]any{
					"recipients": []any{"john@example.com"},
					"subject":    "Order ORD-12345 - PROCESSING",
					"template_data": map[string]any{
						"customer_name": "John doe",
						"order_total":   "$199.99",
						"ship_date":     "Mar 20, 2024",
					},
				},
			},
		},
	}

	opts := workflow.DefaultTemplateProcessingOptions()
	processor := workflow.NewTemplateProcessor(opts)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processor.ProcessTemplateObject(tt.input, tt.context)

			if len(result.Errors) > 0 {
				t.Errorf("unexpected errors: %v", result.Errors)
			}

			if diff := cmp.Diff(tt.want, result.Processed); diff != "" {
				t.Errorf("ProcessTemplateObject() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestTemplateProcessor_WorkflowContext(t *testing.T) {
	t.Parallel()

	// Simulate a real workflow execution context
	context := workflow.TemplateContext{
		"entity_id":   "cust_123",
		"entity_name": "customers",
		"event_type":  "on_update",
		"timestamp":   time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC),
		"user_id":     "user_456",
		"rule_id":     "rule_789",
		"rule_name":   "Customer Update Notification",

		// Customer data
		"id":           "cust_123",
		"name":         "john smith",
		"email":        "john.smith@example.com",
		"status":       "premium",
		"total_orders": 42,
		"last_order": map[string]any{
			"id":     "ord_999",
			"total":  549.99,
			"status": "shipped",
		},

		// Field changes for update events
		"field_changes": map[string]any{
			"status": map[string]any{
				"old_value": "regular",
				"new_value": "premium",
			},
		},
	}

	notificationConfig := map[string]any{
		"recipients": []any{"{{email}}"},
		"subject":    "Status Update - {{name | capitalize}}",
		"body": "Dear {{name | capitalize}},\n\n" +
			"Your status has been updated to {{status | uppercase}}.\n" +
			"You have placed {{total_orders}} orders with us.\n" +
			"Your last order ({{last_order.id}}) totaled {{last_order.total | currency}}.\n\n" +
			"Event: {{event_type}} at {{timestamp | formatDate:datetime}}",
		"priority": "{{status | uppercase}}",
	}

	opts := workflow.DefaultTemplateProcessingOptions()
	processor := workflow.NewTemplateProcessor(opts)

	result := processor.ProcessTemplateObject(notificationConfig, context)

	if len(result.Errors) > 0 {
		t.Errorf("unexpected errors: %v", result.Errors)
	}

	// Check specific fields
	processed := result.Processed.(map[string]any)

	recipients := processed["recipients"].([]any)
	if recipients[0] != "john.smith@example.com" {
		t.Errorf("recipients = %v, want [john.smith@example.com]", recipients)
	}

	subject := processed["subject"].(string)
	if subject != "Status Update - John smith" {
		t.Errorf("subject = %q, want %q", subject, "Status Update - John smith")
	}

	priority := processed["priority"].(string)
	if priority != "PREMIUM" {
		t.Errorf("priority = %q, want %q", priority, "PREMIUM")
	}

	// Check that all variables were resolved
	if len(result.VariablesUsed) < 5 {
		t.Errorf("expected at least 5 variables used, got %d", len(result.VariablesUsed))
	}

	// Verify no warnings (all variables should be found)
	if len(result.Warnings) > 0 {
		t.Errorf("unexpected warnings: %v", result.Warnings)
	}
}

func TestTemplateProcessor_CustomFilters(t *testing.T) {
	t.Parallel()

	// Create custom filters
	customFilters := map[string]workflow.FilterFunc{
		"maskEmail": func(value any, args ...string) (any, error) {
			email := value.(string)
			parts := strings.Split(email, "@")
			if len(parts) != 2 {
				return email, nil
			}
			masked := parts[0][:1] + "***@" + parts[1]
			return masked, nil
		},
		"percentage": func(value any, args ...string) (any, error) {
			num, _ := value.(float64)
			return fmt.Sprintf("%.1f%%", num*100), nil
		},
	}

	opts := workflow.TemplateProcessingOptions{
		CustomFilters: customFilters,
		AllowNested:   true,
	}

	processor := workflow.NewTemplateProcessor(opts)

	template := "Email: {{email | maskEmail}}, Success Rate: {{rate | percentage}}"
	context := workflow.TemplateContext{
		"email": "john.doe@example.com",
		"rate":  0.875,
	}

	result := processor.ProcessTemplate(template, context)

	if len(result.Errors) > 0 {
		t.Errorf("unexpected errors: %v", result.Errors)
	}

	want := "Email: j***@example.com, Success Rate: 87.5%"
	got := result.Processed.(string)
	if got != want {
		t.Errorf("ProcessTemplate() = %q, want %q", got, want)
	}
}

func TestTemplateProcessor_ErrorHandling(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		template      string
		context       workflow.TemplateContext
		opts          workflow.TemplateProcessingOptions
		wantErrors    int
		wantWarnings  int
		errorContains string
	}{
		{
			name:          "invalid variable syntax",
			template:      "Hello {{123invalid}}",
			context:       workflow.TemplateContext{},
			opts:          workflow.DefaultTemplateProcessingOptions(),
			wantErrors:    1,
			errorContains: "invalid variable name",
		},
		{
			name:          "invalid filter name",
			template:      "Hello {{name | 123filter}}",
			context:       workflow.TemplateContext{"name": "John"},
			opts:          workflow.DefaultTemplateProcessingOptions(),
			wantErrors:    1,
			errorContains: "invalid filter name",
		},
		{
			name:         "unknown filter",
			template:     "Hello {{name | unknownFilter}}",
			context:      workflow.TemplateContext{"name": "John"},
			opts:         workflow.DefaultTemplateProcessingOptions(),
			wantErrors:   0, // Filter errors are handled gracefully
			wantWarnings: 0,
		},
		{
			name:     "missing variable non-strict",
			template: "Hello {{missing}}",
			context:  workflow.TemplateContext{},
			opts: workflow.TemplateProcessingOptions{
				StrictMode:   false,
				DefaultValue: "DEFAULT",
			},
			wantWarnings: 1,
		},
		{
			name:     "missing variable strict mode",
			template: "Hello {{missing}}",
			context:  workflow.TemplateContext{},
			opts: workflow.TemplateProcessingOptions{
				StrictMode: true,
			},
			wantErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := workflow.NewTemplateProcessor(tt.opts)
			result := processor.ProcessTemplate(tt.template, tt.context)

			if len(result.Errors) != tt.wantErrors {
				t.Errorf("errors = %d, want %d. Errors: %v", len(result.Errors), tt.wantErrors, result.Errors)
			}

			if len(result.Warnings) != tt.wantWarnings {
				t.Errorf("warnings = %d, want %d. Warnings: %v", len(result.Warnings), tt.wantWarnings, result.Warnings)
			}

			if tt.errorContains != "" && len(result.Errors) > 0 {
				found := false
				for _, err := range result.Errors {
					if strings.Contains(err, tt.errorContains) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error containing %q, got errors: %v", tt.errorContains, result.Errors)
				}
			}
		})
	}
}

func TestTemplateProcessor_RawJSON(t *testing.T) {
	t.Parallel()

	// Test handling of json.RawMessage
	jsonData := json.RawMessage(`{
		"message": "Hello {{name}}",
		"details": {
			"user": "{{user.id}}",
			"status": "{{status | uppercase}}"
		}
	}`)

	var parsed any
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal test data: %v", err)
	}

	context := workflow.TemplateContext{
		"name":   "Alice",
		"status": "active",
		"user": map[string]any{
			"id": "user_123",
		},
	}

	opts := workflow.DefaultTemplateProcessingOptions()
	processor := workflow.NewTemplateProcessor(opts)

	result := processor.ProcessTemplateObject(parsed, context)

	if len(result.Errors) > 0 {
		t.Errorf("unexpected errors: %v", result.Errors)
	}

	// Verify the structure
	processed := result.Processed.(map[string]any)
	if processed["message"] != "Hello Alice" {
		t.Errorf("message = %q, want %q", processed["message"], "Hello Alice")
	}

	details := processed["details"].(map[string]any)
	if details["user"] != "user_123" {
		t.Errorf("user = %q, want %q", details["user"], "user_123")
	}
	if details["status"] != "ACTIVE" {
		t.Errorf("status = %q, want %q", details["status"], "ACTIVE")
	}
}

func TestTemplateProcessor_VariableTracking(t *testing.T) {
	t.Parallel()

	template := "User {{user.name}} ({{user.id}}) has {{orders | first}} order with status {{status | uppercase}}"
	context := workflow.TemplateContext{
		"user": map[string]any{
			"name": "Bob",
			"id":   "usr_456",
		},
		"orders": []string{"ORD-001", "ORD-002"},
		"status": "pending",
	}

	opts := workflow.DefaultTemplateProcessingOptions()
	processor := workflow.NewTemplateProcessor(opts)

	result := processor.ProcessTemplate(template, context)

	if len(result.Errors) > 0 {
		t.Errorf("unexpected errors: %v", result.Errors)
	}

	// Check that all variables were tracked
	if len(result.VariablesUsed) != 4 {
		t.Errorf("expected 4 variables tracked, got %d", len(result.VariablesUsed))
	}

	// Verify variable details
	varMap := make(map[string]workflow.TemplateVariable)
	for _, v := range result.VariablesUsed {
		varMap[v.Name] = v
	}

	// Check nested variable
	if userVar, ok := varMap["user.name"]; ok {
		if userVar.Value != "Bob" {
			t.Errorf("user.name value = %v, want Bob", userVar.Value)
		}
		if userVar.Source != "context" {
			t.Errorf("user.name source = %s, want context", userVar.Source)
		}
	} else {
		t.Error("user.name variable not tracked")
	}

	// Check filtered variable
	if orderVar, ok := varMap["orders | first"]; ok {
		if orderVar.Value != "ORD-001" {
			t.Errorf("orders | first value = %v, want ORD-001", orderVar.Value)
		}
		if orderVar.Source != "computed" {
			t.Errorf("orders | first source = %s, want computed", orderVar.Source)
		}
	} else {
		t.Error("orders | first variable not tracked")
	}
}

func TestTemplateProcessor_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		template string
		context  workflow.TemplateContext
		want     string
	}{
		{
			name:     "empty template",
			template: "",
			context:  workflow.TemplateContext{"name": "Test"},
			want:     "",
		},
		{
			name:     "no closing braces",
			template: "Hello {{name",
			context:  workflow.TemplateContext{"name": "Test"},
			want:     "Hello {{name",
		},
		{
			name:     "multiple spaces in variable",
			template: "Hello {{  name  }}",
			context:  workflow.TemplateContext{"name": "Test"},
			want:     "Hello Test",
		},
		{
			name:     "escaped braces",
			template: "Show literal \\{\\{variable\\}\\}",
			context:  workflow.TemplateContext{},
			want:     "Show literal \\{\\{variable\\}\\}",
		},
		{
			name:     "nil value",
			template: "Value: {{nullValue}}",
			context:  workflow.TemplateContext{"nullValue": nil},
			want:     "Value: ",
		},
		{
			name:     "boolean value",
			template: "Active: {{isActive}}",
			context:  workflow.TemplateContext{"isActive": true},
			want:     "Active: true",
		},
		{
			name:     "numeric values",
			template: "Int: {{intVal}}, Float: {{floatVal}}",
			context: workflow.TemplateContext{
				"intVal":   42,
				"floatVal": 3.14,
			},
			want: "Int: 42, Float: 3.14",
		},
	}

	opts := workflow.DefaultTemplateProcessingOptions()
	processor := workflow.NewTemplateProcessor(opts)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processor.ProcessTemplate(tt.template, tt.context)
			got := result.Processed.(string)
			if got != tt.want {
				t.Errorf("ProcessTemplate() = %q, want %q", got, tt.want)
			}
		})
	}
}

// Benchmark tests
func BenchmarkTemplateProcessor_Simple(b *testing.B) {
	processor := workflow.NewTemplateProcessor(workflow.DefaultTemplateProcessingOptions())
	template := "Hello {{name}}, welcome to {{company}}!"
	context := workflow.TemplateContext{
		"name":    "John",
		"company": "Acme Corp",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = processor.ProcessTemplate(template, context)
	}
}

func BenchmarkTemplateProcessor_Complex(b *testing.B) {
	processor := workflow.NewTemplateProcessor(workflow.DefaultTemplateProcessingOptions())
	template := "User {{user.profile.name | capitalize}} ({{user.id}}) has {{orders | first}} with total {{total | currency}}"
	context := workflow.TemplateContext{
		"user": map[string]any{
			"id": "usr_123",
			"profile": map[string]any{
				"name": "john doe",
			},
		},
		"orders": []string{"ORD-001", "ORD-002"},
		"total":  199.99,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = processor.ProcessTemplate(template, context)
	}
}

func BenchmarkTemplateProcessor_Object(b *testing.B) {
	processor := workflow.NewTemplateProcessor(workflow.DefaultTemplateProcessingOptions())
	obj := map[string]any{
		"greeting": "Hello {{name}}",
		"details": map[string]any{
			"age":  "{{age}}",
			"city": "{{city | uppercase}}",
		},
		"items": []any{
			"Item for {{name}}",
			"Age: {{age}}",
		},
	}
	context := workflow.TemplateContext{
		"name": "Alice",
		"age":  30,
		"city": "seattle",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = processor.ProcessTemplateObject(obj, context)
	}
}
