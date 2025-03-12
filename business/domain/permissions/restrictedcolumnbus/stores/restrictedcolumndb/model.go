package restrictedcolumndb

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/permissions/restrictedcolumnbus"
)

type restrictedColumn struct {
	ID         uuid.UUID `db:"restricted_column_id"`
	TableName  string    `db:"table_name"`
	ColumnName string    `db:"column_name"`
}

func toDBRestrictedColumn(bus restrictedcolumnbus.RestrictedColumn) restrictedColumn {
	return restrictedColumn{
		ID:         bus.ID,
		TableName:  bus.TableName,
		ColumnName: bus.ColumnName,
	}
}

func toBusRestrictedColumn(db restrictedColumn) restrictedcolumnbus.RestrictedColumn {
	return restrictedcolumnbus.RestrictedColumn{
		ID:         db.ID,
		TableName:  db.TableName,
		ColumnName: db.ColumnName,
	}
}

func toBusRestrictedColumnSlice(dbs []restrictedColumn) []restrictedcolumnbus.RestrictedColumn {
	restrictedColumns := make([]restrictedcolumnbus.RestrictedColumn, len(dbs))
	for i, db := range dbs {
		restrictedColumns[i] = toBusRestrictedColumn(db)
	}
	return restrictedColumns
}

// NOTE: Not using append here because this will be called all over the
// middleware and we want to avoid the overhead of append.
func toBusRestrictedColumns(dbs []restrictedColumn) restrictedcolumnbus.RestrictedColumns {
	// Pre-calculate map size
	tableCount := 0
	tableSeen := make(map[string]bool)

	for _, db := range dbs {
		if !tableSeen[db.TableName] {
			tableSeen[db.TableName] = true
			tableCount++
		}
	}

	// Pre-allocate the map
	rcs := make(map[string][]string, tableCount)

	// Pre-allocate slices
	columnCountByTable := make(map[string]int)
	for _, db := range dbs {
		columnCountByTable[db.TableName]++
	}

	for tableName, count := range columnCountByTable {
		rcs[tableName] = make([]string, 0, count)
	}

	// Fill the map
	for _, db := range dbs {
		rcs[db.TableName] = append(rcs[db.TableName], db.ColumnName)
	}

	return rcs
}
