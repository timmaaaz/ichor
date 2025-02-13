package userassetbus

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
)

func TestNewUserAssets(n int, UserIDs, AssetIDs, ApprovedBy, ConditionIDs, ApprovalStatusIDs, FulfillmentStatusIDs uuid.UUIDs) []NewUserAsset {
	newUserAssets := make([]NewUserAsset, n)

	for i := 0; i < n; i++ {
		nua := NewUserAsset{
			UserID:              UserIDs[i%len(UserIDs)],
			AssetID:             AssetIDs[i%len(AssetIDs)],
			ApprovedBy:          ApprovedBy[i%len(ApprovedBy)],
			ApprovalStatusID:    ApprovalStatusIDs[i%len(ApprovalStatusIDs)],
			FulfillmentStatusID: FulfillmentStatusIDs[i%len(FulfillmentStatusIDs)],
			DateReceived:        time.Now().AddDate(0, -i, 0),
			LastMaintenance:     time.Now().AddDate(0, 0, -i*2),
		}
		newUserAssets[i] = nua
	}

	return newUserAssets
}

func TestSeedUserAssets(ctx context.Context, n int, UserIDs, AssetIDs, ApprovedBy, ConditionIDs, ApprovalStatusIDs, FulfillmentStatusIDs uuid.UUIDs, api *Business) ([]UserAsset, error) {
	newUserAssets := TestNewUserAssets(n, UserIDs, AssetIDs, ApprovedBy, ConditionIDs, ApprovalStatusIDs, FulfillmentStatusIDs)

	userAssets := make([]UserAsset, len(newUserAssets))

	for i, nua := range newUserAssets {
		ua, err := api.Create(ctx, nua)
		if err != nil {
			return nil, fmt.Errorf("seeding user asset: idx: %d, : %w", i, err)
		}

		userAssets[i] = ua
	}

	sort.Slice(userAssets, func(i, j int) bool {
		return userAssets[i].ID.String() < userAssets[j].ID.String()
	})
	return userAssets, nil
}
