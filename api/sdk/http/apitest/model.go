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
	"github.com/timmaaaz/ichor/app/domain/core/contactinfosapp"
	"github.com/timmaaaz/ichor/app/domain/finance/costhistoryapp"
	"github.com/timmaaaz/ichor/app/domain/finance/productcostapp"
	"github.com/timmaaaz/ichor/app/domain/inventory/core/brandapp"
	"github.com/timmaaaz/ichor/app/domain/inventory/core/inventoryitemapp"
	"github.com/timmaaaz/ichor/app/domain/inventory/core/physicalattributeapp"
	"github.com/timmaaaz/ichor/app/domain/inventory/core/productapp"
	"github.com/timmaaaz/ichor/app/domain/inventory/core/productcategoryapp"
	"github.com/timmaaaz/ichor/app/domain/location/cityapp"
	"github.com/timmaaaz/ichor/app/domain/location/officeapp"
	"github.com/timmaaaz/ichor/app/domain/location/streetapp"
	"github.com/timmaaaz/ichor/app/domain/lots/lottrackingsapp"
	"github.com/timmaaaz/ichor/app/domain/lots/serialnumberapp"
	"github.com/timmaaaz/ichor/app/domain/movement/inventoryadjustmentapp"
	"github.com/timmaaaz/ichor/app/domain/movement/inventorytransactionapp"
	"github.com/timmaaaz/ichor/app/domain/movement/transferorderapp"
	"github.com/timmaaaz/ichor/app/domain/order/lineitemfulfillmentstatusapp"
	"github.com/timmaaaz/ichor/app/domain/order/orderfulfillmentstatusapp"
	"github.com/timmaaaz/ichor/app/domain/permissions/roleapp"
	"github.com/timmaaaz/ichor/app/domain/permissions/tableaccessapp"
	"github.com/timmaaaz/ichor/app/domain/permissions/userroleapp.go"
	"github.com/timmaaaz/ichor/app/domain/quality/inspectionapp"
	"github.com/timmaaaz/ichor/app/domain/quality/metricsapp"
	"github.com/timmaaaz/ichor/app/domain/supplier/supplierapp"
	"github.com/timmaaaz/ichor/app/domain/supplier/supplierproductapp"
	"github.com/timmaaaz/ichor/app/domain/users/reportstoapp"
	"github.com/timmaaaz/ichor/app/domain/users/status/approvalapp"
	"github.com/timmaaaz/ichor/app/domain/users/status/commentapp"
	"github.com/timmaaaz/ichor/app/domain/users/titleapp"
	"github.com/timmaaaz/ichor/app/domain/warehouse/inventorylocationapp"
	"github.com/timmaaaz/ichor/app/domain/warehouse/warehouseapp"
	"github.com/timmaaaz/ichor/app/domain/warehouse/zoneapp"
	"github.com/timmaaaz/ichor/business/domain/homebus"
	"github.com/timmaaaz/ichor/business/domain/location/countrybus"
	"github.com/timmaaaz/ichor/business/domain/location/regionbus"

	"github.com/timmaaaz/ichor/business/domain/users/userbus"
)

// User extends the dbtest user for api test support.
type User struct {
	userbus.User

	Homes []homebus.Home
	Token string
}

// SeedData represents users for api tests.
type SeedData struct {
	Users                       []User
	Admins                      []User
	Countries                   []countrybus.Country
	Regions                     []regionbus.Region
	Cities                      []cityapp.City
	Streets                     []streetapp.Street
	ValidAssets                 []validassetapp.ValidAsset
	AssetTypes                  []assettypeapp.AssetType
	AssetConditions             []assetconditionapp.AssetCondition
	ApprovalStatuses            []approvalstatusapp.ApprovalStatus
	UserApprovalStatuses        []approvalapp.UserApprovalStatus
	UserApprovalComments        []commentapp.UserApprovalComment
	FulfillmentStatuses         []fulfillmentstatusapp.FulfillmentStatus
	Tags                        []tagapp.Tag
	AssetTags                   []assettagapp.AssetTag
	Titles                      []titleapp.Title
	ReportsTo                   []reportstoapp.ReportsTo
	Offices                     []officeapp.Office
	UserAssets                  []userassetapp.UserAsset
	Assets                      []assetapp.Asset
	ContactInfos                []contactinfosapp.ContactInfos
	Brands                      []brandapp.Brand
	ProductCategories           []productcategoryapp.ProductCategory
	Warehouses                  []warehouseapp.Warehouse
	Roles                       []roleapp.Role
	UserRoles                   []userroleapp.UserRole
	TableAccesses               []tableaccessapp.TableAccess
	Products                    []productapp.Product
	PhysicalAttributes          []physicalattributeapp.PhysicalAttribute
	ProductCosts                []productcostapp.ProductCost
	Suppliers                   []supplierapp.Supplier
	CostHistory                 []costhistoryapp.CostHistory
	SupplierProducts            []supplierproductapp.SupplierProduct
	Metrics                     []metricsapp.Metric
	LotTrackings                []lottrackingsapp.LotTrackings
	Zones                       []zoneapp.Zone
	InventoryLocations          []inventorylocationapp.InventoryLocation
	InventoryItems              []inventoryitemapp.InventoryItem
	Inspections                 []inspectionapp.Inspection
	SerialNumbers               []serialnumberapp.SerialNumber
	InventoryTransactions       []inventorytransactionapp.InventoryTransaction
	InventoryAdjustments        []inventoryadjustmentapp.InventoryAdjustment
	TransferOrders              []transferorderapp.TransferOrder
	OrderFulfillmentStatuses    []orderfulfillmentstatusapp.OrderFulfillmentStatus
	LineItemFulfillmentStatuses []lineitemfulfillmentstatusapp.LineItemFulfillmentStatus
}

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
