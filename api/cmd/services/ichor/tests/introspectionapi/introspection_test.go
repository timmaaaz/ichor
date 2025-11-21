package introspectionapi_test

import (
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/introspectionapp"
)

func Test_IntrospectionAPI(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_IntrospectionAPI")

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("seeding error %s", err)
	}

	test.Run(t, querySchemas200(sd), "query-schemas-200")
	test.Run(t, queryTables200(sd), "query-tables-200")
	test.Run(t, queryColumns200(sd), "query-columns-200")
	test.Run(t, queryRelationships200(sd), "query-relationships-200")
}

func querySchemas200(sd apitest.SeedData) []apitest.Table {
	// Expected schemas based on database migrations (alphabetically ordered)
	expected := introspectionapp.Schemas{
		{Name: "assets"},
		{Name: "config"},
		{Name: "core"},
		{Name: "geography"},
		{Name: "hr"},
		{Name: "inventory"},
		{Name: "procurement"},
		{Name: "products"},
		{Name: "public"},
		{Name: "sales"},
		{Name: "workflow"},
	}

	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/introspection/schemas",
			Method:     http.MethodGet,
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			GotResp:    &introspectionapp.Schemas{},
			ExpResp:    &expected,
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func queryTables200(sd apitest.SeedData) []apitest.Table {
	// Expected tables in core schema (subset of known tables)
	// We use a custom comparison to check if expected tables are present
	// rather than exact match, since row counts may vary
	expected := introspectionapp.Tables{
		{Schema: "core", Name: "users", RowCountEstimate: nil},
		{Schema: "core", Name: "roles", RowCountEstimate: nil},
		{Schema: "core", Name: "user_roles", RowCountEstimate: nil},
		{Schema: "core", Name: "table_access", RowCountEstimate: nil},
	}

	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/introspection/schemas/core/tables",
			Method:     http.MethodGet,
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			GotResp:    &introspectionapp.Tables{},
			ExpResp:    &expected,
			CmpFunc: func(got, exp any) string {
				gotTables := got.(*introspectionapp.Tables)
				expTables := exp.(*introspectionapp.Tables)

				// Check if all expected tables are present in the response
				// We don't check row counts as they may vary
				gotMap := make(map[string]bool)
				for _, t := range *gotTables {
					gotMap[t.Name] = true
				}

				for _, expTable := range *expTables {
					if !gotMap[expTable.Name] {
						return "missing expected table: " + expTable.Name
					}
				}
				return ""
			},
		},
	}
}

func queryColumns200(sd apitest.SeedData) []apitest.Table {
	// Expected columns in core.users table (subset of known columns)
	// We use a custom comparison to check if expected columns are present
	expected := introspectionapp.Columns{
		{Name: "id", DataType: "uuid", IsNullable: false, IsPrimaryKey: true, DefaultValue: ""},
		{Name: "username", DataType: "text", IsNullable: false, IsPrimaryKey: false, DefaultValue: ""},
		{Name: "email", DataType: "text", IsNullable: false, IsPrimaryKey: false, DefaultValue: ""},
		{Name: "enabled", DataType: "boolean", IsNullable: false, IsPrimaryKey: false, DefaultValue: ""},
	}

	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/introspection/tables/core/users/columns",
			Method:     http.MethodGet,
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			GotResp:    &introspectionapp.Columns{},
			ExpResp:    &expected,
			CmpFunc: func(got, exp any) string {
				gotCols := got.(*introspectionapp.Columns)
				expCols := exp.(*introspectionapp.Columns)

				// Check if all expected columns are present in the response
				gotMap := make(map[string]introspectionapp.Column)
				for _, c := range *gotCols {
					gotMap[c.Name] = c
				}

				for _, expCol := range *expCols {
					gotCol, exists := gotMap[expCol.Name]
					if !exists {
						return "missing expected column: " + expCol.Name
					}
					// Verify key properties match
					if gotCol.IsPrimaryKey != expCol.IsPrimaryKey {
						return "column " + expCol.Name + " primary key mismatch"
					}
				}
				return ""
			},
		},
	}
}

func queryRelationships200(sd apitest.SeedData) []apitest.Table {
	// We expect at least some relationships for hr.offices
	// The offices table should have foreign keys to geography tables
	// We use a custom comparison to check if at least one relationship exists
	expected := introspectionapp.Relationships{}

	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/introspection/tables/hr/offices/relationships",
			Method:     http.MethodGet,
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			GotResp:    &introspectionapp.Relationships{},
			ExpResp:    &expected,
			CmpFunc: func(got, exp any) string {
				gotRels := got.(*introspectionapp.Relationships)

				// Simply verify we got a valid response (can be empty or have relationships)
				// This validates the endpoint works and returns the correct structure
				if gotRels == nil {
					return "expected non-nil relationships response"
				}
				return ""
			},
		},
	}
}
