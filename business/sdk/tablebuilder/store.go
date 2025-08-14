package tablebuilder

import (
	"context"
	"fmt"
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

// FetchTableData executes the table configuration and returns the data
func (s *Store) FetchTableData(ctx context.Context, config *Config, params QueryParams) (*TableData, error) {
	startTime := time.Now()

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	// Initialize result
	result := &TableData{
		Data: make([]TableRow, 0),
		Meta: MetaData{
			Config:   config,
			AliasMap: make(map[string]string),
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
				count, err := s.getCount(ctx, &ds, params)
				if err != nil {
					s.log.Infoc(ctx, 3, "failed to get count", "error", err)
					// Don't fail the whole query if count fails
				} else {
					result.Meta.Total = count
					result.Meta.Page = params.Page
					result.Meta.PageSize = params.Limit
					if params.Limit > 0 {
						result.Meta.TotalPages = (count + params.Limit - 1) / params.Limit
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

	fmt.Println("query: ", query)

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

	// Update alias map
	for _, col := range ds.Select.Columns {
		if col.Alias != "" {
			result.Meta.AliasMap[col.Name] = col.Alias
		}
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

// getCount gets the total count for pagination
func (s *Store) getCount(ctx context.Context, ds *DataSource, params QueryParams) (int, error) {
	query, args, err := s.builder.BuildCountQuery(ds, params)
	if err != nil {
		return 0, fmt.Errorf("build count query: %w", err)
	}

	var count int
	if err := s.db.GetContext(ctx, &count, query, args); err != nil {
		return 0, fmt.Errorf("get count: %w", err)
	}

	return count, nil
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

// transformData transforms the data according to configuration
func (s *Store) transformData(data []TableRow, config *Config) []TableRow {
	if len(config.DataSource) == 0 {
		return data
	}

	transformed := make([]TableRow, 0, len(data))

	for _, row := range data {
		newRow := make(TableRow)

		// Extract IDs
		ids := make(map[string]any)

		// Process each field
		for key, value := range row {
			// Check if this is an ID field
			if isIDField(key) {
				// Extract entity name and store in ids
				entityName := extractEntityName(key)
				ids[entityName] = value
			} else {
				// Regular field - check for table_column mapping
				if col := findColumnByKey(config.DataSource[0].Select, key); col != nil {
					if col.TableColumn != "" {
						fieldData := map[string]any{
							"value":       value,
							"tableColumn": col.TableColumn,
						}

						// Include alias if it exists
						if col.Alias != "" {
							fieldData["alias"] = col.Alias
						}

						// Use alias as the key if it exists, otherwise use the original key
						outputKey := key
						if col.Alias != "" {
							outputKey = col.Alias
						}

						newRow[outputKey] = fieldData
					} else {
						newRow[key] = value
					}
				} else {
					newRow[key] = value
				}
			}
		}

		// Add ids
		if len(ids) > 0 {
			newRow["ids"] = ids
		}

		transformed = append(transformed, newRow)
	}

	return transformed
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
		Page:  pg.Number(),
		Limit: pg.RowsPerPage(),
	}

	return s.FetchTableData(ctx, config, params)
}
