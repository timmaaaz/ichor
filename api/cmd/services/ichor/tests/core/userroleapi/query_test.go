package userrole_test

import (
	"net/http"
	"sort"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/core/userroleapp"
	"github.com/timmaaaz/ichor/app/sdk/query"
)

func query200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/core/user-roles?page=1&rows=10",
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &query.Result[userroleapp.UserRole]{},
			ExpResp: &query.Result[userroleapp.UserRole]{
				Page:        1,
				RowsPerPage: 10,
				Total:       4, // 2 from test seed + 2 from SQL seed (admin + floor_worker1) = 4 total
				Items: func() []userroleapp.UserRole {
					// Create a slice with both test data and the SQL-seeded records
					allUserRoles := make([]userroleapp.UserRole, 0, len(sd.UserRoles)+2)

					// Add test-seeded user roles (should be 2 of them)
					allUserRoles = append(allUserRoles, sd.UserRoles...)

					// Add the SQL-seeded user roles (IDs handled in CmpFunc)
					allUserRoles = append(allUserRoles, userroleapp.UserRole{
						UserID: "5cf37266-3473-4006-984f-9325122678b7",
						RoleID: "54bb2165-71e1-41a6-af3e-7da4a0e1e2c1",
					})
					allUserRoles = append(allUserRoles, userroleapp.UserRole{
						UserID: "c0000000-0000-4000-8000-000000000001",
						RoleID: "b0000000-0000-4000-8000-000000000001",
					})

					// Sort by a predictable field since IDs might be random
					sort.Slice(allUserRoles, func(i, j int) bool {
						// Sort by UserID first, then RoleID for deterministic ordering
						if allUserRoles[i].UserID != allUserRoles[j].UserID {
							return allUserRoles[i].UserID < allUserRoles[j].UserID
						}
						return allUserRoles[i].RoleID < allUserRoles[j].RoleID
					})

					return allUserRoles
				}(),
			},
			CmpFunc: func(got any, exp any) string {
				gotResult := got.(*query.Result[userroleapp.UserRole])
				expResult := exp.(*query.Result[userroleapp.UserRole])

				// Sort both by the same criteria for consistent comparison
				sortFunc := func(i, j int, items []userroleapp.UserRole) bool {
					if items[i].UserID != items[j].UserID {
						return items[i].UserID < items[j].UserID
					}
					return items[i].RoleID < items[j].RoleID
				}

				sort.Slice(gotResult.Items, func(i, j int) bool {
					return sortFunc(i, j, gotResult.Items)
				})

				sort.Slice(expResult.Items, func(i, j int) bool {
					return sortFunc(i, j, expResult.Items)
				})

				// SQL-seeded records may have random IDs — copy actual IDs to expected.
				type sqlSeedKey struct{ userID, roleID string }
				sqlSeeds := []sqlSeedKey{
					{"5cf37266-3473-4006-984f-9325122678b7", "54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"},
					{"c0000000-0000-4000-8000-000000000001", "b0000000-0000-4000-8000-000000000001"},
				}
				for _, seed := range sqlSeeds {
					for i, item := range gotResult.Items {
						if item.UserID == seed.userID && item.RoleID == seed.roleID {
							for j, expItem := range expResult.Items {
								if expItem.UserID == seed.userID && expItem.RoleID == seed.roleID {
									expResult.Items[j].ID = gotResult.Items[i].ID
									break
								}
							}
							break
						}
					}
				}

				return cmp.Diff(gotResult, expResult)
			},
		},
	}
	return table
}

func queryByID200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/core/user-roles/" + sd.UserRoles[0].ID,
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &userroleapp.UserRole{},
			ExpResp:    &sd.UserRoles[0],
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func query401(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/core/user-roles?page=1&rows=10",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusForbidden,
			Method:     http.MethodGet,
			GotResp:    &query.Result[userroleapp.UserRole]{},
			ExpResp:    &query.Result[userroleapp.UserRole]{},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}
