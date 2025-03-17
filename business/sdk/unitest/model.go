package unitest

import (
	"context"

	"github.com/timmaaaz/ichor/business/domain/assets/approvalstatusbus"
	"github.com/timmaaaz/ichor/business/domain/assets/assetbus"
	"github.com/timmaaaz/ichor/business/domain/assets/assetconditionbus"
	"github.com/timmaaaz/ichor/business/domain/assets/assettagbus"
	"github.com/timmaaaz/ichor/business/domain/assets/assettypebus"
	"github.com/timmaaaz/ichor/business/domain/assets/fulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/domain/assets/tagbus"
	"github.com/timmaaaz/ichor/business/domain/assets/userassetbus"
	"github.com/timmaaaz/ichor/business/domain/assets/validassetbus"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfobus"
	"github.com/timmaaaz/ichor/business/domain/homebus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/brandbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/physicalattributebus"
	inventoryproductbus "github.com/timmaaaz/ichor/business/domain/inventory/core/productbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/productcategorybus"
	"github.com/timmaaaz/ichor/business/domain/location/citybus"
	"github.com/timmaaaz/ichor/business/domain/location/countrybus"
	"github.com/timmaaaz/ichor/business/domain/location/officebus"
	"github.com/timmaaaz/ichor/business/domain/location/regionbus"
	"github.com/timmaaaz/ichor/business/domain/location/streetbus"
	"github.com/timmaaaz/ichor/business/domain/permissions/permissionsbus"
	"github.com/timmaaaz/ichor/business/domain/permissions/rolebus"
	"github.com/timmaaaz/ichor/business/domain/permissions/tableaccessbus"
	"github.com/timmaaaz/ichor/business/domain/permissions/userrolebus"
	"github.com/timmaaaz/ichor/business/domain/users/reportstobus"
	"github.com/timmaaaz/ichor/business/domain/users/status/approvalbus"
	"github.com/timmaaaz/ichor/business/domain/users/status/commentbus"
	"github.com/timmaaaz/ichor/business/domain/users/titlebus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/domain/warehouse/warehousebus"
	"github.com/timmaaaz/ichor/business/domain/warehouse/zonebus"
)

// User represents an app user specified for the test.
type User struct {
	userbus.User
	Homes []homebus.Home
}

// SeedData represents data that was seeded for the test.
type SeedData struct {
	Users               []User
	Admins              []User
	AssetConditions     []assetconditionbus.AssetCondition
	ValidAssets         []validassetbus.ValidAsset
	Countries           []countrybus.Country
	Regions             []regionbus.Region
	Cities              []citybus.City
	Streets             []streetbus.Street
	ApprovalStatus      []approvalstatusbus.ApprovalStatus
	UserApprovalStatus  []approvalbus.UserApprovalStatus
	UserApprovalComment []commentbus.UserApprovalComment
	FulfillmentStatus   []fulfillmentstatusbus.FulfillmentStatus
	AssetCondition      []assetconditionbus.AssetCondition
	AssetTypes          []assettypebus.AssetType
	Tags                []tagbus.Tag
	AssetTags           []assettagbus.AssetTag
	Title               []titlebus.Title
	ReportsTo           []reportstobus.ReportsTo
	Offices             []officebus.Office
	UserAssets          []userassetbus.UserAsset
	Assets              []assetbus.Asset
	ContactInfo         []contactinfobus.ContactInfo
	Brands              []brandbus.Brand
	ProductCategories   []productcategorybus.ProductCategory
	Warehouses          []warehousebus.Warehouse
	Zones               []zonebus.Zone
	Roles               []rolebus.Role
	UserRoles           []userrolebus.UserRole
	TableAccesses       []tableaccessbus.TableAccess
	UserPermissions     []permissionsbus.UserPermissions
	InventoryProducts   []inventoryproductbus.Product
	PhysicalAttributes  []physicalattributebus.PhysicalAttribute
}

type Table struct {
	Name    string
	ExpResp any
	ExcFunc func(ctx context.Context) any
	CmpFunc func(got any, exp any) string
}
