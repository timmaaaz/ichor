package dbtest

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/assets/approvalstatusbus"
	"github.com/timmaaaz/ichor/business/domain/assets/assetbus"
	"github.com/timmaaaz/ichor/business/domain/assets/assetconditionbus"
	"github.com/timmaaaz/ichor/business/domain/assets/assettagbus"
	"github.com/timmaaaz/ichor/business/domain/assets/assettypebus"
	"github.com/timmaaaz/ichor/business/domain/assets/fulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/domain/assets/tagbus"
	"github.com/timmaaaz/ichor/business/domain/assets/userassetbus"
	"github.com/timmaaaz/ichor/business/domain/assets/validassetbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest/seedmodels"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

func seedAssets(ctx context.Context, busDomain BusDomain, foundation FoundationSeed) error {
	assetTypes, err := assettypebus.TestSeedAssetTypes(ctx, 10, busDomain.AssetType)
	if err != nil {
		return fmt.Errorf("seeding asset types : %w", err)
	}
	assetTypeIDs := make([]uuid.UUID, 0, len(assetTypes))
	for _, at := range assetTypes {
		assetTypeIDs = append(assetTypeIDs, at.ID)
	}

	validAssets, err := validassetbus.TestSeedValidAssetsHistorical(ctx, 10, 2, assetTypeIDs, foundation.Admins[0].ID, busDomain.ValidAsset)
	if err != nil {
		return fmt.Errorf("seeding assets : %w", err)
	}
	validAssetIDs := make([]uuid.UUID, 0, len(validAssets))
	for _, va := range validAssets {
		validAssetIDs = append(validAssetIDs, va.ID)
	}

	conditions, err := assetconditionbus.TestSeedAssetConditions(ctx, 8, busDomain.AssetCondition)
	if err != nil {
		return fmt.Errorf("seeding asset conditions : %w", err)
	}

	conditionIDs := make([]uuid.UUID, len(conditions))
	for i, c := range conditions {
		conditionIDs[i] = c.ID
	}

	assets, err := assetbus.TestSeedAssets(ctx, 15, validAssetIDs, conditionIDs, busDomain.Asset)
	if err != nil {
		return fmt.Errorf("seeding assets : %w", err)
	}
	assetIDs := make([]uuid.UUID, 0, len(assets))
	for _, a := range assets {
		assetIDs = append(assetIDs, a.ID)
	}

	// Query approval statuses (already seeded by seed.sql) for the user-asset FK
	// pool. seed.sql owns these rows; seed-frontend must NOT re-Create them —
	// assets.approval_status has no UNIQUE(name), so a second insert silently
	// double-inserts instead of erroring (mirrors the USD-currency lookup in
	// seedFoundation).
	approvalStatusNames := seedmodels.ApprovalStatusNames
	approvalStatuses := make([]approvalstatusbus.ApprovalStatus, 0, len(approvalStatusNames))
	for _, name := range approvalStatusNames {
		statuses, err := busDomain.ApprovalStatus.Query(ctx, approvalstatusbus.QueryFilter{Name: &name}, approvalstatusbus.DefaultOrderBy, page.MustParse("1", "1"))
		if err != nil {
			return fmt.Errorf("querying approval status %s: %w", name, err)
		}
		if len(statuses) == 0 {
			return fmt.Errorf("approval status %s not found - ensure seed.sql has run", name)
		}
		approvalStatuses = append(approvalStatuses, statuses[0])
	}
	approvalStatusIDs := make([]uuid.UUID, len(approvalStatuses))
	for i, as := range approvalStatuses {
		approvalStatusIDs[i] = as.ID
	}

	// Query fulfillment statuses (already seeded by seed.sql) for the user-asset
	// FK pool. Same rationale as approval statuses above — seed.sql owns these
	// rows and assets.fulfillment_status has no UNIQUE(name).
	fulfillmentStatusNames := seedmodels.FulfillmentStatusNames
	fulfillmentStatuses := make([]fulfillmentstatusbus.FulfillmentStatus, 0, len(fulfillmentStatusNames))
	for _, name := range fulfillmentStatusNames {
		statuses, err := busDomain.FulfillmentStatus.Query(ctx, fulfillmentstatusbus.QueryFilter{Name: &name}, fulfillmentstatusbus.DefaultOrderBy, page.MustParse("1", "1"))
		if err != nil {
			return fmt.Errorf("querying fulfillment status %s: %w", name, err)
		}
		if len(statuses) == 0 {
			return fmt.Errorf("fulfillment status %s not found - ensure seed.sql has run", name)
		}
		fulfillmentStatuses = append(fulfillmentStatuses, statuses[0])
	}
	fulfillmentStatusIDs := make([]uuid.UUID, len(fulfillmentStatuses))
	for i, fs := range fulfillmentStatuses {
		fulfillmentStatusIDs[i] = fs.ID
	}

	reporterIDs := make([]uuid.UUID, len(foundation.Reporters))
	for i, r := range foundation.Reporters {
		reporterIDs[i] = r.ID
	}

	_, err = userassetbus.TestSeedUserAssets(ctx, 25, reporterIDs[:15], assetIDs, reporterIDs[15:], conditionIDs, approvalStatusIDs, fulfillmentStatusIDs, busDomain.UserAsset)
	if err != nil {
		return fmt.Errorf("seeding user assets : %w", err)
	}

	tags, err := tagbus.TestSeedTag(ctx, 10, busDomain.Tag)
	if err != nil {
		return fmt.Errorf("seeding approval statues : %w", err)
	}
	tagIDs := make([]uuid.UUID, 0, len(tags))
	for _, t := range tags {
		tagIDs = append(tagIDs, t.ID)
	}

	_, err = assettagbus.TestSeedAssetTag(ctx, 20, validAssetIDs, tagIDs, busDomain.AssetTag)
	if err != nil {
		return fmt.Errorf("seeding asset tags : %w", err)
	}

	return nil
}
