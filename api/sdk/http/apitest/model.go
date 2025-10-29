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
	"github.com/timmaaaz/ichor/app/domain/config/formapp"
	"github.com/timmaaaz/ichor/app/domain/config/formfieldapp"
	"github.com/timmaaaz/ichor/app/domain/core/contactinfosapp"
	"github.com/timmaaaz/ichor/app/domain/core/pageapp"
	"github.com/timmaaaz/ichor/app/domain/core/roleapp"
	"github.com/timmaaaz/ichor/app/domain/core/rolepageapp"
	"github.com/timmaaaz/ichor/app/domain/core/tableaccessapp"
	"github.com/timmaaaz/ichor/app/domain/core/userroleapp"
	"github.com/timmaaaz/ichor/app/domain/geography/cityapp"
	"github.com/timmaaaz/ichor/app/domain/geography/streetapp"
	"github.com/timmaaaz/ichor/app/domain/hr/approvalapp"
	"github.com/timmaaaz/ichor/app/domain/hr/commentapp"
	"github.com/timmaaaz/ichor/app/domain/hr/officeapp"
	"github.com/timmaaaz/ichor/app/domain/hr/reportstoapp"
	"github.com/timmaaaz/ichor/app/domain/hr/titleapp"
	"github.com/timmaaaz/ichor/app/domain/inventory/inspectionapp"
	"github.com/timmaaaz/ichor/app/domain/inventory/inventoryadjustmentapp"
	"github.com/timmaaaz/ichor/app/domain/inventory/inventoryitemapp"
	"github.com/timmaaaz/ichor/app/domain/inventory/inventorylocationapp"
	"github.com/timmaaaz/ichor/app/domain/inventory/inventorytransactionapp"
	"github.com/timmaaaz/ichor/app/domain/inventory/lottrackingsapp"
	"github.com/timmaaaz/ichor/app/domain/inventory/serialnumberapp"
	"github.com/timmaaaz/ichor/app/domain/inventory/transferorderapp"
	"github.com/timmaaaz/ichor/app/domain/inventory/warehouseapp"
	"github.com/timmaaaz/ichor/app/domain/inventory/zoneapp"
	"github.com/timmaaaz/ichor/app/domain/procurement/purchaseorderapp"
	"github.com/timmaaaz/ichor/app/domain/procurement/purchaseorderlineitemapp"
	"github.com/timmaaaz/ichor/app/domain/procurement/purchaseorderlineitemstatusapp"
	"github.com/timmaaaz/ichor/app/domain/procurement/purchaseorderstatusapp"
	"github.com/timmaaaz/ichor/app/domain/procurement/supplierapp"
	"github.com/timmaaaz/ichor/app/domain/procurement/supplierproductapp"
	"github.com/timmaaaz/ichor/app/domain/products/brandapp"
	"github.com/timmaaaz/ichor/app/domain/products/costhistoryapp"
	"github.com/timmaaaz/ichor/app/domain/products/metricsapp"
	"github.com/timmaaaz/ichor/app/domain/products/physicalattributeapp"
	"github.com/timmaaaz/ichor/app/domain/products/productapp"
	"github.com/timmaaaz/ichor/app/domain/products/productcategoryapp"
	"github.com/timmaaaz/ichor/app/domain/products/productcostapp"
	"github.com/timmaaaz/ichor/app/domain/sales/customersapp"
	"github.com/timmaaaz/ichor/app/domain/sales/lineitemfulfillmentstatusapp"
	"github.com/timmaaaz/ichor/app/domain/sales/orderfulfillmentstatusapp"
	"github.com/timmaaaz/ichor/app/domain/sales/orderlineitemsapp"
	"github.com/timmaaaz/ichor/app/domain/sales/ordersapp"
	"github.com/timmaaaz/ichor/business/domain/geography/countrybus"
	"github.com/timmaaaz/ichor/business/domain/geography/regionbus"
	"github.com/timmaaaz/ichor/business/domain/hr/homebus"
	"github.com/timmaaaz/ichor/business/sdk/tablebuilder"
	"github.com/timmaaaz/ichor/business/sdk/workflow"

	"github.com/timmaaaz/ichor/business/domain/core/userbus"
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
	Customers                   []customersapp.Customers
	Brands                      []brandapp.Brand
	ProductCategories           []productcategoryapp.ProductCategory
	Warehouses                  []warehouseapp.Warehouse
	Roles                       []roleapp.Role
	Pages                       []pageapp.Page
	RolePages                   []rolepageapp.RolePage
	UserRoles                   []userroleapp.UserRole
	TableAccesses               []tableaccessapp.TableAccess
	Products                    []productapp.Product
	PhysicalAttributes          []physicalattributeapp.PhysicalAttribute
	ProductCosts                []productcostapp.ProductCost
	PurchaseOrderLineItemStatuses []purchaseorderlineitemstatusapp.PurchaseOrderLineItemStatus
	PurchaseOrderStatuses       []purchaseorderstatusapp.PurchaseOrderStatus
	PurchaseOrders              []purchaseorderapp.PurchaseOrder
	PurchaseOrderLineItems      []purchaseorderlineitemapp.PurchaseOrderLineItem
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
	Orders                      []ordersapp.Order
	OrderLineItems              []orderlineitemsapp.OrderLineItem
	SimpleTableConfig           *tablebuilder.StoredConfig
	ComplexTableConfig          *tablebuilder.StoredConfig
	PageTableConfig             *tablebuilder.StoredConfig
	PageConfigs                 []tablebuilder.PageConfig
	PageTabConfigs              []tablebuilder.PageTabConfig
	Forms                       []formapp.Form
	FormFields                  []formfieldapp.FormField
	Entities                    []workflow.Entity
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
