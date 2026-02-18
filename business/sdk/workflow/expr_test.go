package workflow_test

import (
	"testing"

	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

func TestEvalExpr(t *testing.T) {
	t.Parallel()

	vars := map[string]any{
		"quantity":    float64(5),
		"unit_price":  float64(10.5),
		"subtotal":    float64(52.5),
		"tax_rate":    float64(0.1),
		"discount":    float64(5),
		"zero":        float64(0),
		"str_num":     "3.5",
		"int_val":     int(8),
		"int64_val":   int64(12),
	}

	tests := []struct {
		name    string
		expr    string
		want    float64
		wantErr bool
	}{
		{
			name: "multiplication",
			expr: "quantity * unit_price",
			want: 52.5,
		},
		{
			name: "addition",
			expr: "quantity + 3",
			want: 8,
		},
		{
			name: "subtraction",
			expr: "subtotal - discount",
			want: 47.5,
		},
		{
			name: "division",
			expr: "subtotal / quantity",
			want: 10.5,
		},
		{
			name: "modulo",
			expr: "int_val % quantity",
			want: 3,
		},
		{
			name: "parentheses",
			expr: "(2 + 3) * 4",
			want: 20,
		},
		{
			name: "complex with parentheses",
			expr: "subtotal * (1 + tax_rate)",
			want: 57.75,
		},
		{
			name: "literal integer",
			expr: "42",
			want: 42,
		},
		{
			name: "literal float",
			expr: "3.14",
			want: 3.14,
		},
		{
			name: "string numeric variable",
			expr: "str_num + quantity",
			want: 8.5,
		},
		{
			name: "int type variable",
			expr: "int_val + quantity",
			want: 13,
		},
		{
			name: "int64 type variable",
			expr: "int64_val * 2",
			want: 24,
		},
		{
			name: "unary negation",
			expr: "-quantity",
			want: -5,
		},
		{
			name: "nested parentheses",
			expr: "((2 + 3) * (4 - 1)) / 3",
			want: 5,
		},
		{
			name:    "unknown variable",
			expr:    "missing_var",
			wantErr: true,
		},
		{
			name:    "division by zero",
			expr:    "quantity / zero",
			wantErr: true,
		},
		{
			name:    "modulo by zero",
			expr:    "quantity % zero",
			wantErr: true,
		},
		{
			name:    "division by zero literal",
			expr:    "5 / 0",
			wantErr: true,
		},
		{
			name:    "missing closing paren",
			expr:    "(quantity * 2",
			wantErr: true,
		},
		{
			name:    "unexpected character",
			expr:    "quantity @ unit_price",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := workflow.EvalExpr(tt.expr, vars)
			if tt.wantErr {
				if err == nil {
					t.Errorf("EvalExpr(%q) expected error, got %v", tt.expr, got)
				}
				return
			}
			if err != nil {
				t.Errorf("EvalExpr(%q) unexpected error: %v", tt.expr, err)
				return
			}
			// Use epsilon comparison for floating point.
			diff := got - tt.want
			if diff < -1e-9 || diff > 1e-9 {
				t.Errorf("EvalExpr(%q) = %v, want %v", tt.expr, got, tt.want)
			}
		})
	}
}

func TestTemplateExprSubstitution(t *testing.T) {
	t.Parallel()

	opts := workflow.DefaultTemplateProcessingOptions()
	processor := workflow.NewTemplateProcessor(opts)

	tests := []struct {
		name     string
		template string
		context  workflow.TemplateContext
		want     string
		wantWarn bool
	}{
		{
			name:     "basic multiplication",
			template: "Total: {{expr: quantity * unit_price}}",
			context: workflow.TemplateContext{
				"quantity":   float64(5),
				"unit_price": float64(10.5),
			},
			want: "Total: 52.5",
		},
		{
			name:     "integer result",
			template: "Count: {{expr: 3 * 4}}",
			context:  workflow.TemplateContext{},
			want:     "Count: 12",
		},
		{
			name:     "expr with currency filter",
			template: "Amount: {{expr: quantity * unit_price | currency:USD}}",
			context: workflow.TemplateContext{
				"quantity":   float64(5),
				"unit_price": float64(10.5),
			},
			want: "Amount: $52.50",
		},
		{
			name:     "expr with round filter",
			template: "Value: {{expr: 10 / 3 | round:2}}",
			context:  workflow.TemplateContext{},
			want:     "Value: 3.33",
		},
		{
			name:     "multiple expr blocks",
			template: "Subtotal: {{expr: quantity * unit_price}}, Tax: {{expr: subtotal * tax_rate}}",
			context: workflow.TemplateContext{
				"quantity":   float64(5),
				"unit_price": float64(10.5),
				"subtotal":   float64(52.5),
				"tax_rate":   float64(0.1),
			},
			want: "Subtotal: 52.5, Tax: 5.25",
		},
		{
			name:     "expr mixed with regular variables",
			template: "Order {{order_id}}: {{expr: quantity * unit_price | currency}}",
			context: workflow.TemplateContext{
				"order_id":   "ORD-001",
				"quantity":   float64(3),
				"unit_price": float64(25.0),
			},
			want: "Order ORD-001: $75.00",
		},
		{
			name:     "unknown variable fails open",
			template: "{{expr: missing_var * 2}}",
			context:  workflow.TemplateContext{},
			want:     "{{expr: missing_var * 2}}",
			wantWarn: true,
		},
		{
			name:     "division by zero fails open",
			template: "Result: {{expr: 5 / 0}}",
			context:  workflow.TemplateContext{},
			want:     "Result: {{expr: 5 / 0}}",
			wantWarn: true,
		},
		{
			name:     "complex parentheses expression",
			template: "Tax: {{expr: (subtotal + shipping) * tax_rate}}",
			context: workflow.TemplateContext{
				"subtotal":  float64(100),
				"shipping":  float64(15),
				"tax_rate":  float64(0.08),
			},
			want: "Tax: 9.2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processor.ProcessTemplate(tt.template, tt.context)

			if len(result.Errors) > 0 {
				t.Errorf("unexpected errors: %v", result.Errors)
			}

			if tt.wantWarn && len(result.Warnings) == 0 {
				t.Errorf("expected warning for failed expr eval, got none")
			}

			got, ok := result.Processed.(string)
			if !ok {
				t.Fatalf("result.Processed is not a string: %T", result.Processed)
			}

			if got != tt.want {
				t.Errorf("ProcessTemplate() = %q, want %q", got, tt.want)
			}
		})
	}
}
