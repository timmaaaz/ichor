package timezonebus_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/business/domain/geography/timezonebus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_Timezone(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_Timezone")

	unitest.Run(t, timezoneQuery(db.BusDomain), "timezone-query")
	unitest.Run(t, timezoneQueryAll(db.BusDomain), "timezone-queryall")
}

func timezoneQuery(busdomain dbtest.BusDomain) []unitest.Table {
	table := []unitest.Table{
		{
			Name:    "Query",
			ExpResp: 5, // Expect 5 results
			ExcFunc: func(ctx context.Context) any {
				tzs, err := busdomain.Timezone.Query(ctx, timezonebus.QueryFilter{}, timezonebus.DefaultOrderBy, page.MustParse("1", "5"))
				if err != nil {
					return err
				}
				return len(tzs)
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func timezoneQueryAll(busdomain dbtest.BusDomain) []unitest.Table {
	table := []unitest.Table{
		{
			Name:    "QueryAll",
			ExpResp: true, // Just verify we get results
			ExcFunc: func(ctx context.Context) any {
				tzs, err := busdomain.Timezone.QueryAll(ctx)
				if err != nil {
					return err
				}
				// Return true if we got timezones
				return len(tzs) > 0
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}
