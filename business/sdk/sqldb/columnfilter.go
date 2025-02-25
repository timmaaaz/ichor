package sqldb

import (
	"strings"

	"github.com/timmaaaz/ichor/foundation/logger"
)

// ColumnFilter provides functionality for filtering column lists
type ColumnFilter struct {
	log *logger.Logger
}

// NewColumnFilter returns a new ColumnFilter
func NewColumnFilter(log *logger.Logger) *ColumnFilter {
	return &ColumnFilter{log: log}
}

// GetFilteredColumns returns column names with specified columns removed
func (cf *ColumnFilter) GetFilteredColumns(allColumns []string, restrictedColumns []string) []string {
	if len(restrictedColumns) == 0 {
		return allColumns
	}

	// Create map for quick lookup
	restrictedMap := make(map[string]struct{})
	for _, rc := range restrictedColumns {
		restrictedMap[rc] = struct{}{}
	}

	// Filter columns
	filteredColumns := make([]string, 0, len(allColumns))
	for _, col := range allColumns {
		if _, restricted := restrictedMap[col]; !restricted {
			filteredColumns = append(filteredColumns, col)
		}
	}

	return filteredColumns
}

// GetColumnString returns a comma-separated string of filtered columns
func (cf *ColumnFilter) GetColumnString(allColumns []string, restrictedColumns []string) string {
	filtered := cf.GetFilteredColumns(allColumns, restrictedColumns)
	return strings.Join(filtered, ", ")
}
