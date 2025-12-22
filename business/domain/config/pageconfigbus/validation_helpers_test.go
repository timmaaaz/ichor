package pageconfigbus

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/uuid"
)

// TestSanitizeValue tests value sanitization for error messages
func TestSanitizeValue(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  interface{}
	}{
		{
			name:  "nil value",
			value: nil,
			want:  nil,
		},
		{
			name:  "short string",
			value: "hello",
			want:  "hello",
		},
		{
			name:  "long string truncated",
			value: "this is a very long string that exceeds the maximum length allowed for error messages and should be truncated",
			want:  "this is a very long string that exceeds the maximum length allowed for error messages and should ...",
		},
		{
			name:  "potential secret redacted",
			value: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0",
			want:  "[REDACTED]",
		},
		{
			name:  "number not sanitized",
			value: 42,
			want:  42,
		},
		{
			name:  "float not sanitized",
			value: 3.14,
			want:  3.14,
		},
		{
			name:  "bool not sanitized",
			value: true,
			want:  true,
		},
		{
			name:  "complex type description",
			value: struct{ Name string }{Name: "test"},
			want:  "<struct { Name string }>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeValue(tt.value)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sanitizeValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Note: collectReferences function doesn't exist in the current implementation
// The reference validation is done differently in the actual code

// TestValidateSingleContent tests single content validation
func TestValidateSingleContent(t *testing.T) {
	ctx := context.Background()
	tableID := uuid.New()

	tests := []struct {
		name       string
		content    PageContentExport
		wantErrors int
		wantCodes  []string
	}{
		{
			name: "valid table content",
			content: PageContentExport{
				ContentType:   "table",
				TableConfigID: tableID,
			},
			wantErrors: 0,
		},
		{
			name: "invalid content type",
			content: PageContentExport{
				ContentType: "invalid_type",
			},
			wantErrors: 1,
			wantCodes:  []string{ErrCodeInvalidType},
		},
		{
			name: "table without tableConfigId",
			content: PageContentExport{
				ContentType:   "table",
				TableConfigID: uuid.Nil,
			},
			wantErrors: 1,
			wantCodes:  []string{ErrCodeRequiredField},
		},
		{
			name: "form without formId",
			content: PageContentExport{
				ContentType: "form",
				FormID:      uuid.Nil,
			},
			wantErrors: 1,
			wantCodes:  []string{ErrCodeRequiredField},
		},
		{
			name: "chart without chartConfigId",
			content: PageContentExport{
				ContentType:   "chart",
				ChartConfigID: uuid.Nil,
			},
			wantErrors: 1,
			wantCodes:  []string{ErrCodeRequiredField},
		},
		{
			name: "valid chart content",
			content: PageContentExport{
				ContentType:   "chart",
				ChartConfigID: uuid.New(),
			},
			wantErrors: 0,
		},
		{
			name: "tabs without label",
			content: PageContentExport{
				ContentType: "tabs",
				Label:       "",
			},
			wantErrors: 1,
			wantCodes:  []string{ErrCodeRequiredField},
		},
		{
			name: "text without label",
			content: PageContentExport{
				ContentType: "text",
				Label:       "",
			},
			wantErrors: 1,
			wantCodes:  []string{ErrCodeRequiredField},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validateSingleContent(ctx, tt.content, 0, "contents")

			if len(errors) != tt.wantErrors {
				t.Errorf("validateSingleContent() error count = %v, want %v (errors: %+v)",
					len(errors), tt.wantErrors, errors)
			}

			// Check error codes if specified
			if len(tt.wantCodes) > 0 {
				for i, wantCode := range tt.wantCodes {
					if i < len(errors) && errors[i].Code != wantCode {
						t.Errorf("validateSingleContent() error[%d].Code = %v, want %v",
							i, errors[i].Code, wantCode)
					}
				}
			}
		})
	}
}

// TestValidateLayout tests layout configuration validation
func TestValidateLayout(t *testing.T) {
	tests := []struct {
		name       string
		layout     LayoutConfig
		wantErrors int
		wantPath   string
	}{
		{
			name: "valid layout",
			layout: LayoutConfig{
				ColSpan: &ResponsiveValue{
					Xs: intPtr(12),
					Md: intPtr(6),
				},
			},
			wantErrors: 0,
		},
		{
			name: "colSpan too small",
			layout: LayoutConfig{
				ColSpan: &ResponsiveValue{
					Md: intPtr(0),
				},
			},
			wantErrors: 1,
			wantPath:   "layout.colSpan.md",
		},
		{
			name: "colSpan too large",
			layout: LayoutConfig{
				ColSpan: &ResponsiveValue{
					Lg: intPtr(15),
				},
			},
			wantErrors: 1,
			wantPath:   "layout.colSpan.lg",
		},
		{
			name: "multiple invalid breakpoints",
			layout: LayoutConfig{
				ColSpan: &ResponsiveValue{
					Xs: intPtr(0),
					Sm: intPtr(13),
					Md: intPtr(-1),
				},
			},
			wantErrors: 3,
		},
		{
			name: "nil colSpan valid",
			layout: LayoutConfig{
				ColSpan: nil,
			},
			wantErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validateLayout(tt.layout, "test")

			if len(errors) != tt.wantErrors {
				t.Errorf("validateLayout() error count = %v, want %v", len(errors), tt.wantErrors)
			}

			if tt.wantPath != "" && len(errors) > 0 {
				found := false
				for _, err := range errors {
					if err.Field == "test."+tt.wantPath {
						found = true
						if err.Code != "RANGE_ERROR" {
							t.Errorf("validateLayout() error code = %v, want RANGE_ERROR", err.Code)
						}
						break
					}
				}
				if !found {
					t.Errorf("validateLayout() expected error at path %v not found", tt.wantPath)
				}
			}
		})
	}
}

// TestCalculateNestingDepth tests nesting depth calculation
func TestCalculateNestingDepth(t *testing.T) {
	tests := []struct {
		name      string
		contents  []PageContentExport
		wantDepth int
	}{
		{
			name:      "empty contents",
			contents:  []PageContentExport{},
			wantDepth: 0,
		},
		{
			name: "single level",
			contents: []PageContentExport{
				{ContentType: "table"},
			},
			wantDepth: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			depth := calculateNestingDepth(tt.contents, 1)
			if depth != tt.wantDepth {
				t.Errorf("calculateNestingDepth() = %v, want %v", depth, tt.wantDepth)
			}
		})
	}
}

// TestDetectCircularReferences tests circular reference detection
func TestDetectCircularReferences(t *testing.T) {
	idA := uuid.New()
	idB := uuid.New()
	idC := uuid.New()

	tests := []struct {
		name       string
		contents   []PageContentExport
		wantErrors bool
	}{
		{
			name:       "empty contents",
			contents:   []PageContentExport{},
			wantErrors: false,
		},
		{
			name: "no circular references",
			contents: []PageContentExport{
				{ID: idA, ParentID: uuid.Nil},
				{ID: idB, ParentID: idA},
				{ID: idC, ParentID: idB},
			},
			wantErrors: false,
		},
		{
			name: "simple circular reference",
			contents: []PageContentExport{
				{ID: idA, ParentID: idB},
				{ID: idB, ParentID: idA},
			},
			wantErrors: true,
		},
		{
			name: "complex circular reference",
			contents: []PageContentExport{
				{ID: idA, ParentID: idC},
				{ID: idB, ParentID: idA},
				{ID: idC, ParentID: idB},
			},
			wantErrors: true,
		},
		{
			name: "self-reference",
			contents: []PageContentExport{
				{ID: idA, ParentID: idA},
			},
			wantErrors: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := detectCircularReferences(tt.contents)

			hasErrors := len(errors) > 0
			if hasErrors != tt.wantErrors {
				t.Errorf("detectCircularReferences() hasErrors = %v, want %v (errors: %+v)",
					hasErrors, tt.wantErrors, errors)
			}

			if hasErrors && errors[0].Code != ErrCodeCircularReference {
				t.Errorf("detectCircularReferences() error code = %v, want %v",
					errors[0].Code, ErrCodeCircularReference)
			}
		})
	}
}

// TestContains tests the contains helper function
func TestContains(t *testing.T) {
	tests := []struct {
		name  string
		slice []string
		item  string
		want  bool
	}{
		{
			name:  "item exists",
			slice: []string{"a", "b", "c"},
			item:  "b",
			want:  true,
		},
		{
			name:  "item does not exist",
			slice: []string{"a", "b", "c"},
			item:  "d",
			want:  false,
		},
		{
			name:  "empty slice",
			slice: []string{},
			item:  "a",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := contains(tt.slice, tt.item)
			if got != tt.want {
				t.Errorf("contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper functions for tests

func intPtr(i int) *int {
	return &i
}
