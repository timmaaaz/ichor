package vproductbus_test

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/business/domain/productbus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/domain/vproductbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_VProduct(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_Product")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	unitest.Run(t, query(db.BusDomain, sd), "query")
}

// =============================================================================

func insertSeedData(busDomain dbtest.BusDomain) (unitest.SeedData, error) {
	ctx := context.Background()

	usrs, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.User, busDomain.User)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	prds, err := productbus.TestGenerateSeedProducts(ctx, 2, busDomain.Product, usrs[0].ID)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding products : %w", err)
	}

	tu1 := unitest.User{
		User:     usrs[0],
		Products: prds,
	}

	// -------------------------------------------------------------------------

	usrs, err = userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	prds, err = productbus.TestGenerateSeedProducts(ctx, 2, busDomain.Product, usrs[0].ID)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding products : %w", err)
	}

	tu2 := unitest.User{
		User:     usrs[0],
		Products: prds,
	}

	// -------------------------------------------------------------------------

	sd := unitest.SeedData{
		Admins: []unitest.User{tu2},
		Users:  []unitest.User{tu1},
	}

	return sd, nil
}

// =============================================================================

func toVProduct(usr userbus.User, prd productbus.Product) vproductbus.Product {
	return vproductbus.Product{
		ID:          prd.ID,
		UserID:      prd.UserID,
		Name:        prd.Name,
		Cost:        prd.Cost,
		Quantity:    prd.Quantity,
		DateCreated: prd.DateCreated,
		DateUpdated: prd.DateUpdated,
		UserName:    usr.Username,
	}
}

func toVProducts(usr userbus.User, prds []productbus.Product) []vproductbus.Product {
	items := make([]vproductbus.Product, len(prds))
	for i, prd := range prds {
		items[i] = toVProduct(usr, prd)
	}

	return items
}

// =============================================================================

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	prds := toVProducts(sd.Admins[0].User, sd.Admins[0].Products)
	prds = append(prds, toVProducts(sd.Users[0].User, sd.Users[0].Products)...)

	sort.Slice(prds, func(i, j int) bool {
		return prds[i].ID.String() <= prds[j].ID.String()
	})

	table := []unitest.Table{
		{
			Name:    "all",
			ExpResp: prds,
			ExcFunc: func(ctx context.Context) any {
				filter := vproductbus.QueryFilter{
					Name: dbtest.ProductNamePointer("Name"),
				}

				resp, err := busDomain.VProduct.Query(ctx, filter, vproductbus.DefaultOrderBy, page.MustParse("1", "10"))
				if err != nil {
					return err
				}

				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.([]vproductbus.Product)
				if !exists {
					return "error occurred"
				}

				expResp := exp.([]vproductbus.Product)

				for i := range gotResp {
					if gotResp[i].DateCreated.Format(time.RFC3339) == expResp[i].DateCreated.Format(time.RFC3339) {
						expResp[i].DateCreated = gotResp[i].DateCreated
					}

					if gotResp[i].DateUpdated.Format(time.RFC3339) == expResp[i].DateUpdated.Format(time.RFC3339) {
						expResp[i].DateUpdated = gotResp[i].DateUpdated
					}
				}

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}
