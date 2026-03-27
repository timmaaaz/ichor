# Test Failure: Test_InventoryItem/query-200-basic

- **Package**: `github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/inventory/inventoryitemapi`
- **Duration**: 0.01s

## Failure Output

```
b2-e696-459d-985e-c16c027c9156","quantity":"114","reserved_quantity":"0","allocated_quantity":"0","minimum_stock":"10","maximum_stock":"228","reorder_point":"20","economic_order_quantity":"50","safety_stock":"15","avg_daily_usage":"5","created_date":"2026-03-13 17:27:50 +0000 UTC","updated_date":"2026-03-13 17:27:50 +0000 UTC"},{"id":"1a4087e6-c9ab-4bd1-b206-e29e11083b6d","product_id":"833bd7a6-240a-4d2b-ab43-55ea5ab70f98","location_id":"821bfaca-104c-4025-9dda-ec90789359b9","quantity":"137","reserved_quantity":"0","allocated_quantity":"0","minimum_stock":"10","maximum_stock":"274","reorder_point":"20","economic_order_quantity":"50","safety_stock":"15","avg_daily_usage":"5","created_date":"2026-03-13 17:27:50 +0000 UTC","updated_date":"2026-03-13 17:27:50 +0000 UTC"},{"id":"226a6328-4725-4afb-9a05-fdeca67a7162","product_id":"85f5a8e9-772f-40e9-b729-92d061edc3a5","location_id":"5b945402-7fde-4e96-8e09-e7c7afab2c02","quantity":"107","reserved_quantity":"0","allocated_quantity":"0","minimum_stock":"10","maximum_
stock":"214","reorder_point":"20","economic_order_quantity":"50","safety_stock":"15","avg_daily_usage":"5","created_date":"2026-03-13 17:27:50 +0000 UTC","updated_date":"2026-03-13 17:27:50 +0000 UTC"},{"id":"2343d3e6-7587-44f4-a422-9aa90ca6e752","product_id":"833bd7a6-240a-4d2b-ab43-55ea5ab70f98","location_id":"10001e7a-f176-4102-a144-89bffbd66dbb","quantity":"125","reserved_quantity":"0","allocated_quantity":"0","minimum_stock":"10","maximum_stock":"250","reorder_point":"20","economic_order_quantity":"50","safety_stock":"15","avg_daily_usage":"5","created_date":"2026-03-13 17:27:50 +0000 UTC","updated_date":"2026-03-13 17:27:50 +0000 UTC"},{"id":"273c8c21-bc5e-47ce-ad5b-294b0f2ada75","product_id":"833bd7a6-240a-4d2b-ab43-55ea5ab70f98","location_id":"5b945402-7fde-4e96-8e09-e7c7afab2c02","quantity":"132","reserved_quantity":"0","allocated_quantity":"0","minimum_stock":"10","maximum_stock":"264","reorder_point":"20","economic_order_quantity":"50","safety_stock":"15","avg_daily_usage":"5","created_date":"2026-
03-13 17:27:50 +0000 UTC","updated_date":"2026-03-13 17:27:50 +0000 UTC"},{"id":"28786329-04ee-46a0-baed-e33f6369d433","product_id":"833bd7a6-240a-4d2b-ab43-55ea5ab70f98","location_id":"ca5d921b-a577-40ac-b628-fb8c30fc3fd4","quantity":"144","reserved_quantity":"0","allocated_quantity":"0","minimum_stock":"10","maximum_stock":"288","reorder_point":"20","economic_order_quantity":"50","safety_stock":"15","avg_daily_usage":"5","created_date":"2026-03-13 17:27:50 +0000 UTC","updated_date":"2026-03-13 17:27:50 +0000 UTC"},{"id":"2ce870f6-2bc2-47a4-8c3d-8c32985dd0f5","product_id":"833bd7a6-240a-4d2b-ab43-55ea5ab70f98","location_id":"1b8d403c-0d68-46d7-ad1c-999ddb1021b6","quantity":"126","reserved_quantity":"0","allocated_quantity":"0","minimum_stock":"10","maximum_stock":"252","reorder_point":"20","economic_order_quantity":"50","safety_stock":"15","avg_daily_usage":"5","created_date":"2026-03-13 17:27:50 +0000 UTC","updated_date":"2026-03-13 17:27:50 +0000 UTC"},{"id":"33a5aa75-50a1-4485-bbc5-cb61a2520428","product_
id":"833bd7a6-240a-4d2b-ab43-55ea5ab70f98","location_id":"e46fed35-ac5a-4b49-ad4b-8a6ec23295ba","quantity":"147","reserved_quantity":"0","allocated_quantity":"0","minimum_stock":"10","maximum_stock":"294","reorder_point":"20","economic_order_quantity":"50","safety_stock":"15","avg_daily_usage":"5","created_date":"2026-03-13 17:27:50 +0000 UTC","updated_date":"2026-03-13 17:27:50 +0000 UTC"}],"total":50,"page":1,"rows_per_page":10}
    apitest.go:73: DIFF
    apitest.go:74:   &query.Result[github.com/timmaaaz/ichor/app/domain/inventory/inventoryitemapp.InventoryItem]{
          	Items: []inventoryitemapp.InventoryItem{
        - 		{
        - 			ID:                    "029a59da-b09f-48d1-8907-5e9078a9246c",
        - 			ProductID:             "833bd7a6-240a-4d2b-ab43-55ea5ab70f98",
        - 			LocationID:            "ddfff789-3bf3-4941-a49b-5e449332f4d1",
        - 			Quantity:              "146",
        - 			ReservedQuantity:      "0",
        - 			AllocatedQuantity:     "0",
        - 			MinimumStock:          "10",
        - 			MaximumStock:          "292",
        - 			ReorderPoint:          "20",
        - 			EconomicOrderQuantity: "50",
        - 			SafetyStock:           "15",
        - 			AvgDailyUsage:         "5",
        - 			CreatedDate:           "2026-03-13 17:27:50 +0000 UTC",
        - 			UpdatedDate:           "2026-03-13 17:27:50 +0000 UTC",
        - 		},
        - 		{
        - 			ID:                    "117942da-b134-48fd-ad7b-7589d7c587b4",
        - 			ProductID:             "85f5a8e9-772f-40e9-b729-92d061edc3a5",
        - 			LocationID:            "ddfff789-3bf3-4941-a49b-5e449332f4d1",
        - 			Quantity:              "121",
        - 			ReservedQuantity:      "0",
        - 			AllocatedQuantity:     "0",
        - 			MinimumStock:          "10",
        - 			MaximumStock:          "242",
        - 			ReorderPoint:          "20",
        - 			EconomicOrderQuantity: "50",
        - 			SafetyStock:           "15",
        - 			AvgDailyUsage:         "5",
        - 			CreatedDate:           "2026-03-13 17:27:50 +0000 UTC",
        - 			UpdatedDate:           "2026-03-13 17:27:50 +0000 UTC",
        - 		},
        - 		{
        - 			ID:                    "13beffac-fa04-4bb7-ad77-2d36decf92f0",
        - 			ProductID:             "85f5a8e9-772f-40e9-b729-92d061edc3a5",
        - 			LocationID:            "a862aeb2-e696-459d-985e-c16c027c9156",
        - 			Quantity:              "114",
        - 			ReservedQuantity:      "0",
        - 			AllocatedQuantity:     "0",
        - 			MinimumStock:          "10",
        - 			MaximumStock:          "228",
        - 			ReorderPoint:          "20",
        - 			EconomicOrderQuantity: "50",
        - 			SafetyStock:           "15",
        - 			AvgDailyUsage:         "5",
        - 			CreatedDate:           "2026-03-13 17:27:50 +0000 UTC",
        - 			UpdatedDate:           "2026-03-13 17:27:50 +0000 UTC",
        - 		},
        - 		{
        - 			ID:                    "1a4087e6-c9ab-4bd1-b206-e29e11083b6d",
        - 			ProductID:             "833bd7a6-240a-4d2b-ab43-55ea5ab70f98",
        - 			LocationID:            "821bfaca-104c-4025-9dda-ec90789359b9",
        - 			Quantity:              "137",
        - 			ReservedQuantity:      "0",
        - 			AllocatedQuantity:     "0",
        - 			MinimumStock:          "10",
        - 			MaximumStock:          "274",
        - 			ReorderPoint:          "20",
        - 			EconomicOrderQuantity: "50",
        - 			SafetyStock:           "15",
        - 			AvgDailyUsage:         "5",
        - 			CreatedDate:           "2026-03-13 17:27:50 +0000 UTC",
        - 			UpdatedDate:           "2026-03-13 17:27:50 +0000 UTC",
        - 		},
        - 		{
        - 			ID:                    "226a6328-4725-4afb-9a05-fdeca67a7162",
        - 			ProductID:             "85f5a8e9-772f-40e9-b729-92d061edc3a5",
        - 			LocationID:            "5b945402-7fde-4e96-8e09-e7c7afab2c02",
        - 			Quantity:              "107",
        - 			ReservedQuantity:      "0",
        - 			AllocatedQuantity:     "0",
        - 			MinimumStock:          "10",
        - 			MaximumStock:          "214",
        - 			ReorderPoint:          "20",
        - 			EconomicOrderQuantity: "50",
        - 			SafetyStock:           "15",
        - 			AvgDailyUsage:         "5",
        - 			CreatedDate:           "2026-03-13 17:27:50 +0000 UTC",
        - 			UpdatedDate:           "2026-03-13 17:27:50 +0000 UTC",
        - 		},
        + 		{
        + 			ID:                    "2ce870f6-2bc2-47a4-8c3d-8c32985dd0f5",
        + 			ProductID:             "833bd7a6-240a-4d2b-ab43-55ea5ab70f98",
        + 			LocationID:            "1b8d403c-0d68-46d7-ad1c-999ddb1021b6",
        + 			Quantity:              "126",
        + 			ReservedQuantity:      "0",
        + 			AllocatedQuantity:     "0",
        + 			MinimumStock:          "10",
        + 			MaximumStock:          "252",
        + 			ReorderPoint:          "20",
        + 			EconomicOrderQuantity: "50",
        + 			SafetyStock:           "15",
        + 			AvgDailyUsage:         "5",
        + 			CreatedDate:           "2026-03-13 17:27:50 +0000 UTC",
        + 			UpdatedDate:           "2026-03-13 17:27:50 +0000 UTC",
        + 		},
        + 		{
        + 			ID:                    "d08b2c76-f924-49d8-95b5-151c35c2c83f",
        + 			ProductID:             "833bd7a6-240a-4d2b-ab43-55ea5ab70f98",
        + 			LocationID:            "267cc967-2596-49f3-8838-1b5146922959",
        + 			Quantity:              "127",
        + 			ReservedQuantity:      "0",
        + 			AllocatedQuantity:     "0",
        + 			MinimumStock:          "10",
        + 			MaximumStock:          "254",
        + 			ReorderPoint:          "20",
        + 			EconomicOrderQuantity: "50",
        + 			SafetyStock:           "15",
        + 			AvgDailyUsage:         "5",
        + 			CreatedDate:           "2026-03-13 17:27:50 +0000 UTC",
        + 			UpdatedDate:           "2026-03-13 17:27:50 +0000 UTC",
        + 		},
        + 		{
        + 			ID:                    "fc572eae-594f-429e-9527-df0a951fc036",
        + 			ProductID:             "833bd7a6-240a-4d2b-ab43-55ea5ab70f98",
        + 			LocationID:            "297ce289-0b50-4568-ae1b-a12ccc09f3b0",
        + 			Quantity:              "128",
        + 			ReservedQuantity:      "0",
        + 			AllocatedQuantity:     "0",
        + 			MinimumStock:          "10",
        + 			MaximumStock:          "256",
        + 			ReorderPoint:          "20",
        + 			EconomicOrderQuantity: "50",
        + 			SafetyStock:           "15",
        + 			AvgDailyUsage:         "5",
        + 			CreatedDate:           "2026-03-13 17:27:50 +0000 UTC",
        + 			UpdatedDate:           "2026-03-13 17:27:50 +0000 UTC",
        + 		},
        + 		{
        + 			ID:                    "60317aa2-44e4-42fa-978a-55821c02a851",
        + 			ProductID:             "833bd7a6-240a-4d2b-ab43-55ea5ab70f98",
        + 			LocationID:            "38373eaf-7f55-4d11-beea-2cd58e0eba15",
        + 			Quantity:              "129",
        + 			ReservedQuantity:      "0",
        + 			AllocatedQuantity:     "0",
        + 			MinimumStock:          "10",
        + 			MaximumStock:          "258",
        + 			ReorderPoint:          "20",
        + 			EconomicOrderQuantity: "50",
        + 			SafetyStock:           "15",
        + 			AvgDailyUsage:         "5",
        + 			CreatedDate:           "2026-03-13 17:27:50 +0000 UTC",
        + 			UpdatedDate:           "2026-03-13 17:27:50 +0000 UTC",
        + 		},
        + 		{
        + 			ID:                    "7917346a-4ec5-44f0-b789-dc4cc0c693c7",
        + 			ProductID:             "833bd7a6-240a-4d2b-ab43-55ea5ab70f98",
        + 			LocationID:            "480da296-e69d-48be-894a-11cf06b3be3b",
        + 			Quantity:              "130",
        + 			ReservedQuantity:      "0",
        + 			AllocatedQuantity:     "0",
        + 			MinimumStock:          "10",
        + 			MaximumStock:          "260",
        + 			ReorderPoint:          "20",
        + 			EconomicOrderQuantity: "50",
        + 			SafetyStock:           "15",
        + 			AvgDailyUsage:         "5",
        + 			CreatedDate:           "2026-03-13 17:27:50 +0000 UTC",
        + 			UpdatedDate:           "2026-03-13 17:27:50 +0000 UTC",
        + 		},
        + 		{
        + 			ID:                    "e9dbef72-487d-4e6c-bb8c-3d31dc83e2c9",
        + 			ProductID:             "833bd7a6-240a-4d2b-ab43-55ea5ab70f98",
        + 			LocationID:            "50050b00-bdc7-4710-ac27-302df535ad21",
        + 			Quantity:              "131",
        + 			ReservedQuantity:      "0",
        + 			AllocatedQuantity:     "0",
        + 			MinimumStock:          "10",
        + 			MaximumStock:          "262",
        + 			ReorderPoint:          "20",
        + 			EconomicOrderQuantity: "50",
        + 			SafetyStock:           "15",
        + 			AvgDailyUsage:         "5",
        + 			CreatedDate:           "2026-03-13 17:27:50 +0000 UTC",
        + 			UpdatedDate:           "2026-03-13 17:27:50 +0000 UTC",
        + 		},
        - 		{
        - 			ID:                    "28786329-04ee-46a0-baed-e33f6369d433",
        - 			ProductID:             "833bd7a6-240a-4d2b-ab43-55ea5ab70f98",
        - 			LocationID:            "ca5d921b-a577-40ac-b628-fb8c30fc3fd4",
        - 			Quantity:              "144",
        - 			ReservedQuantity:      "0",
        - 			AllocatedQuantity:     "0",
        - 			MinimumStock:          "10",
        - 			MaximumStock:          "288",
        - 			ReorderPoint:          "20",
        - 			EconomicOrderQuantity: "50",
        - 			SafetyStock:           "15",
        - 			AvgDailyUsage:         "5",
        - 			CreatedDate:           "2026-03-13 17:27:50 +0000 UTC",
        - 			UpdatedDate:           "2026-03-13 17:27:50 +0000 UTC",
        - 		},
          			ID: strings.Join({
        - 				"2ce",
          				"8",
        - 				"70",
          				"f",
        - 				"6-2bc2-47a4-8c3d-8c32985dd0f5",
        + 				"b95b12-1e53-4117-99f6-e3be9d8cc3aa",
          			}, ""),
          			ProductID: "833bd7a6-240a-4d2b-ab43-55ea5ab70f98",
          			LocationID: strings.Join({
        - 				"1b8d403c-0d68-46d7-ad1c-999ddb1021b6",
        + 				"614c0afb-1bd8-4350-aa4e-2b085b609764",
          			}, ""),
        - 			Quantity:              "126",
        + 			Quantity:              "133",
          			ReservedQuantity:      "0",
          			AllocatedQuantity:     "0",
          			MinimumStock:          "10",
        - 			MaximumStock:          "252",
        + 			MaximumStock:          "266",
          			ReorderPoint:          "20",
          			EconomicOrderQuantity: "50",
          			... // 4 identical fields
          		},
          			ID: strings.Join({
        - 				"33a5aa75-50a1-4485-bbc5-cb61a2520428",
        + 				"46d8976e-ad19-4bce-8d44-ee5330c31a79",
          			}, ""),
          			ProductID: "833bd7a6-240a-4d2b-ab43-55ea5ab70f98",
          			LocationID: strings.Join({
        - 				"e46fed35-ac5a-4b49-ad4b-8a6ec23295ba",
        + 				"678e86c5-6448-4782-bbd4-5c5cb04d1dc2",
          			}, ""),
        - 			Quantity:              "147",
        + 			Quantity:              "134",
          			ReservedQuantity:      "0",
          			AllocatedQuantity:     "0",
          			MinimumStock:          "10",
        - 			MaximumStock:          "294",
        + 			MaximumStock:          "268",
          			ReorderPoint:          "20",
          			EconomicOrderQuantity: "50",
          			... // 4 identical fields
          		},
          	},
          	Total:       50,
          	Page:        1,
          	RowsPerPage: 10,
          }
    apitest.go:75: GOT
    apitest.go:76: &query.Result[github.com/timmaaaz/ichor/app/domain/inventory/inventoryitemapp.InventoryItem]{Items:[]inventoryitemapp.InventoryItem{inventoryitemapp.InventoryItem{ID:"029a59da-b09f-48d1-8907-5e9078a9246c", ProductID:"833bd7a6-240a-4d2b-ab43-55ea5ab70f98", LocationID:"ddfff789-3bf3-4941-a49b-5e449332f4d1", Quantity:"146", ReservedQuantity:"0", AllocatedQuantity:"0", MinimumStock:"10", MaximumStock:"292", ReorderPoint:"20", EconomicOrderQuantity:"50", SafetyStock:"15", AvgDailyUsage:"5", CreatedDate:"2026-03-13 17:27:50 +0000 UTC", UpdatedDate:"2026-03-13 17:27:50 +0000 UTC"}, inventoryitemapp.InventoryItem{ID:"117942da-b134-48fd-ad7b-7589d7c587b4", ProductID:"85f5a8e9-772f-40e9-b729-92d061edc3a5", LocationID:"ddfff789-3bf3-4941-a49b-5e449332f4d1", Quantity:"121", ReservedQuantity:"0", AllocatedQuantity:"0", MinimumStock:"10", MaximumStock:"242", ReorderPoint:"20", EconomicOrderQuantity:"50", SafetyStock:"15", AvgDailyUsage:"5", CreatedDate:"2026-03-13 17:27:50 +0000 UTC", UpdatedDate:"2026-0
3-13 17:27:50 +0000 UTC"}, inventoryitemapp.InventoryItem{ID:"13beffac-fa04-4bb7-ad77-2d36decf92f0", ProductID:"85f5a8e9-772f-40e9-b729-92d061edc3a5", LocationID:"a862aeb2-e696-459d-985e-c16c027c9156", Quantity:"114", ReservedQuantity:"0", AllocatedQuantity:"0", MinimumStock:"10", MaximumStock:"228", ReorderPoint:"20", EconomicOrderQuantity:"50", SafetyStock:"15", AvgDailyUsage:"5", CreatedDate:"2026-03-13 17:27:50 +0000 UTC", UpdatedDate:"2026-03-13 17:27:50 +0000 UTC"}, inventoryitemapp.InventoryItem{ID:"1a4087e6-c9ab-4bd1-b206-e29e11083b6d", ProductID:"833bd7a6-240a-4d2b-ab43-55ea5ab70f98", LocationID:"821bfaca-104c-4025-9dda-ec90789359b9", Quantity:"137", ReservedQuantity:"0", AllocatedQuantity:"0", MinimumStock:"10", MaximumStock:"274", ReorderPoint:"20", EconomicOrderQuantity:"50", SafetyStock:"15", AvgDailyUsage:"5", CreatedDate:"2026-03-13 17:27:50 +0000 UTC", UpdatedDate:"2026-03-13 17:27:50 +0000 UTC"}, inventoryitemapp.InventoryItem{ID:"226a6328-4725-4afb-9a05-fdeca67a7162", ProductID:"85f5a8e9-772
f-40e9-b729-92d061edc3a5", LocationID:"5b945402-7fde-4e96-8e09-e7c7afab2c02", Quantity:"107", ReservedQuantity:"0", AllocatedQuantity:"0", MinimumStock:"10", MaximumStock:"214", ReorderPoint:"20", EconomicOrderQuantity:"50", SafetyStock:"15", AvgDailyUsage:"5", CreatedDate:"2026-03-13 17:27:50 +0000 UTC", UpdatedDate:"2026-03-13 17:27:50 +0000 UTC"}, inventoryitemapp.InventoryItem{ID:"2343d3e6-7587-44f4-a422-9aa90ca6e752", ProductID:"833bd7a6-240a-4d2b-ab43-55ea5ab70f98", LocationID:"10001e7a-f176-4102-a144-89bffbd66dbb", Quantity:"125", ReservedQuantity:"0", AllocatedQuantity:"0", MinimumStock:"10", MaximumStock:"250", ReorderPoint:"20", EconomicOrderQuantity:"50", SafetyStock:"15", AvgDailyUsage:"5", CreatedDate:"2026-03-13 17:27:50 +0000 UTC", UpdatedDate:"2026-03-13 17:27:50 +0000 UTC"}, inventoryitemapp.InventoryItem{ID:"273c8c21-bc5e-47ce-ad5b-294b0f2ada75", ProductID:"833bd7a6-240a-4d2b-ab43-55ea5ab70f98", LocationID:"5b945402-7fde-4e96-8e09-e7c7afab2c02", Quantity:"132", ReservedQuantity:"0", Allocate
dQuantity:"0", MinimumStock:"10", MaximumStock:"264", ReorderPoint:"20", EconomicOrderQuantity:"50", SafetyStock:"15", AvgDailyUsage:"5", CreatedDate:"2026-03-13 17:27:50 +0000 UTC", UpdatedDate:"2026-03-13 17:27:50 +0000 UTC"}, inventoryitemapp.InventoryItem{ID:"28786329-04ee-46a0-baed-e33f6369d433", ProductID:"833bd7a6-240a-4d2b-ab43-55ea5ab70f98", LocationID:"ca5d921b-a577-40ac-b628-fb8c30fc3fd4", Quantity:"144", ReservedQuantity:"0", AllocatedQuantity:"0", MinimumStock:"10", MaximumStock:"288", ReorderPoint:"20", EconomicOrderQuantity:"50", SafetyStock:"15", AvgDailyUsage:"5", CreatedDate:"2026-03-13 17:27:50 +0000 UTC", UpdatedDate:"2026-03-13 17:27:50 +0000 UTC"}, inventoryitemapp.InventoryItem{ID:"2ce870f6-2bc2-47a4-8c3d-8c32985dd0f5", ProductID:"833bd7a6-240a-4d2b-ab43-55ea5ab70f98", LocationID:"1b8d403c-0d68-46d7-ad1c-999ddb1021b6", Quantity:"126", ReservedQuantity:"0", AllocatedQuantity:"0", MinimumStock:"10", MaximumStock:"252", ReorderPoint:"20", EconomicOrderQuantity:"50", SafetyStock:"15", AvgDa
ilyUsage:"5", CreatedDate:"2026-03-13 17:27:50 +0000 UTC", UpdatedDate:"2026-03-13 17:27:50 +0000 UTC"}, inventoryitemapp.InventoryItem{ID:"33a5aa75-50a1-4485-bbc5-cb61a2520428", ProductID:"833bd7a6-240a-4d2b-ab43-55ea5ab70f98", LocationID:"e46fed35-ac5a-4b49-ad4b-8a6ec23295ba", Quantity:"147", ReservedQuantity:"0", AllocatedQuantity:"0", MinimumStock:"10", MaximumStock:"294", ReorderPoint:"20", EconomicOrderQuantity:"50", SafetyStock:"15", AvgDailyUsage:"5", CreatedDate:"2026-03-13 17:27:50 +0000 UTC", UpdatedDate:"2026-03-13 17:27:50 +0000 UTC"}}, Total:50, Page:1, RowsPerPage:10}
    apitest.go:77: EXP
    apitest.go:78: &query.Result[github.com/timmaaaz/ichor/app/domain/inventory/inventoryitemapp.InventoryItem]{Items:[]inventoryitemapp.InventoryItem{inventoryitemapp.InventoryItem{ID:"2343d3e6-7587-44f4-a422-9aa90ca6e752", ProductID:"833bd7a6-240a-4d2b-ab43-55ea5ab70f98", LocationID:"10001e7a-f176-4102-a144-89bffbd66dbb", Quantity:"125", ReservedQuantity:"0", AllocatedQuantity:"0", MinimumStock:"10", MaximumStock:"250", ReorderPoint:"20", EconomicOrderQuantity:"50", SafetyStock:"15", AvgDailyUsage:"5", CreatedDate:"2026-03-13 17:27:50 +0000 UTC", UpdatedDate:"2026-03-13 17:27:50 +0000 UTC"}, inventoryitemapp.InventoryItem{ID:"2ce870f6-2bc2-47a4-8c3d-8c32985dd0f5", ProductID:"833bd7a6-240a-4d2b-ab43-55ea5ab70f98", LocationID:"1b8d403c-0d68-46d7-ad1c-999ddb1021b6", Quantity:"126", ReservedQuantity:"0", AllocatedQuantity:"0", MinimumStock:"10", MaximumStock:"252", ReorderPoint:"20", EconomicOrderQuantity:"50", SafetyStock:"15", AvgDailyUsage:"5", CreatedDate:"2026-03-13 17:27:50 +0000 UTC", UpdatedDate:"2026-0
3-13 17:27:50 +0000 UTC"}, inventoryitemapp.InventoryItem{ID:"d08b2c76-f924-49d8-95b5-151c35c2c83f", ProductID:"833bd7a6-240a-4d2b-ab43-55ea5ab70f98", LocationID:"267cc967-2596-49f3-8838-1b5146922959", Quantity:"127", ReservedQuantity:"0", AllocatedQuantity:"0", MinimumStock:"10", MaximumStock:"254", ReorderPoint:"20", EconomicOrderQuantity:"50", SafetyStock:"15", AvgDailyUsage:"5", CreatedDate:"2026-03-13 17:27:50 +0000 UTC", UpdatedDate:"2026-03-13 17:27:50 +0000 UTC"}, inventoryitemapp.InventoryItem{ID:"fc572eae-594f-429e-9527-df0a951fc036", ProductID:"833bd7a6-240a-4d2b-ab43-55ea5ab70f98", LocationID:"297ce289-0b50-4568-ae1b-a12ccc09f3b0", Quantity:"128", ReservedQuantity:"0", AllocatedQuantity:"0", MinimumStock:"10", MaximumStock:"256", ReorderPoint:"20", EconomicOrderQuantity:"50", SafetyStock:"15", AvgDailyUsage:"5", CreatedDate:"2026-03-13 17:27:50 +0000 UTC", UpdatedDate:"2026-03-13 17:27:50 +0000 UTC"}, inventoryitemapp.InventoryItem{ID:"60317aa2-44e4-42fa-978a-55821c02a851", ProductID:"833bd7a6-240
a-4d2b-ab43-55ea5ab70f98", LocationID:"38373eaf-7f55-4d11-beea-2cd58e0eba15", Quantity:"129", ReservedQuantity:"0", AllocatedQuantity:"0", MinimumStock:"10", MaximumStock:"258", ReorderPoint:"20", EconomicOrderQuantity:"50", SafetyStock:"15", AvgDailyUsage:"5", CreatedDate:"2026-03-13 17:27:50 +0000 UTC", UpdatedDate:"2026-03-13 17:27:50 +0000 UTC"}, inventoryitemapp.InventoryItem{ID:"7917346a-4ec5-44f0-b789-dc4cc0c693c7", ProductID:"833bd7a6-240a-4d2b-ab43-55ea5ab70f98", LocationID:"480da296-e69d-48be-894a-11cf06b3be3b", Quantity:"130", ReservedQuantity:"0", AllocatedQuantity:"0", MinimumStock:"10", MaximumStock:"260", ReorderPoint:"20", EconomicOrderQuantity:"50", SafetyStock:"15", AvgDailyUsage:"5", CreatedDate:"2026-03-13 17:27:50 +0000 UTC", UpdatedDate:"2026-03-13 17:27:50 +0000 UTC"}, inventoryitemapp.InventoryItem{ID:"e9dbef72-487d-4e6c-bb8c-3d31dc83e2c9", ProductID:"833bd7a6-240a-4d2b-ab43-55ea5ab70f98", LocationID:"50050b00-bdc7-4710-ac27-302df535ad21", Quantity:"131", ReservedQuantity:"0", Allocate
dQuantity:"0", MinimumStock:"10", MaximumStock:"262", ReorderPoint:"20", EconomicOrderQuantity:"50", SafetyStock:"15", AvgDailyUsage:"5", CreatedDate:"2026-03-13 17:27:50 +0000 UTC", UpdatedDate:"2026-03-13 17:27:50 +0000 UTC"}, inventoryitemapp.InventoryItem{ID:"273c8c21-bc5e-47ce-ad5b-294b0f2ada75", ProductID:"833bd7a6-240a-4d2b-ab43-55ea5ab70f98", LocationID:"5b945402-7fde-4e96-8e09-e7c7afab2c02", Quantity:"132", ReservedQuantity:"0", AllocatedQuantity:"0", MinimumStock:"10", MaximumStock:"264", ReorderPoint:"20", EconomicOrderQuantity:"50", SafetyStock:"15", AvgDailyUsage:"5", CreatedDate:"2026-03-13 17:27:50 +0000 UTC", UpdatedDate:"2026-03-13 17:27:50 +0000 UTC"}, inventoryitemapp.InventoryItem{ID:"8fb95b12-1e53-4117-99f6-e3be9d8cc3aa", ProductID:"833bd7a6-240a-4d2b-ab43-55ea5ab70f98", LocationID:"614c0afb-1bd8-4350-aa4e-2b085b609764", Quantity:"133", ReservedQuantity:"0", AllocatedQuantity:"0", MinimumStock:"10", MaximumStock:"266", ReorderPoint:"20", EconomicOrderQuantity:"50", SafetyStock:"15", AvgDa
ilyUsage:"5", CreatedDate:"2026-03-13 17:27:50 +0000 UTC", UpdatedDate:"2026-03-13 17:27:50 +0000 UTC"}, inventoryitemapp.InventoryItem{ID:"46d8976e-ad19-4bce-8d44-ee5330c31a79", ProductID:"833bd7a6-240a-4d2b-ab43-55ea5ab70f98", LocationID:"678e86c5-6448-4782-bbd4-5c5cb04d1dc2", Quantity:"134", ReservedQuantity:"0", AllocatedQuantity:"0", MinimumStock:"10", MaximumStock:"268", ReorderPoint:"20", EconomicOrderQuantity:"50", SafetyStock:"15", AvgDailyUsage:"5", CreatedDate:"2026-03-13 17:27:50 +0000 UTC", UpdatedDate:"2026-03-13 17:27:50 +0000 UTC"}}, Total:50, Page:1, RowsPerPage:10}
    apitest.go:79: Should get the expected response
--- FAIL: Test_InventoryItem/query-200-basic (0.01s)
```

## Fix
- **File**: `business/domain/inventory/inventoryitembus/testutil.go:52`
- **Classification**: test bug
- **Change**: Changed sort order in `TestSeedInventoryItems` from `(product_id, location_id)` to `id ASC` to match `DefaultOrderBy = order.NewBy(OrderByID, order.ASC)`
- **Verified**: `go test -v -run Test_InventoryItem/query-200-basic github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/inventory/inventoryitemapi` ✓
