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

func toBusRestrictedColumns(dbs []restrictedColumn) []restrictedcolumnbus.RestrictedColumn {
	restrictedColumns := make([]restrictedcolumnbus.RestrictedColumn, len(dbs))
	for i, db := range dbs {
		restrictedColumns[i] = toBusRestrictedColumn(db)
	}
	return restrictedColumns
}
