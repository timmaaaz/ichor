package presenceapi

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/foundation/web"
)

// ActiveWorkersHub defines the subset of AlertHub needed by this handler.
type ActiveWorkersHub interface {
	ConnectedUserIDs() []uuid.UUID
}

type api struct {
	alertHub ActiveWorkersHub
}

func newAPI(alertHub ActiveWorkersHub) *api {
	return &api{alertHub: alertHub}
}

// activeWorkersResponse is the JSON response shape for the presence endpoint.
type activeWorkersResponse struct {
	Count   int        `json:"count"`
	UserIDs []string   `json:"userIds"`
}

func (r activeWorkersResponse) Encode() ([]byte, string, error) {
	data, err := json.Marshal(r)
	return data, "application/json", err
}

func (api *api) activeWorkers(_ context.Context, _ *http.Request) web.Encoder {
	userIDs := api.alertHub.ConnectedUserIDs()

	ids := make([]string, len(userIDs))
	for i, uid := range userIDs {
		ids[i] = uid.String()
	}

	return activeWorkersResponse{
		Count:   len(userIDs),
		UserIDs: ids,
	}
}
