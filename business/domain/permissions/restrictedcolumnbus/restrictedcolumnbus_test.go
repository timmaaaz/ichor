package restrictedcolumnbus_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/business/domain/permissions/restrictedcolumnbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_RestrictedColumn(t *testing.T) {
	db := dbtest.NewDatabase(t, "Test_RestrictedColumn")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}
	unitest.Run(t, query(db.BusDomain, sd), "query")
	unitest.Run(t, create(db.BusDomain), "create")
	unitest.Run(t, delete(db.BusDomain, sd), "delete")
}

func insertSeedData(busDomain dbtest.BusDomain) (unitest.SeedData, error) {
	ctx := context.Background()

	restrictedColumns, err := restrictedcolumnbus.TestSeedRestrictedColumns(ctx, busDomain.RestrictedColumn)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding restricted columns : %w", err)
	}

	return unitest.SeedData{
		RestrictedColumns: restrictedColumns,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Query",
			ExpResp: []restrictedcolumnbus.RestrictedColumn{
				sd.RestrictedColumns[0],
				sd.RestrictedColumns[1],
				sd.RestrictedColumns[2],
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.RestrictedColumn.Query(ctx, restrictedcolumnbus.QueryFilter{}, restrictedcolumnbus.DefaultOrderBy, page.MustParse("1", "3"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.([]restrictedcolumnbus.RestrictedColumn)
				if !exists {
					return "error occurred"
				}

				return cmp.Diff(exp.([]restrictedcolumnbus.RestrictedColumn), gotResp)
			},
		},
	}
}

func create(busDomain dbtest.BusDomain) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Create",
			ExpResp: restrictedcolumnbus.RestrictedColumn{
				TableName:  "valid_assets",
				ColumnName: "serial_number",
			},
			ExcFunc: func(ctx context.Context) any {
				resp, err := busDomain.RestrictedColumn.Create(ctx, restrictedcolumnbus.NewRestrictedColumn{
					TableName:  "valid_assets",
					ColumnName: "serial_number",
				})

				if err != nil {
					return err
				}

				return resp
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(restrictedcolumnbus.RestrictedColumn)
				if !exists {
					return "error occurred"
				}

				expResp, exists := exp.(restrictedcolumnbus.RestrictedColumn)
				if !exists {
					return "error occurred"
				}
				expResp.ID = gotResp.ID

				return cmp.Diff(expResp, gotResp)
			},
		},
	}
}

func delete(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name:    "Delete",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				return busDomain.RestrictedColumn.Delete(ctx, sd.RestrictedColumns[0])
			},
			CmpFunc: func(got, exp any) string {
				if got != nil {
					return "error occurred"
				}
				return ""
			},
		},
	}
}
