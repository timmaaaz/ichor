package formfieldbus

// AllowedFieldTypes defines valid form field types
var AllowedFieldTypes = map[string]bool{
	"text":           true,
	"textarea":       true,
	"number":         true,
	"email":          true,
	"tel":            true,
	"date":           true,
	"datetime":       true,
	"time":           true,
	"boolean":        true,
	"hidden":         true,
	"smart-combobox": true,
	"dropdown":       true,
	"enum":           true,
	"currency":       true,
	"percent":        true,
	"lineitems":      true,
}

// AllowedLineItemFieldTypes defines valid field types within line items
var AllowedLineItemFieldTypes = map[string]bool{
	"text":     true,
	"textarea": true,
	"number":   true,
	"currency": true,
	"percent":  true,
	"dropdown": true,
	"enum":     true,
	"hidden":   true,
	"date":     true,
}

// AuditFields defines the standard audit columns.
// NOTE: These are used by deep/schema validation (Part 3) to check if audit
// columns exist in the database table. Structural validation does not use this
// because it cannot verify column existence without a database connection.
var AuditFields = []string{
	"created_by",
	"created_date",
	"updated_by",
	"updated_date",
}
