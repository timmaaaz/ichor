package apitest

import (
	"github.com/timmaaaz/ichor/app/domain/assets/approvalstatusapp"
	"github.com/timmaaaz/ichor/app/domain/assets/assetapp"
	"github.com/timmaaaz/ichor/app/domain/assets/assetconditionapp"
	"github.com/timmaaaz/ichor/app/domain/assets/assettagapp"
	"github.com/timmaaaz/ichor/app/domain/assets/assettypeapp"
	"github.com/timmaaaz/ichor/app/domain/assets/fulfillmentstatusapp"
	"github.com/timmaaaz/ichor/app/domain/assets/tagapp"
	"github.com/timmaaaz/ichor/app/domain/assets/userassetapp"
	"github.com/timmaaaz/ichor/app/domain/assets/validassetapp"
	"github.com/timmaaaz/ichor/app/domain/core/contactinfoapp"
	"github.com/timmaaaz/ichor/app/domain/inventory/core/brandapp"
	"github.com/timmaaaz/ichor/app/domain/inventory/core/productcategoryapp"
	"github.com/timmaaaz/ichor/app/domain/location/cityapp"
	"github.com/timmaaaz/ichor/app/domain/location/officeapp"
	"github.com/timmaaaz/ichor/app/domain/location/streetapp"
	"github.com/timmaaaz/ichor/app/domain/permissions/roleapp"
	"github.com/timmaaaz/ichor/app/domain/users/reportstoapp"
	"github.com/timmaaaz/ichor/app/domain/users/status/approvalapp"
	"github.com/timmaaaz/ichor/app/domain/users/status/commentapp"
	"github.com/timmaaaz/ichor/app/domain/users/titleapp"
	"github.com/timmaaaz/ichor/business/domain/homebus"
	"github.com/timmaaaz/ichor/business/domain/location/countrybus"
	"github.com/timmaaaz/ichor/business/domain/location/regionbus"
	"github.com/timmaaaz/ichor/business/domain/permissions/tableaccessbus"
	"github.com/timmaaaz/ichor/business/domain/permissions/userrolebus"
	"github.com/timmaaaz/ichor/business/domain/productbus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
)

// User extends the dbtest user for api test support.
type User struct {
	userbus.User
	Products []productbus.Product
	Homes    []homebus.Home
	Token    string
}

// SeedData represents users for api tests.
type SeedData struct {
	Users                []User
	Admins               []User
	Countries            []countrybus.Country
	Regions              []regionbus.Region
	Cities               []cityapp.City
	Streets              []streetapp.Street
	ValidAssets          []validassetapp.ValidAsset
	AssetTypes           []assettypeapp.AssetType
	AssetConditions      []assetconditionapp.AssetCondition
	ApprovalStatuses     []approvalstatusapp.ApprovalStatus
	UserApprovalStatuses []approvalapp.UserApprovalStatus
	UserApprovalComments []commentapp.UserApprovalComment
	FulfillmentStatuses  []fulfillmentstatusapp.FulfillmentStatus
	Tags                 []tagapp.Tag
	AssetTags            []assettagapp.AssetTag
	Titles               []titleapp.Title
	ReportsTo            []reportstoapp.ReportsTo
	Offices              []officeapp.Office
	UserAssets           []userassetapp.UserAsset
	Assets               []assetapp.Asset
	ContactInfo          []contactinfoapp.ContactInfo
	Brands               []brandapp.Brand
	ProductCategories    []productcategoryapp.ProductCategory
	Roles                []roleapp.Role
	UserRoles            []userrolebus.UserRole
	TableAccesses        []tableaccessbus.TableAccess
}

// Table represent fields needed for running an api test.
type Table struct {
	Name       string
	URL        string
	Token      string
	Method     string
	StatusCode int
	Input      any
	GotResp    any
	ExpResp    any
	CmpFunc    func(got any, exp any) string
}
