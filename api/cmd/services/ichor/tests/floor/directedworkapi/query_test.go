package directedworkapi_test

import (
	"net/http"
	"testing"

	"github.com/timmaaaz/ichor/api/domain/http/floor/directedworkapi"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

func Test_DirectedWork_Query(t *testing.T) {
	at := apitest.StartTest(t, "Test_DirectedWork_Query")

	sd, err := insertSeedData(at.DB, at.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	at.Run(t, queryTests(sd), "queryNext")
}

func queryTests(sd DirectedWorkSeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "unassigned-worker-returns-null",
			URL:        "/v1/floor/work/next",
			Token:      sd.Unassigned.Token,
			Method:     http.MethodGet,
			StatusCode: http.StatusOK,
			GotResp:    &directedworkapi.Response{},
			ExpResp: &directedworkapi.Response{
				WorkItem: nil,
			},
			CmpFunc: func(got, exp any) string {
				g := got.(*directedworkapi.Response)
				if g.WorkItem != nil {
					return "expected work_item=null for unassigned worker, got non-nil"
				}
				return ""
			},
		},
		{
			Name:       "worker-with-one-pending-pick-returns-it",
			URL:        "/v1/floor/work/next",
			Token:      sd.Worker.Token,
			Method:     http.MethodGet,
			StatusCode: http.StatusOK,
			GotResp:    &directedworkapi.Response{},
			ExpResp: &directedworkapi.Response{
				WorkItem: &directedworkapi.WorkItem{
					Type:   directedworkapi.WorkItemTypePick,
					Status: directedworkapi.WorkItemStatusPending,
				},
			},
			CmpFunc: func(got, exp any) string {
				g := got.(*directedworkapi.Response)
				if g.WorkItem == nil {
					return "expected a work item, got nil"
				}
				if g.WorkItem.Type != directedworkapi.WorkItemTypePick {
					return "expected type=pick, got " + string(g.WorkItem.Type)
				}
				if g.WorkItem.Status != directedworkapi.WorkItemStatusPending {
					return "expected status=pending, got " + string(g.WorkItem.Status)
				}
				return ""
			},
		},
	}
}
