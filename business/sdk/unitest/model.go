package unitest

import (
	"context"

	"github.com/timmaaaz/ichor/business/domain/approvalstatusbus"
	"github.com/timmaaaz/ichor/business/domain/assetbus"
	"github.com/timmaaaz/ichor/business/domain/assetconditionbus"
	"github.com/timmaaaz/ichor/business/domain/assettagbus"
	"github.com/timmaaaz/ichor/business/domain/assettypebus"
	"github.com/timmaaaz/ichor/business/domain/fulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/domain/homebus"
	"github.com/timmaaaz/ichor/business/domain/location/citybus"
	"github.com/timmaaaz/ichor/business/domain/location/countrybus"
	"github.com/timmaaaz/ichor/business/domain/location/regionbus"
	"github.com/timmaaaz/ichor/business/domain/location/streetbus"
	"github.com/timmaaaz/ichor/business/domain/officebus"
	"github.com/timmaaaz/ichor/business/domain/productbus"
	"github.com/timmaaaz/ichor/business/domain/reportstobus"
	"github.com/timmaaaz/ichor/business/domain/tagbus"
	"github.com/timmaaaz/ichor/business/domain/titlebus"
	"github.com/timmaaaz/ichor/business/domain/userassetbus"
	"github.com/timmaaaz/ichor/business/domain/userbus"
	"github.com/timmaaaz/ichor/business/domain/validassetbus"
)

// User represents an app user specified for the test.
type User struct {
	userbus.User
	Products []productbus.Product
	Homes    []homebus.Home
}

// SeedData represents data that was seeded for the test.
type SeedData struct {
	Users             []User
	Admins            []User
	AssetConditions   []assetconditionbus.AssetCondition
	ValidAssets       []validassetbus.ValidAsset
	Countries         []countrybus.Country
	Regions           []regionbus.Region
	Cities            []citybus.City
	Streets           []streetbus.Street
	ApprovalStatus    []approvalstatusbus.ApprovalStatus
	FulfillmentStatus []fulfillmentstatusbus.FulfillmentStatus
	AssetCondition    []assetconditionbus.AssetCondition
	AssetTypes        []assettypebus.AssetType
	Tags              []tagbus.Tag
	AssetTags         []assettagbus.AssetTag
	Title             []titlebus.Title
	ReportsTo         []reportstobus.ReportsTo
	Offices           []officebus.Office
	UserAssets        []userassetbus.UserAsset
	Assets            []assetbus.Asset
}

type Table struct {
	Name    string
	ExpResp any
	ExcFunc func(ctx context.Context) any
	CmpFunc func(got any, exp any) string
}
