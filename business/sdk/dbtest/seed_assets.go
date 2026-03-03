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

	// Create approval statuses matching seed.sql
	approvalStatusNames := seedmodels.ApprovalStatusNames
	approvalStatuses := make([]approvalstatusbus.ApprovalStatus, 0, len(approvalStatusNames))
	for _, name := range approvalStatusNames {
		as, err := busDomain.ApprovalStatus.Create(ctx, approvalstatusbus.NewApprovalStatus{
			IconID: uuid.New(),
			Name:   name,
		})
		if err != nil {
			return fmt.Errorf("seeding approval status %s: %w", name, err)
		}
		approvalStatuses = append(approvalStatuses, as)
	}
	approvalStatusIDs := make([]uuid.UUID, len(approvalStatuses))
	for i, as := range approvalStatuses {
		approvalStatusIDs[i] = as.ID
	}

	// Create fulfillment statuses matching seed.sql
	fulfillmentStatusNames := seedmodels.FulfillmentStatusNames
	fulfillmentStatuses := make([]fulfillmentstatusbus.FulfillmentStatus, 0, len(fulfillmentStatusNames))
	for _, name := range fulfillmentStatusNames {
		fs, err := busDomain.FulfillmentStatus.Create(ctx, fulfillmentstatusbus.NewFulfillmentStatus{
			IconID: uuid.New(),
			Name:   name,
		})
		if err != nil {
			return fmt.Errorf("seeding fulfillment status %s: %w", name, err)
		}
		fulfillmentStatuses = append(fulfillmentStatuses, fs)
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
