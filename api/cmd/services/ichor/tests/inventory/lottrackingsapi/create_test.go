package lottrackingsapi_test

import (
	"net/http"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/business/domain/inventory/lottrackingsbus"
	"github.com/timmaaaz/ichor/foundation/timeutil"

	"github.com/timmaaaz/ichor/app/domain/inventory/lottrackingsapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {

	md := lottrackingsbus.RandomDate()
	ed := lottrackingsbus.RandomDate()
	rd := lottrackingsbus.RandomDate()

	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/inventory/lot-trackings",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &lottrackingsapp.NewLotTrackings{
				SupplierProductID: sd.SupplierProducts[0].SupplierProductID,
				LotNumber:         "LotNumber",
				ManufactureDate:   md.Format(timeutil.FORMAT),
				ExpirationDate:    ed.Format(timeutil.FORMAT),
				RecievedDate:      rd.Format(timeutil.FORMAT),
				Quantity:          "15",
				QualityStatus:     "poor",
			},
			GotResp: &lottrackingsapp.LotTrackings{},
			ExpResp: &lottrackingsapp.LotTrackings{
				SupplierProductID: sd.SupplierProducts[0].SupplierProductID,
				LotNumber:         "LotNumber",
				ManufactureDate:   md.Format(timeutil.FORMAT),
				ExpirationDate:    ed.Format(timeutil.FORMAT),
				RecievedDate:      rd.Format(timeutil.FORMAT),
				Quantity:          "15",
				QualityStatus:     "poor",
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*lottrackingsapp.LotTrackings)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*lottrackingsapp.LotTrackings)
				expResp.LotID = gotResp.LotID
				expResp.UpdatedDate = gotResp.UpdatedDate
				expResp.CreatedDate = gotResp.CreatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func create400(sd apitest.SeedData) []apitest.Table {

	md := lottrackingsbus.RandomDate()
	ed := lottrackingsbus.RandomDate()
	rd := lottrackingsbus.RandomDate()

	return []apitest.Table{
		{
			Name:       "missing-supplier-product-id",
			URL:        "/v1/inventory/lot-trackings",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &lottrackingsapp.NewLotTrackings{
				LotNumber:       "LotNumber",
				ManufactureDate: md.Format(timeutil.FORMAT),
				ExpirationDate:  ed.Format(timeutil.FORMAT),
				RecievedDate:    rd.Format(timeutil.FORMAT),
				Quantity:        "15",
				QualityStatus:   "poor",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"supplier_product_id\",\"error\":\"supplier_product_id is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-lot-number",
			URL:        "/v1/inventory/lot-trackings",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &lottrackingsapp.NewLotTrackings{
				SupplierProductID: sd.SupplierProducts[0].SupplierProductID,
				ManufactureDate:   md.Format(timeutil.FORMAT),
				ExpirationDate:    ed.Format(timeutil.FORMAT),
				RecievedDate:      rd.Format(timeutil.FORMAT),
				Quantity:          "15",
				QualityStatus:     "poor",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"lot_number\",\"error\":\"lot_number is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-manufacture-date",
			URL:        "/v1/inventory/lot-trackings",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &lottrackingsapp.NewLotTrackings{
				SupplierProductID: sd.SupplierProducts[0].SupplierProductID,
				LotNumber:         "LotNumber",
				ExpirationDate:    ed.Format(timeutil.FORMAT),
				RecievedDate:      rd.Format(timeutil.FORMAT),
				Quantity:          "15",
				QualityStatus:     "poor",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"manufacture_date\",\"error\":\"manufacture_date is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-expiration-date",
			URL:        "/v1/inventory/lot-trackings",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &lottrackingsapp.NewLotTrackings{
				SupplierProductID: sd.SupplierProducts[0].SupplierProductID,
				LotNumber:         "LotNumber",
				ManufactureDate:   md.Format(timeutil.FORMAT),
				RecievedDate:      rd.Format(timeutil.FORMAT),
				Quantity:          "15",
				QualityStatus:     "poor",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"expiration_date\",\"error\":\"expiration_date is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-recieved-date",
			URL:        "/v1/inventory/lot-trackings",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &lottrackingsapp.NewLotTrackings{
				SupplierProductID: sd.SupplierProducts[0].SupplierProductID,
				LotNumber:         "LotNumber",
				ManufactureDate:   md.Format(timeutil.FORMAT),
				ExpirationDate:    ed.Format(timeutil.FORMAT),
				Quantity:          "15",
				QualityStatus:     "poor",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"received_date\",\"error\":\"received_date is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-quantity",
			URL:        "/v1/inventory/lot-trackings",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &lottrackingsapp.NewLotTrackings{
				SupplierProductID: sd.SupplierProducts[0].SupplierProductID,
				LotNumber:         "LotNumber",
				ManufactureDate:   md.Format(timeutil.FORMAT),
				ExpirationDate:    ed.Format(timeutil.FORMAT),
				RecievedDate:      rd.Format(timeutil.FORMAT),
				QualityStatus:     "poor",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"quantity\",\"error\":\"quantity is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-quality-status",
			URL:        "/v1/inventory/lot-trackings",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &lottrackingsapp.NewLotTrackings{
				SupplierProductID: sd.SupplierProducts[0].SupplierProductID,
				LotNumber:         "LotNumber",
				ManufactureDate:   md.Format(timeutil.FORMAT),
				ExpirationDate:    ed.Format(timeutil.FORMAT),
				RecievedDate:      rd.Format(timeutil.FORMAT),
				Quantity:          "15",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"quality_status\",\"error\":\"quality_status is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-manufacture-date",
			URL:        "/v1/inventory/lot-trackings",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &lottrackingsapp.NewLotTrackings{
				SupplierProductID: sd.SupplierProducts[0].SupplierProductID,
				LotNumber:         "LotNumber",
				ManufactureDate:   md.Format(time.RFC1123),
				ExpirationDate:    ed.Format(timeutil.FORMAT),
				RecievedDate:      rd.Format(timeutil.FORMAT),
				Quantity:          "15",
				QualityStatus:     "poor",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `toBusNewLotTrackings: failed to parse time: parsing time "`+md.Format(time.RFC1123)+`" as "`+timeutil.FORMAT+`": cannot parse "`+md.Format(time.RFC1123)+`" as "2006"`),
			CmpFunc: func(got, exp any) string {

				gotResp, exists := got.(*errs.Error)
				if !exists {
					return "got is not of type errs.Error"
				}

				expResp, exists := exp.(*errs.Error)
				if !exists {
					return "exp is not of type errs.Error"
				}

				return cmp.Diff(gotResp.Error(), expResp.Error(), cmpopts.EquateErrors())
			},
		},
		{
			Name:       "malformed-expiration-date",
			URL:        "/v1/inventory/lot-trackings",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &lottrackingsapp.NewLotTrackings{
				SupplierProductID: sd.SupplierProducts[0].SupplierProductID,
				LotNumber:         "LotNumber",
				ManufactureDate:   md.Format(timeutil.FORMAT),
				ExpirationDate:    ed.Format(time.RFC1123),
				RecievedDate:      rd.Format(timeutil.FORMAT),
				Quantity:          "15",
				QualityStatus:     "poor",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `toBusNewLotTrackings: failed to parse time: parsing time "`+ed.Format(time.RFC1123)+`" as "`+timeutil.FORMAT+`": cannot parse "`+ed.Format(time.RFC1123)+`" as "2006"`),
			CmpFunc: func(got, exp any) string {

				gotResp, exists := got.(*errs.Error)
				if !exists {
					return "got is not of type errs.Error"
				}

				expResp, exists := exp.(*errs.Error)
				if !exists {
					return "exp is not of type errs.Error"
				}

				return cmp.Diff(gotResp.Error(), expResp.Error(), cmpopts.EquateErrors())
			},
		},
		{
			Name:       "malformed-recieved-date",
			URL:        "/v1/inventory/lot-trackings",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &lottrackingsapp.NewLotTrackings{
				SupplierProductID: sd.SupplierProducts[0].SupplierProductID,
				LotNumber:         "LotNumber",
				ManufactureDate:   md.Format(timeutil.FORMAT),
				ExpirationDate:    ed.Format(timeutil.FORMAT),
				RecievedDate:      rd.Format(time.RFC1123),
				Quantity:          "15",
				QualityStatus:     "poor",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `toBusNewLotTrackings: failed to parse time: parsing time "`+rd.Format(time.RFC1123)+`" as "`+timeutil.FORMAT+`": cannot parse "`+rd.Format(time.RFC1123)+`" as "2006"`),
			CmpFunc: func(got, exp any) string {

				gotResp, exists := got.(*errs.Error)
				if !exists {
					return "got is not of type errs.Error"
				}

				expResp, exists := exp.(*errs.Error)
				if !exists {
					return "exp is not of type errs.Error"
				}

				return cmp.Diff(gotResp.Error(), expResp.Error(), cmpopts.EquateErrors())
			},
		},

		{
			Name:       "malformed-supplier-product-id",
			URL:        "/v1/inventory/lot-trackings",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &lottrackingsapp.NewLotTrackings{
				SupplierProductID: "not-a-uuid",
				LotNumber:         "LotNumber",
				ManufactureDate:   md.Format(timeutil.FORMAT),
				ExpirationDate:    ed.Format(timeutil.FORMAT),
				RecievedDate:      rd.Format(timeutil.FORMAT),
				Quantity:          "15",
				QualityStatus:     "poor",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"supplier_product_id\",\"error\":\"supplier_product_id must be at least 36 characters in length\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func create409(sd apitest.SeedData) []apitest.Table {
	md := lottrackingsbus.RandomDate()
	ed := lottrackingsbus.RandomDate()
	rd := lottrackingsbus.RandomDate()
	return []apitest.Table{
		{
			Name:       "supplier-product-id-not-valid-fk",
			URL:        "/v1/inventory/lot-trackings",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusConflict,
			Input: &lottrackingsapp.NewLotTrackings{
				SupplierProductID: uuid.New().String(),
				LotNumber:         "LotNumber",
				ManufactureDate:   md.Format(timeutil.FORMAT),
				ExpirationDate:    ed.Format(timeutil.FORMAT),
				RecievedDate:      rd.Format(timeutil.FORMAT),
				Quantity:          "15",
				QualityStatus:     "poor",
			},
			ExpResp: errs.Newf(errs.Aborted, "create: namedexeccontext: foreign key violation"),
			GotResp: &errs.Error{},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func create401(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "empty token",
			URL:        "/v1/inventory/lot-trackings",
			Token:      "&nbsp;",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad token",
			URL:        "/v1/inventory/lot-trackings",
			Token:      sd.Admins[0].Token[:10],
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad sig",
			URL:        "/v1/inventory/lot-trackings",
			Token:      sd.Admins[0].Token + "A",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "authentication failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "roleadminonly",
			URL:        "/v1/inventory/lot-trackings",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission CREATE for table: inventory.lot_trackings"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}
