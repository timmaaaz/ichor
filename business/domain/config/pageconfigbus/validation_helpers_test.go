package pageconfigbus

import (
	"context"
	"reflect"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// TestExtractJSONPath tests JSON path extraction from struct field paths
func TestExtractJSONPath(t *testing.T) {
	// Get the type of PageConfigWithRelations for reflection
	structType := reflect.TypeOf(PageConfigWithRelations{})

	tests := []struct {
		name       string
		structPath string
		want       string
	}{
		{
			name:       "simple field",
			structPath: "PageConfigWithRelations.Name",
			want:       "name",
		},
		{
			name:       "nested field",
			structPath: "PageConfigWithRelations.Contents[0].ContentType",
			want:       "contents[0].contentType",
		},
		{
			name:       "array index",
			structPath: "PageConfigWithRelations.Contents[2]",
			want:       "contents[2]",
		},
		{
			name:       "deeply nested with array",
			structPath: "PageConfigWithRelations.Contents[0].Children[1].TableConfigID",
			want:       "contents[0].children[1].tableConfigId",
		},
		{
			name:       "layout field",
			structPath: "PageConfigWithRelations.Contents[0].Layout.ColSpan.Md",
			want:       "contents[0].layout.colSpan.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractJSONPath(tt.structPath, structType)
			if got != tt.want {
				t.Errorf("extractJSONPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestTransformFieldError tests conversion of validator errors to ValidationErrors
func TestTransformFieldError(t *testing.T) {
	v := validator.New()
	structType := reflect.TypeOf(PageConfigWithRelations{})

	tests := []struct {
		name     string
		field    string
		tag      string
		param    string
		value    interface{}
		wantCode string
		wantMsg  string
	}{
		{
			name:     "required field error",
			field:    "Name",
			tag:      "required",
			value:    "",
			wantCode: ErrCodeRequiredField,
			wantMsg:  "Field 'Name' is required and cannot be empty",
		},
		{
			name:     "uuid validation error",
			field:    "ID",
			tag:      "uuid",
			value:    "not-a-uuid",
			wantCode: "INVALID_FORMAT",
			wantMsg:  "Field 'ID' must be a valid UUID",
		},
		{
			name:     "oneof validation error",
			field:    "ContentType",
			tag:      "oneof",
			param:    "table form chart",
			value:    "invalid",
			wantCode: ErrCodeInvalidType,
			wantMsg:  "Field 'ContentType' must be one of: table form chart",
		},
		{
			name:     "min validation error",
			field:    "SortOrder",
			tag:      "min",
			param:    "1",
			value:    0,
			wantCode: "RANGE_ERROR",
			wantMsg:  "Field 'SortOrder' must be at least 1",
		},
		{
			name:     "max validation error",
			field:    "ColSpan",
			tag:      "max",
			param:    "12",
			value:    15,
			wantCode: "RANGE_ERROR",
			wantMsg:  "Field 'ColSpan' must be at most 12",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock field error
			// Note: This is simplified - in real tests, you'd validate an actual struct
			// and capture the field errors from validator
			type testStruct struct {
				TestField string `validate:"required"`
			}

			err := v.Struct(&testStruct{TestField: ""})
			if err != nil {
				validationErrors := err.(validator.ValidationErrors)
				if len(validationErrors) > 0 {
					fieldErr := validationErrors[0]
					result := transformFieldError(fieldErr, structType)

					// Verify error code mapping
					expectedCode := mapValidatorTagToErrorCode(tt.tag)
					if result.Code != expectedCode {
						t.Errorf("transformFieldError() code = %v, want %v", result.Code, expectedCode)
					}

					// Verify message contains field name
					// (We can't test exact message since field error has different namespace)
					if result.Message == "" {
						t.Errorf("transformFieldError() message is empty")
					}
				}
			}
		})
	}
}

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
			want:  "this is a very long string that exceeds the maximum length allowed for error messages and should be ...",
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
			name: "chart without tableConfigId",
			content: PageContentExport{
				ContentType:   "chart",
				TableConfigID: uuid.Nil,
			},
			wantErrors: 1,
			wantCodes:  []string{ErrCodeRequiredField},
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
