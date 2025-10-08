package data_test

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/dataapp"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/tablebuilder"
)

func update200(sd apitest.SeedData) []apitest.Table {

	updatedSimple := SimpleConfig
	updatedSimple.DataSource[0].Select.Columns = append(updatedSimple.DataSource[0].Select.Columns,
		// New columns
		tablebuilder.ColumnDefinition{Name: "minimum_stock", TableColumn: "inventory_items.minimum_stock"},
		tablebuilder.ColumnDefinition{Name: "maximum_stock", TableColumn: "inventory_items.maximum_stock"},
	)

	jsonConfig, err := json.Marshal(updatedSimple)
	if err != nil {
		panic(err)
	}
	var raw *json.RawMessage
	tmp := json.RawMessage(jsonConfig)
	raw = &tmp

	return []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/data/%s", sd.SimpleTableConfig.ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: dataapp.UpdateTableConfig{
				Config: raw,
			},
			GotResp: &dataapp.TableConfig{},
			ExpResp: &dataapp.TableConfig{
				Name:        sd.SimpleTableConfig.Name,
				Description: sd.SimpleTableConfig.Description,
				Config:      json.RawMessage(jsonConfig),
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*dataapp.TableConfig)
				if !exists {
					return "could not convert got to *dataapp.TableConfig"
				}
				expResp, exists := exp.(*dataapp.TableConfig)
				if !exists {
					return "could not convert exp to *dataapp.TableConfig"
				}

				// Copy over the generated fields from got to exp
				expResp.ID = gotResp.ID
				expResp.CreatedBy = gotResp.CreatedBy
				expResp.UpdatedBy = gotResp.UpdatedBy
				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate

				dbtest.NormalizeJSONFields(gotResp, &expResp)

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}
