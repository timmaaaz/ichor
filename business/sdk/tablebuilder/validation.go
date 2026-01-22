package tablebuilder

import (
	"fmt"
	"regexp"
	"strings"
)

// =============================================================================
// Validation Result Types
// =============================================================================

// ValidationResult contains all validation errors and warnings
type ValidationResult struct {
	Errors   []ValidationError
	Warnings []ValidationWarning
}

// ValidationError represents a single validation error
type ValidationError struct {
	Field   string // Dot-notation path to the field (e.g., "data_source[0].filters[0].operator")
	Message string // Human-readable error message
	Code    string // Machine-readable error code (e.g., "REQUIRED", "INVALID_VALUE")
}

// ValidationWarning represents a non-fatal validation warning
type ValidationWarning struct {
	Field   string
	Message string
}

// HasErrors returns true if there are any validation errors
func (r *ValidationResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// AddError adds a validation error to the result
func (r *ValidationResult) AddError(field, message, code string) {
	r.Errors = append(r.Errors, ValidationError{
		Field:   field,
		Message: message,
		Code:    code,
	})
}

// AddWarning adds a validation warning to the result
func (r *ValidationResult) AddWarning(field, message string) {
	r.Warnings = append(r.Warnings, ValidationWarning{
		Field:   field,
		Message: message,
	})
}

// Error implements the error interface, returning a formatted error message
func (r *ValidationResult) Error() string {
	if !r.HasErrors() {
		return ""
	}

	var msgs []string
	for _, err := range r.Errors {
		msgs = append(msgs, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return strings.Join(msgs, "; ")
}

// =============================================================================
// Metric Validation Functions
// =============================================================================

// columnRefPattern validates column references (table.column, schema.table.column, or column)
var columnRefPattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*(\.[a-zA-Z_][a-zA-Z0-9_]*)*$`)

// ValidateMetricConfig validates a metric configuration
func ValidateMetricConfig(metric MetricConfig) error {
	if metric.Name == "" {
		return fmt.Errorf("metric name is required")
	}

	if _, ok := AllowedAggregateFunctions[metric.Function]; !ok {
		return fmt.Errorf("invalid aggregate function: %s", metric.Function)
	}

	if metric.Column == "" && metric.Expression == nil {
		return fmt.Errorf("metric must have column or expression")
	}

	if metric.Column != "" && metric.Expression != nil {
		return fmt.Errorf("metric cannot have both column and expression")
	}

	if metric.Column != "" {
		if !isValidColumnReference(metric.Column) {
			return fmt.Errorf("invalid column reference: %s", metric.Column)
		}
	}

	if metric.Expression != nil {
		if err := ValidateExpressionConfig(metric.Expression); err != nil {
			return fmt.Errorf("invalid expression: %w", err)
		}
	}

	return nil
}

// ValidateExpressionConfig validates an expression configuration
func ValidateExpressionConfig(expr *ExpressionConfig) error {
	if expr == nil {
		return fmt.Errorf("expression is nil")
	}

	if _, ok := AllowedOperators[expr.Operator]; !ok {
		return fmt.Errorf("invalid operator: %s", expr.Operator)
	}

	if len(expr.Columns) < 2 {
		return fmt.Errorf("expression requires at least 2 columns")
	}

	// Validate column names (no SQL injection)
	for _, col := range expr.Columns {
		if !isValidColumnReference(col) {
			return fmt.Errorf("invalid column reference: %s", col)
		}
	}

	return nil
}

// ValidateGroupByConfig validates a group by configuration
func ValidateGroupByConfig(groupBy *GroupByConfig) error {
	if groupBy == nil {
		return nil // GroupBy is optional
	}

	if groupBy.Column == "" {
		return fmt.Errorf("group by column is required")
	}

	if groupBy.Interval != "" {
		if _, ok := AllowedIntervals[groupBy.Interval]; !ok {
			return fmt.Errorf("invalid interval: %s", groupBy.Interval)
		}
	}

	// When Expression is true, Column contains raw SQL (e.g., "EXTRACT(DOW FROM created_date)")
	// and won't match the identifier pattern. The query builder handles raw SQL safely using
	// goqu.L(). This is safe because chart configurations are admin-controlled backend data.
	if !groupBy.Expression && !isValidColumnReference(groupBy.Column) {
		return fmt.Errorf("invalid column reference: %s", groupBy.Column)
	}

	if groupBy.Alias != "" && !isValidColumnReference(groupBy.Alias) {
		return fmt.Errorf("invalid alias: %s", groupBy.Alias)
	}

	return nil
}

// isValidColumnReference checks if a column reference is safe
// Allows: table.column, schema.table.column, column
// Disallows: SQL keywords, special characters, etc.
func isValidColumnReference(col string) bool {
	return columnRefPattern.MatchString(col)
}

// =============================================================================
// Comprehensive Configuration Validation
// =============================================================================

// ValidateConfig performs comprehensive validation on the configuration.
// Returns a ValidationResult containing all errors and warnings found.
// This is more thorough than the basic Validate() method.
func (c *Config) ValidateConfig() *ValidationResult {
	result := &ValidationResult{}

	// 1. Root-level validation
	c.validateRoot(result)

	// 2. DataSource validation
	for i, ds := range c.DataSource {
		c.validateDataSource(result, ds, fmt.Sprintf("data_source[%d]", i))
	}

	// 3. VisualSettings validation (for table widgets)
	if c.WidgetType != "chart" {
		c.validateVisualSettings(result)
	}

	// 4. Permissions validation
	c.validatePermissions(result)

	return result
}

// validateRoot validates root-level Config fields
func (c *Config) validateRoot(result *ValidationResult) {
	if c.Title == "" {
		result.AddError("title", "title is required", "REQUIRED")
	}

	if c.WidgetType != "" && !AllowedWidgetTypes[c.WidgetType] {
		result.AddError("widget_type", fmt.Sprintf("invalid widget type: %s", c.WidgetType), "INVALID_VALUE")
	}

	if c.RefreshMode != "" && !AllowedRefreshModes[c.RefreshMode] {
		result.AddError("refresh_mode", fmt.Sprintf("invalid refresh mode: %s", c.RefreshMode), "INVALID_VALUE")
	}

	if c.RefreshInterval < 0 {
		result.AddError("refresh_interval", "refresh_interval must be >= 0", "INVALID_VALUE")
	}

	if c.PositionX < 0 {
		result.AddError("position_x", "position_x must be >= 0", "INVALID_VALUE")
	}

	if c.PositionY < 0 {
		result.AddError("position_y", "position_y must be >= 0", "INVALID_VALUE")
	}

	if c.Width < 0 {
		result.AddError("width", "width must be >= 0", "INVALID_VALUE")
	}

	if c.Height < 0 {
		result.AddError("height", "height must be >= 0", "INVALID_VALUE")
	}

	if len(c.DataSource) == 0 {
		result.AddError("data_source", "at least one data source is required", "REQUIRED")
	}
}

// validateDataSource validates a DataSource configuration
func (c *Config) validateDataSource(result *ValidationResult, ds DataSource, prefix string) {
	if ds.Source == "" {
		result.AddError(prefix+".source", "source is required", "REQUIRED")
	}

	if ds.Type != "" && !AllowedDataSourceTypes[ds.Type] {
		result.AddError(prefix+".type", fmt.Sprintf("invalid type: %s", ds.Type), "INVALID_VALUE")
	}

	if ds.Rows < 0 {
		result.AddError(prefix+".rows", "rows must be >= 0", "INVALID_VALUE")
	}

	// Validate filters
	for i, f := range ds.Filters {
		c.validateFilter(result, f, fmt.Sprintf("%s.filters[%d]", prefix, i))
	}

	// Validate sorts
	for i, s := range ds.Sort {
		c.validateSort(result, s, fmt.Sprintf("%s.sort[%d]", prefix, i))
	}

	// Validate joins
	for i, j := range ds.Joins {
		c.validateJoin(result, j, fmt.Sprintf("%s.joins[%d]", prefix, i))
	}

	// Validate select columns
	for i, col := range ds.Select.Columns {
		c.validateColumnDefinition(result, col, fmt.Sprintf("%s.select.columns[%d]", prefix, i))
	}

	// Validate foreign tables
	for i, ft := range ds.Select.ForeignTables {
		c.validateForeignTable(result, ft, fmt.Sprintf("%s.select.foreign_tables[%d]", prefix, i))
	}

	// Validate computed columns
	for i, cc := range ds.Select.ClientComputedColumns {
		c.validateComputedColumn(result, cc, fmt.Sprintf("%s.select.client_computed_columns[%d]", prefix, i))
	}

	// Validate metrics (for charts)
	for i, m := range ds.Metrics {
		c.validateMetric(result, m, fmt.Sprintf("%s.metrics[%d]", prefix, i))
	}

	// Validate group_by (for charts)
	for i, g := range ds.GroupBy {
		c.validateGroupBy(result, g, fmt.Sprintf("%s.group_by[%d]", prefix, i))
	}
}

// validateFilter validates a Filter configuration
func (c *Config) validateFilter(result *ValidationResult, f Filter, prefix string) {
	if f.Column == "" {
		result.AddError(prefix+".column", "column is required", "REQUIRED")
	} else if !isValidColumnReference(f.Column) {
		result.AddError(prefix+".column", fmt.Sprintf("invalid column reference: %s", f.Column), "INVALID_FORMAT")
	}

	if f.Operator == "" {
		result.AddError(prefix+".operator", "operator is required", "REQUIRED")
	} else if !AllowedFilterOperators[f.Operator] {
		result.AddError(prefix+".operator", fmt.Sprintf("invalid operator: %s", f.Operator), "INVALID_VALUE")
	}
}

// validateSort validates a Sort configuration
func (c *Config) validateSort(result *ValidationResult, s Sort, prefix string) {
	if s.Column == "" {
		result.AddError(prefix+".column", "column is required", "REQUIRED")
	} else if !isValidColumnReference(s.Column) {
		result.AddError(prefix+".column", fmt.Sprintf("invalid column reference: %s", s.Column), "INVALID_FORMAT")
	}

	if s.Direction == "" {
		result.AddError(prefix+".direction", "direction is required", "REQUIRED")
	} else if !AllowedSortDirections[s.Direction] {
		result.AddError(prefix+".direction", fmt.Sprintf("invalid direction: %s", s.Direction), "INVALID_VALUE")
	}

	if s.Priority < 0 {
		result.AddError(prefix+".priority", "priority must be >= 0", "INVALID_VALUE")
	}
}

// validateJoin validates a Join configuration
func (c *Config) validateJoin(result *ValidationResult, j Join, prefix string) {
	if j.Table == "" {
		result.AddError(prefix+".table", "table is required", "REQUIRED")
	}

	if j.Type == "" {
		result.AddError(prefix+".type", "type is required", "REQUIRED")
	} else if !AllowedJoinTypes[j.Type] {
		result.AddError(prefix+".type", fmt.Sprintf("invalid join type: %s", j.Type), "INVALID_VALUE")
	}

	if j.On == "" {
		result.AddError(prefix+".on", "on condition is required", "REQUIRED")
	}
}

// validateColumnDefinition validates a ColumnDefinition configuration
func (c *Config) validateColumnDefinition(result *ValidationResult, col ColumnDefinition, prefix string) {
	if col.Name == "" && col.TableColumn == "" {
		result.AddError(prefix, "column must have name or table_column", "REQUIRED")
	}

	if col.Name != "" && !isValidColumnReference(col.Name) {
		result.AddError(prefix+".name", fmt.Sprintf("invalid column name: %s", col.Name), "INVALID_FORMAT")
	}

	if col.Alias != "" && !isValidColumnReference(col.Alias) {
		result.AddError(prefix+".alias", fmt.Sprintf("invalid alias: %s", col.Alias), "INVALID_FORMAT")
	}

	if col.TableColumn != "" && !isValidColumnReference(col.TableColumn) {
		result.AddError(prefix+".table_column", fmt.Sprintf("invalid table_column: %s", col.TableColumn), "INVALID_FORMAT")
	}
}

// validateForeignTable validates a ForeignTable configuration
func (c *Config) validateForeignTable(result *ValidationResult, ft ForeignTable, prefix string) {
	if ft.Table == "" {
		result.AddError(prefix+".table", "table is required", "REQUIRED")
	}

	if ft.RelationshipFrom == "" {
		result.AddError(prefix+".relationship_from", "relationship_from is required", "REQUIRED")
	} else if !isValidColumnReference(ft.RelationshipFrom) {
		result.AddError(prefix+".relationship_from", fmt.Sprintf("invalid column reference: %s", ft.RelationshipFrom), "INVALID_FORMAT")
	}

	if ft.RelationshipTo == "" {
		result.AddError(prefix+".relationship_to", "relationship_to is required", "REQUIRED")
	} else if !isValidColumnReference(ft.RelationshipTo) {
		result.AddError(prefix+".relationship_to", fmt.Sprintf("invalid column reference: %s", ft.RelationshipTo), "INVALID_FORMAT")
	}

	if ft.JoinType != "" && !AllowedJoinTypes[ft.JoinType] {
		result.AddError(prefix+".join_type", fmt.Sprintf("invalid join type: %s", ft.JoinType), "INVALID_VALUE")
	}

	if ft.RelationshipDirection != "" && !AllowedRelationshipDirections[ft.RelationshipDirection] {
		result.AddError(prefix+".relationship_direction", fmt.Sprintf("invalid relationship direction: %s", ft.RelationshipDirection), "INVALID_VALUE")
	}

	// Validate columns in the foreign table
	for i, col := range ft.Columns {
		c.validateColumnDefinition(result, col, fmt.Sprintf("%s.columns[%d]", prefix, i))
	}

	// Recursively validate nested foreign tables
	for i, nested := range ft.ForeignTables {
		c.validateForeignTable(result, nested, fmt.Sprintf("%s.foreign_tables[%d]", prefix, i))
	}
}

// validateComputedColumn validates a ComputedColumn configuration
func (c *Config) validateComputedColumn(result *ValidationResult, cc ComputedColumn, prefix string) {
	if cc.Name == "" {
		result.AddError(prefix+".name", "name is required", "REQUIRED")
	} else if !isValidColumnReference(cc.Name) {
		result.AddError(prefix+".name", fmt.Sprintf("invalid computed column name: %s", cc.Name), "INVALID_FORMAT")
	}

	if cc.Expression == "" {
		result.AddError(prefix+".expression", "expression is required", "REQUIRED")
	}
}

// validateVisualSettings validates the VisualSettings configuration
func (c *Config) validateVisualSettings(result *ValidationResult) {
	prefix := "visual_settings"

	// Validate each column config
	for name, col := range c.VisualSettings.Columns {
		colPrefix := fmt.Sprintf("%s.columns[%s]", prefix, name)

		if col.Type == "" {
			result.AddError(colPrefix+".type", "type is required", "REQUIRED")
		} else if !IsValidColumnType(col.Type) {
			result.AddError(colPrefix+".type", fmt.Sprintf("invalid column type: %s", col.Type), "INVALID_VALUE")
		}

		if col.Align != "" && !AllowedAlignments[col.Align] {
			result.AddError(colPrefix+".align", fmt.Sprintf("invalid alignment: %s", col.Align), "INVALID_VALUE")
		}

		if col.Width < 0 {
			result.AddError(colPrefix+".width", "width must be >= 0", "INVALID_VALUE")
		}

		// Validate format config
		if col.Format != nil {
			c.validateFormatConfig(result, col.Format, colPrefix+".format")
		}

		// Validate editable config
		if col.Editable != nil {
			c.validateEditableConfig(result, col.Editable, colPrefix+".editable")
		}

		// Validate lookup config when type is "lookup"
		if col.Type == "lookup" && col.Lookup == nil {
			result.AddError(colPrefix+".lookup", "lookup config required when type is 'lookup'", "REQUIRED")
		}

		if col.Lookup != nil {
			c.validateLookupConfig(result, col.Lookup, colPrefix+".lookup")
		}

		// Validate link config
		if col.Link != nil {
			c.validateLinkConfig(result, col.Link, colPrefix+".link")
		}
	}

	// Validate pagination config
	if c.VisualSettings.Pagination != nil {
		c.validatePaginationConfig(result, c.VisualSettings.Pagination, prefix+".pagination")
	}

	// Validate conditional formatting
	for i, cf := range c.VisualSettings.ConditionalFormatting {
		c.validateConditionalFormat(result, cf, fmt.Sprintf("%s.conditional_formatting[%d]", prefix, i))
	}

	// Validate row actions
	for name, action := range c.VisualSettings.RowActions {
		c.validateAction(result, action, fmt.Sprintf("%s.row_actions[%s]", prefix, name))
	}

	// Validate table actions
	for name, action := range c.VisualSettings.TableActions {
		c.validateAction(result, action, fmt.Sprintf("%s.table_actions[%s]", prefix, name))
	}
}

// validateFormatConfig validates a FormatConfig configuration
func (c *Config) validateFormatConfig(result *ValidationResult, f *FormatConfig, prefix string) {
	if f.Type != "" && !AllowedFormatTypes[f.Type] {
		result.AddError(prefix+".type", fmt.Sprintf("invalid format type: %s", f.Type), "INVALID_VALUE")
	}

	if f.Precision < 0 {
		result.AddError(prefix+".precision", "precision must be >= 0", "INVALID_VALUE")
	}
}

// validateEditableConfig validates an EditableConfig configuration
func (c *Config) validateEditableConfig(result *ValidationResult, e *EditableConfig, prefix string) {
	if e.Type == "" {
		result.AddError(prefix+".type", "type is required", "REQUIRED")
	} else if !AllowedEditableTypes[e.Type] {
		result.AddError(prefix+".type", fmt.Sprintf("invalid editable type: %s", e.Type), "INVALID_VALUE")
	}
}

// validateLookupConfig validates a LookupConfig configuration
func (c *Config) validateLookupConfig(result *ValidationResult, l *LookupConfig, prefix string) {
	if l.Entity == "" {
		result.AddError(prefix+".entity", "entity is required", "REQUIRED")
	}

	if l.LabelColumn == "" {
		result.AddError(prefix+".label_column", "label_column is required", "REQUIRED")
	}

	if l.ValueColumn == "" {
		result.AddError(prefix+".value_column", "value_column is required", "REQUIRED")
	}
}

// validateLinkConfig validates a LinkConfig configuration
func (c *Config) validateLinkConfig(result *ValidationResult, l *LinkConfig, prefix string) {
	if l.URL == "" {
		result.AddError(prefix+".url", "url is required", "REQUIRED")
	}

	if l.Label == "" {
		result.AddError(prefix+".label", "label is required", "REQUIRED")
	}
}

// validatePaginationConfig validates a PaginationConfig configuration
func (c *Config) validatePaginationConfig(result *ValidationResult, p *PaginationConfig, prefix string) {
	for i, size := range p.PageSizes {
		if size <= 0 {
			result.AddError(fmt.Sprintf("%s.page_sizes[%d]", prefix, i), "page size must be > 0", "INVALID_VALUE")
		}
	}

	if p.DefaultPageSize < 0 {
		result.AddError(prefix+".default_page_size", "default_page_size must be >= 0", "INVALID_VALUE")
	}

	if p.DefaultPageSize > 0 && len(p.PageSizes) > 0 {
		found := false
		for _, size := range p.PageSizes {
			if size == p.DefaultPageSize {
				found = true
				break
			}
		}
		if !found {
			result.AddWarning(prefix+".default_page_size", "default_page_size should be in page_sizes")
		}
	}
}

// validateConditionalFormat validates a ConditionalFormat configuration
func (c *Config) validateConditionalFormat(result *ValidationResult, cf ConditionalFormat, prefix string) {
	if cf.Column == "" {
		result.AddError(prefix+".column", "column is required", "REQUIRED")
	}

	if cf.Condition == "" {
		result.AddError(prefix+".condition", "condition is required", "REQUIRED")
	} else if !AllowedFilterOperators[cf.Condition] {
		result.AddError(prefix+".condition", fmt.Sprintf("invalid condition: %s", cf.Condition), "INVALID_VALUE")
	}

	if cf.Condition2 != "" && !AllowedFilterOperators[cf.Condition2] {
		result.AddError(prefix+".condition2", fmt.Sprintf("invalid condition2: %s", cf.Condition2), "INVALID_VALUE")
	}
}

// validateAction validates an Action configuration
func (c *Config) validateAction(result *ValidationResult, a Action, prefix string) {
	if a.Name == "" {
		result.AddError(prefix+".name", "name is required", "REQUIRED")
	}

	if a.Label == "" {
		result.AddError(prefix+".label", "label is required", "REQUIRED")
	}

	if a.ActionType == "" {
		result.AddError(prefix+".action_type", "action_type is required", "REQUIRED")
	} else if !AllowedActionTypes[a.ActionType] {
		result.AddError(prefix+".action_type", fmt.Sprintf("invalid action_type: %s", a.ActionType), "INVALID_VALUE")
	}

	// URL is required for link type
	if a.ActionType == "link" && a.URL == "" {
		result.AddError(prefix+".url", "url is required for link action type", "REQUIRED")
	}
}

// validateMetric validates a MetricConfig using the existing ValidateMetricConfig function
func (c *Config) validateMetric(result *ValidationResult, m MetricConfig, prefix string) {
	if err := ValidateMetricConfig(m); err != nil {
		result.AddError(prefix, err.Error(), "INVALID_CONFIG")
	}
}

// validateGroupBy validates a GroupByConfig using the existing ValidateGroupByConfig function
func (c *Config) validateGroupBy(result *ValidationResult, g GroupByConfig, prefix string) {
	if err := ValidateGroupByConfig(&g); err != nil {
		result.AddError(prefix, err.Error(), "INVALID_CONFIG")
	}
}

// validatePermissions validates the Permissions configuration
func (c *Config) validatePermissions(result *ValidationResult) {
	prefix := "permissions"

	for i, action := range c.Permissions.Actions {
		if !AllowedPermissionActions[action] {
			result.AddError(fmt.Sprintf("%s.actions[%d]", prefix, i), fmt.Sprintf("invalid action: %s", action), "INVALID_VALUE")
		}
	}
}
