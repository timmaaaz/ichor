package mid

import (
	"context"
	"net/http"

	"github.com/timmaaaz/ichor/app/sdk/mid"
	"github.com/timmaaaz/ichor/foundation/web"
)

// RestrictFields restricts the fields in the response.
func RestrictFields() web.MidFunc {
	midFunc := func(ctx context.Context, r *http.Request, next mid.HandlerFunc) mid.Encoder {
		return mid.RestrictFields(ctx, next)
	}

	return addMidFunc(midFunc)
}
