package tablebuilder

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages table query operations
type Store struct {
	log     *logger.Logger
	db      *sqlx.DB
	builder *QueryBuilder
	eval    *Evaluator
}

// NewStore creates a new table builder store
func NewStore(log *logger.Logger, db *sqlx.DB) *Store {
	return &Store{
		log:     log,
		db:      db,
		builder: NewQueryBuilder(),
		eval:    NewEvaluator(),
	}
}

func (s *Store) FetchTableDataCount(ctx context.Context, config *Config, params QueryParams) (int, error) {
	if err := config.Validate(); err != nil {
		return 0, fmt.Errorf("validate config: %w", err)
	}

	if len(config.DataSource) == 0 {
		return 0, fmt.Errorf("no data sources defined")
	}

	// Only the primary data source is used for count
	ds := &config.DataSource[0]

	// Create a copy of params without pagination for counting
	countParams := QueryParams{
		Filters: params.Filters,
		Sort:    params.Sort, // Sort doesn't affect count, but keep for consistency
		Dynamic: params.Dynamic,
		// Explicitly exclude Page and Rows
	}

	count, err := s.GetCount(ctx, ds, countParams)
	if err != nil {
		return 0, fmt.Errorf("get count: %w", err)
	}

	return count, nil
}

// FetchTableData executes the table configuration and returns the data
func (s *Store) FetchTableData(ctx context.Context, config *Config, params QueryParams) (*TableData, error) {
	startTime := time.Now()

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	result := &TableData{
		Data: make([]TableRow, 0),
		Meta: MetaData{
			// NEW: Build metadata upfront
			Columns:       s.buildColumnMetadata(config),
			Relationships: s.buildRelationshipMetadata(config),
		},
	}

	// Process each data source
	for i, ds := range config.DataSource {
		data, err := s.processDataSource(ctx, &ds, params, i, result)
		if err != nil {
			return nil, fmt.Errorf("process data source %d: %w", i, err)
		}

		if i == 0 {
			// Primary data source
			result.Data = data

			// Get total count for pagination
			if params.Page > 0 {
				count, err := s.GetCount(ctx, &ds, params)
				if err != nil {
					s.log.Infoc(ctx, 3, "failed to get count", "error", err)
					// Don't fail the whole query if count fails
				} else {
					result.Meta.Total = count
					result.Meta.Page = params.Page
					result.Meta.PageSize = params.Rows
					if params.Rows > 0 {
						result.Meta.TotalPages = (count + params.Rows - 1) / params.Rows
					}
				}
			}
		} else {
			// Secondary data sources - merge with primary data
			if err := s.mergeData(result.Data, data, &ds); err != nil {
				return nil, fmt.Errorf("merge data: %w", err)
			}
		}
	}

	// Apply computed columns
	if len(config.DataSource) > 0 && len(config.DataSource[0].Select.ClientComputedColumns) > 0 {
		if err := s.applyComputedColumns(result.Data, config.DataSource[0].Select.ClientComputedColumns); err != nil {
			s.log.Infoc(ctx, 3, "failed to apply computed columns", "error", err)
			// Don't fail if computed columns fail
		}
	}

	// Transform data for final output
	result.Data = s.transformData(result.Data, config)

	// Set execution time
	result.Meta.ExecutionTime = time.Since(startTime).Milliseconds()

	return result, nil
}

// processDataSource processes a single data source
func (s *Store) processDataSource(ctx context.Context, ds *DataSource, params QueryParams, index int, result *TableData) ([]TableRow, error) {
	// Build the query
	query, args, err := s.builder.BuildQuery(ds, params, index == 0)
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	s.log.Infoc(ctx, 4, "executing query", "query", query, "args", args)

	// Execute the query based on type
	var rows []TableRow

	switch ds.Type {
	case "rpc", "function":
		rows, err = s.executeRPC(ctx, ds.Source, args)
	case "viewcount":
		rows, err = s.executeCount(ctx, query, args)
	default:
		rows, err = s.executeQuery(ctx, query, args)
	}

	if err != nil {
		return nil, fmt.Errorf("execute query: %w", err)
	}

	return rows, nil
}

// executeQuery executes a regular SELECT query
func (s *Store) executeQuery(ctx context.Context, query string, args map[string]interface{}) ([]TableRow, error) {
	rows, err := s.db.NamedQueryContext(ctx, query, args)
	if err != nil {
		return nil, fmt.Errorf("named query: %w", err)
	}
	defer rows.Close()

	var results []TableRow
	for rows.Next() {
		row := make(TableRow)
		if err := rows.MapScan(row); err != nil {
			return nil, fmt.Errorf("map scan: %w", err)
		}
		results = append(results, row)
	}

	return results, nil
}

// executeRPC executes a stored procedure/function
func (s *Store) executeRPC(ctx context.Context, funcName string, args map[string]interface{}) ([]TableRow, error) {
	// Build the function call
	query := fmt.Sprintf("SELECT * FROM %s(:args)", funcName)

	rows, err := s.db.NamedQueryContext(ctx, query, args)
	if err != nil {
		return nil, fmt.Errorf("execute rpc: %w", err)
	}
	defer rows.Close()

	var results []TableRow
	for rows.Next() {
		row := make(TableRow)
		if err := rows.MapScan(row); err != nil {
			return nil, fmt.Errorf("map scan: %w", err)
		}
		results = append(results, row)
	}

	return results, nil
}

// executeCount executes a COUNT query
func (s *Store) executeCount(ctx context.Context, query string, args map[string]interface{}) ([]TableRow, error) {
	var count int
	row := s.db.QueryRowxContext(ctx, query, args)
	if err := row.Scan(&count); err != nil {
		return nil, fmt.Errorf("scan count: %w", err)
	}

	// Return count as a row
	return []TableRow{
		{"count": count},
	}, nil
}

func (s *Store) GetCount(ctx context.Context, ds *DataSource, params QueryParams) (int, error) {
	query, args, err := s.builder.BuildCountQuery(ds, params)
	if err != nil {
		return 0, fmt.Errorf("build count query: %w", err)
	}

	var result struct {
		Count int `db:"count"`
	}

	rows, err := s.db.NamedQueryContext(ctx, query, args)
	if err != nil {
		return 0, fmt.Errorf("named query: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return 0, fmt.Errorf("no count result")
	}

	if err := rows.StructScan(&result); err != nil {
		return 0, fmt.Errorf("scan count: %w", err)
	}

	return result.Count, nil
}

// mergeData merges secondary data source results with primary data
func (s *Store) mergeData(primary []TableRow, secondary []TableRow, ds *DataSource) error {
	if ds.ParentSource == "" || ds.SelectBy == "" {
		return fmt.Errorf("parent_source and select_by required for merging")
	}

	// Parse parent source to get the key field
	parentKey := s.parseParentKey(ds.ParentSource)

	// Create a map of secondary data by select_by field
	secondaryMap := make(map[interface{}][]TableRow)
	for _, row := range secondary {
		if key, ok := row[ds.SelectBy]; ok {
			secondaryMap[key] = append(secondaryMap[key], row)
		}
	}

	// Merge with primary data
	for _, pRow := range primary {
		if parentValue, ok := pRow[parentKey]; ok {
			if matches, found := secondaryMap[parentValue]; found {
				// Add matched rows based on data source type
				switch ds.Type {
				case "viewcount":
					pRow[ds.Source] = len(matches)
				default:
					pRow[ds.Source] = matches
				}
			}
		}
	}

	return nil
}

// parseParentKey extracts the key field from parent source notation
func (s *Store) parseParentKey(parentSource string) string {
	// Handle format like "orders_base.order_id"
	// Return just the last part
	parts := splitPath(parentSource)
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return parentSource
}

// applyComputedColumns applies computed column expressions
func (s *Store) applyComputedColumns(data []TableRow, columns []ComputedColumn) error {
	for _, row := range data {
		for _, col := range columns {
			value, err := s.eval.Evaluate(col.Expression, row)
			if err != nil {
				// Log but don't fail

				// TODO: Evaluate if we need to pass context in here
				ctx := context.Background()

				s.log.Info(ctx, "failed to evaluate expression",
					"column", col.Name,
					"expression", col.Expression,
					"error", err)
				row[col.Name] = nil
			} else {
				row[col.Name] = value
			}
		}
	}
	return nil
}

// NEW: Build column metadata
// In buildColumnMetadata, at the top:
func (s *Store) buildColumnMetadata(config *Config) []ColumnMetadata {
	ds := config.DataSource[0]
	metadata := make([]ColumnMetadata, 0)

	// Determine the primary table name
	primaryTable := strings.TrimSuffix(ds.Source, "_base")

	for _, col := range ds.Select.Columns {
		meta := ColumnMetadata{
			DatabaseName: col.Name,
			Field:        getFieldName(col),
			DisplayName:  getDisplayName(col),
			Type:         inferColumnType(col.Name),
			SourceSchema: ds.Schema,
		}

		// Parse table_column for source info
		if col.TableColumn != "" {
			parts := strings.Split(col.TableColumn, ".")
			if len(parts) == 2 {
				meta.SourceTable = parts[0]
				meta.SourceColumn = parts[1]

				// Mark as primary key if it's the primary table's id
				if parts[0] == primaryTable && parts[1] == "id" {
					meta.IsPrimaryKey = true
					meta.Hidden = true
				}
			}
		}

		// Check for foreign keys
		if strings.HasSuffix(col.Name, "_id") && !meta.IsPrimaryKey {
			meta.IsForeignKey = true
			meta.Hidden = true
		}

		// Visual settings can override hidden
		if _, ok := config.VisualSettings.Columns[meta.Field]; ok {
			meta.Hidden = false
		}

		// Apply visual settings
		if vs, ok := config.VisualSettings.Columns[meta.Field]; ok {
			if vs.Header != "" {
				meta.DisplayName = vs.Header
			}
			meta.Header = vs.Header
			meta.Width = vs.Width
			meta.Align = vs.Align
			meta.Sortable = vs.Sortable
			meta.Filterable = vs.Filterable
			meta.Format = vs.Format
			meta.Editable = vs.Editable
			meta.Link = vs.Link
		}

		metadata = append(metadata, meta)
	}

	// Process foreign table columns
	metadata = append(metadata, s.buildForeignColumnMetadata(ds.Select.ForeignTables, config)...)

	// Process computed columns
	for _, cc := range ds.Select.ClientComputedColumns {
		meta := ColumnMetadata{
			DatabaseName: cc.Name,
			Field:        cc.Name,
			DisplayName:  cc.Name,
			Type:         "computed",
		}

		if vs, ok := config.VisualSettings.Columns[cc.Name]; ok {
			if vs.Header != "" {
				meta.DisplayName = vs.Header
			}
			meta.Header = vs.Header
			meta.Width = vs.Width
			meta.Align = vs.Align
			meta.Format = vs.Format
		}

		metadata = append(metadata, meta)
	}

	return metadata
}

func (s *Store) buildForeignColumnMetadata(foreignTables []ForeignTable, config *Config) []ColumnMetadata {
	metadata := make([]ColumnMetadata, 0)

	for _, ft := range foreignTables {
		// Use alias if present, otherwise use table name
		tableRef := ft.Table
		if ft.Alias != "" {
			tableRef = ft.Alias
		}

		for _, col := range ft.Columns {
			meta := ColumnMetadata{
				DisplayName:  getDisplayName(col),
				Field:        getFieldName(col),
				Type:         inferColumnType(col.Name),
				SourceTable:  tableRef, // Use alias or table name
				SourceColumn: col.Name,
				SourceSchema: ft.Schema,
			}

			if col.TableColumn != "" {
				parts := strings.Split(col.TableColumn, ".")
				if len(parts) >= 2 {
					meta.SourceTable = parts[len(parts)-2]
					meta.SourceColumn = parts[len(parts)-1]
				}
			}

			if vs, ok := config.VisualSettings.Columns[meta.Field]; ok {
				meta.Header = vs.Header
				meta.Width = vs.Width
				meta.Align = vs.Align
				meta.Sortable = vs.Sortable
				meta.Filterable = vs.Filterable
				meta.Format = vs.Format
				meta.Link = vs.Link
			}

			metadata = append(metadata, meta)
		}

		// Recursively process nested foreign tables
		metadata = append(metadata, s.buildForeignColumnMetadata(ft.ForeignTables, config)...)
	}

	return metadata
}

// NEW: Build relationship metadata
func (s *Store) buildRelationshipMetadata(config *Config) []RelationshipInfo {
	if len(config.DataSource) == 0 {
		return nil
	}

	relationships := make([]RelationshipInfo, 0)
	ds := config.DataSource[0]

	for _, ft := range ds.Select.ForeignTables {
		relationships = append(relationships, s.extractRelationships(ds.Source, ft)...)
	}

	return relationships
}

func (s *Store) extractRelationships(baseTable string, ft ForeignTable) []RelationshipInfo {
	relationships := make([]RelationshipInfo, 0)

	// Parse the relationship
	fromParts := strings.Split(ft.RelationshipFrom, ".")
	toParts := strings.Split(ft.RelationshipTo, ".")

	rel := RelationshipInfo{
		Type: "many-to-one", // Default assumption
	}

	if len(fromParts) == 2 {
		rel.FromTable = fromParts[0]
		rel.FromColumn = fromParts[1]
	} else {
		rel.FromTable = baseTable
		rel.FromColumn = fromParts[0]
	}

	if len(toParts) == 2 {
		rel.ToTable = toParts[0]
		rel.ToColumn = toParts[1]
	} else {
		rel.ToTable = ft.Table
		rel.ToColumn = toParts[0]
	}

	relationships = append(relationships, rel)

	// Recursively process nested relationships
	for _, nested := range ft.ForeignTables {
		relationships = append(relationships, s.extractRelationships(ft.Table, nested)...)
	}

	return relationships
}

// Helper: Returns the key in the data row
func getFieldName(col ColumnDefinition) string {
	if col.Alias != "" {
		return col.Alias // "current_stock"
	}
	return col.Name // "quantity"
}

// Helper: Returns default display name
func getDisplayName(col ColumnDefinition) string {
	if col.Alias != "" {
		return col.Alias // "current_stock"
	}
	return col.Name // "quantity"
}

func inferColumnType(name string) string {
	lower := strings.ToLower(name)

	if strings.HasSuffix(lower, "_id") || lower == "id" {
		return "uuid"
	}
	if strings.Contains(lower, "date") || strings.Contains(lower, "time") {
		return "datetime"
	}
	if strings.Contains(lower, "quantity") || strings.Contains(lower, "count") || strings.Contains(lower, "price") {
		return "number"
	}
	if strings.Contains(lower, "is_") || strings.Contains(lower, "has_") {
		return "boolean"
	}

	return "string"
}

// transformData transforms the data according to configuration
func (s *Store) transformData(data []TableRow, config *Config) []TableRow {
	if len(config.DataSource) == 0 {
		return data
	}

	// Just return clean data - let metadata handle the rest
	return data
}

// Helper functions

func isIDField(key string) bool {
	return key == "id" ||
		endsWith(key, ".id") ||
		endsWith(key, "_id")
}

func extractEntityName(key string) string {
	if key == "id" {
		return "id"
	}

	// For patterns like "table.id" or "table_id"
	if endsWith(key, ".id") {
		parts := splitPath(key)
		if len(parts) >= 2 {
			return parts[len(parts)-2]
		}
	}

	if endsWith(key, "_id") {
		// Remove the "_id" suffix
		return key[:len(key)-3]
	}

	return key
}

func findColumnByKey(config SelectConfig, key string) *ColumnDefinition {
	for i := range config.Columns {
		if config.Columns[i].Alias == key || config.Columns[i].Name == key {
			return &config.Columns[i]
		}
	}

	// Check foreign tables recursively
	for _, ft := range config.ForeignTables {
		if col := findColumnInForeignTableByKey(ft, key); col != nil {
			return col
		}
	}

	return nil
}

func findColumnInForeignTableByKey(ft ForeignTable, key string) *ColumnDefinition {
	for i := range ft.Columns {
		if ft.Columns[i].Alias == key || ft.Columns[i].Name == key {
			return &ft.Columns[i]
		}
	}

	// Check nested foreign tables
	for _, nested := range ft.ForeignTables {
		if col := findColumnInForeignTableByKey(nested, key); col != nil {
			return col
		}
	}

	return nil
}

// Utility functions

func endsWith(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

func splitPath(path string) []string {
	// Simple string split on "."
	result := []string{}
	current := ""
	for _, ch := range path {
		if ch == '.' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

// QueryByPage queries table data with page support
func (s *Store) QueryByPage(ctx context.Context, config *Config, pg page.Page) (*TableData, error) {
	params := QueryParams{
		Page: pg.Number(),
		Rows: pg.RowsPerPage(),
	}

	return s.FetchTableData(ctx, config, params)
}
