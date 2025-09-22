package assettagbus_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/business/domain/assets/assettagbus"
	"github.com/timmaaaz/ichor/business/domain/assets/assettypebus"
	"github.com/timmaaaz/ichor/business/domain/assets/tagbus"
	"github.com/timmaaaz/ichor/business/domain/assets/validassetbus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_AssetTags(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_AssetTags")
	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	unitest.Run(t, query(db.BusDomain, sd), "query")
	unitest.Run(t, create(db.BusDomain, sd), "create")
	unitest.Run(t, update(db.BusDomain, sd), "update")
	unitest.Run(t, delete(db.BusDomain, sd), "delete")
}

func insertSeedData(busDomain dbtest.BusDomain) (unitest.SeedData, error) {
	ctx := context.Background()

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding user : %w", err)
	}

	// ============= Asset Creation =================
	types, err := assettypebus.TestSeedAssetTypes(ctx, 3, busDomain.AssetType)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding asset types : %w", err)
	}

	typeIDs := make([]uuid.UUID, 0, len(types))
	for _, t := range types {
		typeIDs = append(typeIDs, t.ID)
	}

	assets, err := validassetbus.TestSeedValidAssets(ctx, 10, typeIDs, admins[0].ID, busDomain.ValidAsset)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding assets : %w", err)
	}

	validAssetIDs := make([]uuid.UUID, 0, len(assets))
	for _, a := range assets {
		validAssetIDs = append(validAssetIDs, a.ID)
	}

	// ============= Tag Creation =================

	tags, err := tagbus.TestSeedTag(ctx, 15, busDomain.Tag)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding tags : %w", err)
	}

	tagIDs := make([]uuid.UUID, 0, len(tags))
	for _, t := range tags {
		tagIDs = append(tagIDs, t.ID)
	}

	// ============= Asset-Tag Creation =================

	assetTags, err := assettagbus.TestSeedAssetTag(ctx, 20, validAssetIDs, tagIDs, busDomain.AssetTag)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding asset tags : %w", err)
	}

	return unitest.SeedData{
		Admins:      []unitest.User{{User: admins[0]}},
		ValidAssets: assets,
		Tags:        tags,
		AssetTags:   assetTags,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Query",
			ExpResp: []assettagbus.AssetTag{
				sd.AssetTags[0],
				sd.AssetTags[1],
				sd.AssetTags[2],
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.AssetTag.Query(ctx, assettagbus.QueryFilter{}, assettagbus.DefaultOrderBy, page.MustParse("1", "3"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp interface{}) string {
				gotResp, exists := got.([]assettagbus.AssetTag)
				if !exists {
					return "error occurred"
				}

				expResp, exists := exp.([]assettagbus.AssetTag)
				if !exists {
					return "error occurred"
				}

				return cmp.Diff(expResp, gotResp)
			},
		},
	}
}

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name:    "Create",
			ExpResp: sd.AssetTags[0],
			ExcFunc: func(ctx context.Context) any {
				nat := assettagbus.NewAssetTag{
					ValidAssetID: sd.AssetTags[0].ValidAssetID,
					TagID:        sd.AssetTags[0].TagID,
				}

				got, err := busDomain.AssetTag.Create(ctx, nat)
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp interface{}) string {
				gotResp, exists := got.(assettagbus.AssetTag)
				if !exists {
					return "error occurred"
				}

				expResp, exists := exp.(assettagbus.AssetTag)
				if !exists {
					return "error occurred"
				}

				expResp.ID = gotResp.ID

				return cmp.Diff(expResp, gotResp)
			},
		},
	}
}

func update(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "update",
			ExpResp: assettagbus.AssetTag{
				ID:           sd.AssetTags[0].ID,
				ValidAssetID: sd.AssetTags[1].ValidAssetID,
				TagID:        sd.AssetTags[1].TagID,
			},
			ExcFunc: func(ctx context.Context) any {
				uat := assettagbus.UpdateAssetTag{
					ValidAssetID: &sd.AssetTags[1].ValidAssetID,
					TagID:        &sd.AssetTags[1].TagID,
				}

				got, err := busDomain.AssetTag.Update(ctx, sd.AssetTags[0], uat)
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp interface{}) string {
				gotResp, exists := got.(assettagbus.AssetTag)
				if !exists {
					return "error occurred"
				}

				expResp, exists := exp.(assettagbus.AssetTag)
				if !exists {
					return "error occurred"
				}

				return cmp.Diff(expResp, gotResp)
			},
		},
	}
	return table
}

func delete(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name:    "delete",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				err := busDomain.AssetTag.Delete(ctx, sd.AssetTags[0])
				if err != nil {
					return err
				}
				return nil
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}
