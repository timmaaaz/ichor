package tablebuilder

// =============================================================================
// Centralized Whitelists for Configuration Validation
// =============================================================================
// This file contains all allowed values for the tablebuilder configuration system.
// Using centralized whitelists ensures consistency and makes it easy to add/remove
// valid values in one place.

// AllowedDataSourceTypes defines valid data source types for DataSource.Type
var AllowedDataSourceTypes = map[string]bool{
	"query":     true, // Direct SQL query
	"view":      true, // Database view
	"viewcount": true, // Count query for views
	"rpc":       true, // Remote procedure call
}

// AllowedJoinTypes defines valid SQL join types for Join.Type and ForeignTable.JoinType
var AllowedJoinTypes = map[string]bool{
	"inner": true,
	"left":  true,
	"right": true,
	"full":  true,
}

// AllowedFilterOperators defines valid filter comparison operators for Filter.Operator
var AllowedFilterOperators = map[string]bool{
	"eq":          true, // equals
	"neq":         true, // not equals
	"gt":          true, // greater than
	"gte":         true, // greater than or equal
	"lt":          true, // less than
	"lte":         true, // less than or equal
	"in":          true, // in array
	"like":        true, // pattern match (case-sensitive)
	"ilike":       true, // pattern match (case-insensitive)
	"is_null":     true, // IS NULL check
	"is_not_null": true, // IS NOT NULL check
}

// AllowedSortDirections defines valid sort directions for Sort.Direction
var AllowedSortDirections = map[string]bool{
	"asc":  true,
	"desc": true,
}

// AllowedAlignments defines valid text alignments for ColumnConfig.Align
var AllowedAlignments = map[string]bool{
	"left":   true,
	"center": true,
	"right":  true,
}

// AllowedFormatTypes defines valid format types for FormatConfig.Type
var AllowedFormatTypes = map[string]bool{
	"number":   true,
	"currency": true,
	"date":     true,
	"datetime": true,
	"boolean":  true,
	"percent":  true,
}

// AllowedEditableTypes defines valid editable input types for EditableConfig.Type
var AllowedEditableTypes = map[string]bool{
	"text":     true,
	"number":   true,
	"checkbox": true,
	"boolean":  true, // Alias for checkbox (renders as toggle/checkbox for boolean fields)
	"select":   true,
	"date":     true,
	"textarea": true,
}

// AllowedActionTypes defines valid action types for Action.ActionType
var AllowedActionTypes = map[string]bool{
	"link":   true,
	"modal":  true,
	"export": true,
	"print":  true,
	"custom": true,
}

// AllowedWidgetTypes defines valid widget types for Config.WidgetType
var AllowedWidgetTypes = map[string]bool{
	"table": true,
	"chart": true,
}

// AllowedRefreshModes defines valid refresh modes for Config.RefreshMode
var AllowedRefreshModes = map[string]bool{
	"polling": true,
	"manual":  true,
}

// AllowedPermissionActions defines valid permission actions for Permissions.Actions
var AllowedPermissionActions = map[string]bool{
	"view":    true,
	"edit":    true,
	"delete":  true,
	"export":  true,
	"create":  true,
	"adjust":  true, // For inventory adjustments
	"approve": true, // For approval workflows
	"reject":  true, // For approval workflows
}

// AllowedChartTypes defines valid chart types for chart widgets
// These are exposed as constants in model.go but centralized here for validation
var AllowedChartTypes = map[string]bool{
	ChartTypeLine:        true,
	ChartTypeBar:         true,
	ChartTypeStackedBar:  true,
	ChartTypeStackedArea: true,
	ChartTypeCombo:       true,
	ChartTypeKPI:         true,
	ChartTypeGauge:       true,
	ChartTypePie:         true,
	ChartTypeWaterfall:   true,
	ChartTypeFunnel:      true,
	ChartTypeHeatmap:     true,
	ChartTypeTreemap:     true,
	ChartTypeGantt:       true,
}

// AllowedRelationshipDirections defines valid relationship directions for ForeignTable
var AllowedRelationshipDirections = map[string]bool{
	"parent_to_child": true,
	"child_to_parent": true,
}
