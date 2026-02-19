package tablebuilder

import "strings"

// ValidColumnTypes defines the allowed column types for table builder configurations.
// All columns in VisualSettings.Columns must have a Type field set to one of these values.
var ValidColumnTypes = map[string]bool{
	"string":   true, // Text, varchar, char types
	"number":   true, // Integer, decimal, numeric types
	"datetime": true, // Date, time, timestamp types
	"boolean":  true, // Boolean type
	"uuid":     true, // UUID type
	"status":   true, // Enum/status fields (renders as dropdown)
	"computed": true, // Client-computed columns
	"lookup":   true, // Lookup dropdown fields (FK references with searchable dropdown)
}

// IsValidColumnType returns true if the given type is a valid column type.
func IsValidColumnType(colType string) bool {
	return ValidColumnTypes[colType]
}

// pgTypeMappings maps PostgreSQL data types from pg_catalog
// to frontend filter types used by the Table Builder.
//
// Type categories:
//   - UUID type → "uuid"
//   - Date/time types → "datetime"
//   - Numeric types → "number"
//   - Boolean → "boolean"
//   - Text/char types → "string"
//   - JSON types → "string"
//   - Unknown → "string" (safe default)
var pgTypeMappings = map[string]string{
	// UUID type
	"uuid": "uuid",

	// Date/time types
	"timestamp":                   "datetime",
	"timestamp without time zone": "datetime",
	"timestamp with time zone":    "datetime",
	"date":                        "datetime",
	"time":                        "datetime",
	"time without time zone":      "datetime",
	"time with time zone":         "datetime",
	"interval":                    "datetime",

	// Numeric types
	"integer":          "number",
	"bigint":           "number",
	"smallint":         "number",
	"numeric":          "number",
	"decimal":          "number",
	"real":             "number",
	"double precision": "number",
	"serial":           "number",
	"bigserial":        "number",
	"smallserial":      "number",
	"money":            "number",

	// Boolean type
	"boolean": "boolean",

	// Text/character types
	"character varying": "string",
	"varchar":           "string",
	"character":         "string",
	"char":              "string",
	"text":              "string",
	"citext":            "string",

	// JSON types
	"json":  "string",
	"jsonb": "string",
}

// MapPostgreSQLType converts PostgreSQL data types from pg_catalog
// to frontend filter types used by the Table Builder.
//
// The function handles:
//   - Direct type matches (e.g., "uuid", "boolean")
//   - Types with modifiers (e.g., "numeric(10,2)" → "number")
//   - Array types (e.g., "integer[]" → "string")
//
// Returns "string" as safe default for unknown types.
func MapPostgreSQLType(pgType string) string {
	lower := strings.ToLower(pgType)

	// Direct lookup for exact matches
	if mapped, ok := pgTypeMappings[lower]; ok {
		return mapped
	}

	// Handle types with modifiers (e.g., "numeric(10,2)")
	if idx := strings.IndexByte(lower, '('); idx != -1 {
		baseType := lower[:idx]
		if mapped, ok := pgTypeMappings[baseType]; ok {
			return mapped
		}
	}

	// Handle array types
	if strings.Contains(lower, "array") || strings.HasSuffix(lower, "[]") {
		return "string"
	}

	// Default to string for unknown types
	return "string"
}
